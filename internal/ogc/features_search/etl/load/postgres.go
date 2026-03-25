package load

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	t "github.com/PDOK/gokoala/internal/ogc/features_search/etl/transform"
	"github.com/jackc/pgx/v5"
	pgxgeom "github.com/twpayne/pgx-geom"
)

const (
	alphaPartition = "_alpha"
	betaPartition  = "_beta"

	postgresDetachErr = "already pending detach in partitioned table"
)

type Postgres struct {
	db *pgx.Conn

	partitionToLoad   string
	partitionToDetach string
}

func NewPostgres(dbConn string) (*Postgres, error) {
	ctx := context.Background()
	db, err := pgx.Connect(ctx, dbConn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	err = createExtensions(ctx, db)
	if err != nil {
		return nil, err
	}
	// add support for Go <-> PostGIS conversions
	if err := pgxgeom.Register(ctx, db); err != nil {
		return nil, err
	}
	return &Postgres{db: db}, nil
}

func (p *Postgres) Close() {
	_ = p.db.Close(context.Background())
}

func (p *Postgres) PreLoad(collectionID string, index string) error {
	tablePrefix := index + "_" + collectionID
	tables := []string{tablePrefix + alphaPartition, tablePrefix + betaPartition}

	for _, table := range tables {
		tableIsPartition, err := p.isPartition(table, index)
		if err != nil {
			return fmt.Errorf("error querying partition status for collection: %s Error: %w", collectionID, err)
		}
		if tableIsPartition {
			p.partitionToDetach = table
		} else if p.partitionToLoad == "" {
			p.partitionToLoad = table
		}
	}

	if p.partitionToLoad != "" {
		var srid int
		if err := p.db.QueryRow(context.Background(), `select find_srid('public', $1, 'geometry')`, index).Scan(&srid); err != nil {
			return fmt.Errorf("error finding SRID of search index: %w", err)
		}
		_, err := p.db.Exec(context.Background(), fmt.Sprintf(tableDefinition, p.partitionToLoad, srid, ""))
		if err != nil {
			return fmt.Errorf("error creating table (which will later be attached as a partition): %w", err)
		}
		_, err = p.db.Exec(context.Background(), fmt.Sprintf(`truncate table %[1]s;`, p.partitionToLoad))
		if err != nil {
			return fmt.Errorf("error truncating table: %w", err)
		}
		if err = p.createCheck(collectionID); err != nil {
			return err
		}
		if err = p.createIndexes(p.partitionToLoad, true); err != nil {
			return err
		}
	}
	return nil
}

func (p *Postgres) Load(records []t.SearchIndexRecord) (int64, error) {
	loaded, err := p.db.CopyFrom(
		context.Background(),
		pgx.Identifier{p.partitionToLoad},
		[]string{"feature_id", "external_fid", "collection_id", "collection_version", "display_name", "suggest", "geometry_type", "bbox", "geometry"},
		pgx.CopyFromSlice(len(records), func(i int) ([]any, error) {
			r := records[i]
			return []any{r.FeatureID, r.ExternalFid, r.CollectionID, r.CollectionVersion, r.DisplayName, r.Suggest, r.GeometryType, r.Bbox, r.Geometry}, nil
		}),
	)
	if err != nil {
		return -1, fmt.Errorf("unable to copy records: %w", err)
	}
	return loaded, nil
}

func (p *Postgres) PostLoad(collectionID string, index string, revision string) error {
	if p.partitionToDetach != "" {
	RETRY:
		detach := fmt.Sprintf(`alter table %[1]s detach partition %[2]s concurrently;`, index, p.partitionToDetach)
		_, err := p.db.Exec(context.Background(), detach)
		if err != nil {
			if strings.Contains(err.Error(), postgresDetachErr) {
				log.Printf("(another) partition is already being detached from index %s. "+
					"Retrying detach of partition %s in 1 minute.", index, p.partitionToLoad)
				time.Sleep(1 * time.Minute)
				goto RETRY
			}
			return fmt.Errorf("error detaching partition %s from index %s. Error: %w", p.partitionToLoad, index, err)
		}
	}

	attach := fmt.Sprintf(`alter table %[1]s attach partition %[2]s for values in ('%[3]s');`, index, p.partitionToLoad, collectionID)
	_, err := p.db.Exec(context.Background(), attach)
	if err != nil {
		return fmt.Errorf("error attaching table %s as partition of index %s. Error: %w", p.partitionToLoad, index, err)
	}

	for _, indexName := range indexNames {
		attachIndex := fmt.Sprintf(`alter index %[2]s attach partition %[1]s_%[2]s;`, p.partitionToLoad, indexName)
		_, err = p.db.Exec(context.Background(), attachIndex)
		if err != nil {
			return fmt.Errorf("error attaching partition index to parent index: %s: %w", indexName, err)
		}
	}

	metadata := fmt.Sprintf(`
        insert into %[1]s_metadata (collection_id, revision)
        values ('%[2]s', '%[3]s')
        on conflict (collection_id)
        do update set revision = '%[3]s', revision_date = now();`, index, collectionID, revision)
	_, err = p.db.Exec(context.Background(), metadata)
	if err != nil {
		return fmt.Errorf("error updating metadata table of index %s. Error: %w", index, err)
	}
	return nil
}

func (p *Postgres) Optimize(index string) error {
	log.Println("perform targeted VACUUM + ANALYZE on loaded partition")
	_, err := p.db.Exec(context.Background(), fmt.Sprintf(`vacuum analyze %s;`, p.partitionToLoad))
	if err != nil {
		return fmt.Errorf("failed optimizing: error performing vacuum analyze on loaded partition: %w", err)
	}
	_, err = p.db.Exec(context.Background(), fmt.Sprintf(`analyze %s;`, index))
	if err != nil {
		return fmt.Errorf("failed optimizing: error performing analyze on search index: %w", err)
	}

	// Execute pg_prewarm on all partitions and important indexes (forcing these into Postgres shared_buffers memory)
	log.Println("prewarming partitions")
	preWarmCall := fmt.Sprintf(
		`select gokoala_prewarm_partitions(idx_suffixes := array['%s'])`, strings.Join(indexesToPreWarm, "','"))
	_, err = p.db.Exec(context.Background(), preWarmCall)
	if err != nil {
		return fmt.Errorf("failed optimizing: prewarm function failed: %w", err)
	}
	return nil
}

// GetRevision get the revision of a collection in the search index
func (p *Postgres) GetRevision(collectionID string, index string) (string, error) {
	currentRevision := ""
	err := p.db.QueryRow(context.Background(), fmt.Sprintf(`
        select revision
        from %[2]s_metadata
        where collection_id = '%[1]s';`, collectionID, index)).Scan(&currentRevision)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("no revision found for collection '%s' in index '%s'", collectionID, index)
			return currentRevision, nil
		}
		if strings.Contains(err.Error(), fmt.Sprintf("relation \"%[1]s_metadata\" does not exist", index)) {
			log.Printf("metadata table for index '%s' does not exist", index)
			return currentRevision, nil
		}
		return "", fmt.Errorf("error getting revision: %w", err)
	}
	return currentRevision, nil
}

func (p *Postgres) isPartition(collectionID string, index string) (bool, error) {
	result := false
	err := p.db.QueryRow(context.Background(), `select exists (
	        select 1
	        from pg_catalog.pg_inherits as i
	        join pg_catalog.pg_class as parent on i.inhparent = parent.oid
	        join pg_catalog.pg_class as child on i.inhrelid = child.oid
	        where parent.relname = $1
	          and child.relname = $2
	          and parent.relkind = 'p' -- p means partitioned table
	    ) as is_partition_of_search_index;`, index, collectionID).Scan(&result)
	return result, err
}
