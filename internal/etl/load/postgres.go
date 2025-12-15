package load

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	t "github.com/PDOK/gokoala/internal/etl/transform"
	"github.com/jackc/pgx/v5"
	pgxgeom "github.com/twpayne/pgx-geom"
	"golang.org/x/text/language"
)

const (
	indexNameFullText = "ts_idx"
	indexNameGeometry = "geometry_idx"
	indexNameBbox     = "bbox_idx"
	indexNamePreRank  = "pre_rank_idx"

	alphaPartition = "_alpha"
	betaPartition  = "_beta"

	postgresDetachErr = "already pending detach in partitioned table"
)

var (
	postgresExtensions = []string{"postgis", "unaccent"}

	indexNames = []string{indexNameFullText, indexNameGeometry, indexNameBbox, indexNamePreRank}

	//nolint:dupword
	tableDefinition = `
	create table if not exists %[1]s (
		id 					serial,
		feature_id 			text 					 not null,
		external_fid        text                     null,
		collection_id 		text					 not null,
		collection_version 	int 					 not null,
		display_name 		text					 not null,
		suggest 			text					 not null,
		geometry_type 		geometry_type			 not null,
		bbox 				geometry(polygon, %[2]d) null,
		geometry            geometry(point, %[2]d)   null,
	    ts                  tsvector                 generated always as (to_tsvector('custom_dict', suggest)) stored
	) %[3]s;`

	metadataTableDefinition = `
	create table if not exists %[1]s_metadata (
		collection_id      text                      not null,
		collection_version uuid                      not null,
		updated_date       timestamptz default now() not null
	);`
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

func (p *Postgres) PostLoad(collectionID string, index string, collectionVersion string) error {
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
        insert into %[1]s_metadata (collection_id, collection_version)
        values ('%[2]s', '%[3]s')
        on conflict (collection_id)
        do update set collection_version = '%[3]s';`, index, collectionID, collectionVersion)
	_, err = p.db.Exec(context.Background(), metadata)
	if err != nil {
		return fmt.Errorf("error updating metadata table of index %s. Error: %w", index, err)
	}
	return nil
}

func (p *Postgres) Optimize() error {
	_, err := p.db.Exec(context.Background(), `vacuum analyze;`)
	if err != nil {
		return fmt.Errorf("failed optimizing: error performing vacuum analyze: %w", err)
	}
	return nil
}

// Get the current version of a collection in the search index
func (p *Postgres) GetVersion(collectionID string, index string) (string, error) {
	currentVersion := ""
	err := p.db.QueryRow(context.Background(), fmt.Sprintf(`
        select collection_version
        from %[2]s_metadata
        where collection_id = '%[1]s';`, collectionID, index)).Scan(&currentVersion)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("no version found for collection '%s' in index '%s'", collectionID, index)
			return currentVersion, nil
		}
		if strings.Contains(err.Error(), fmt.Sprintf("relation \"%[1]s_metadata\" does not exist", index)) {
			log.Printf("metadata table for index '%s' does not exist", index)
			return currentVersion, nil
		}
		return "", fmt.Errorf("error getting version: %w", err)
	}
	return currentVersion, nil
}

// Init initialize search index. Should be idempotent!
//
// Since not all DDL in Postgres support the "if not exists" syntax we use a bit
// of pl/pgsql to make it idempotent.
func (p *Postgres) Init(index string, srid int, lang language.Tag) error {
	geometryType := `
		do $$ begin
		    create type geometry_type as enum ('POINT', 'MULTIPOINT', 'LINESTRING', 'MULTILINESTRING', 'POLYGON', 'MULTIPOLYGON');
		exception
		    when duplicate_object then null;
		end $$;`
	_, err := p.db.Exec(context.Background(), geometryType)
	if err != nil {
		return fmt.Errorf("error creating geometry type: %w", err)
	}

	textSearchConfig := `
		do $$ begin
		    create text search configuration custom_dict (copy = simple);
		exception
		    when unique_violation then null;
		end $$;`
	_, err = p.db.Exec(context.Background(), textSearchConfig)
	if err != nil {
		return fmt.Errorf("error creating text search configuration: %w", err)
	}

	// This adds the 'unaccent' extension to allow searching with/without diacritics. Must happen in separate transaction.
	alterTextSearchConfig := `
		do $$ begin
			alter text search configuration custom_dict
			  alter mapping for hword, hword_part, word
			  with unaccent, simple;
		exception
		    when unique_violation then null;
		end $$;`
	_, err = p.db.Exec(context.Background(), alterTextSearchConfig)
	if err != nil {
		return fmt.Errorf("error altering text search configuration: %w", err)
	}

	// create search index table
	_, err = p.db.Exec(context.Background(), fmt.Sprintf(tableDefinition, index, srid, "partition by list(collection_id)"))
	if err != nil {
		return fmt.Errorf("error creating search index table: %w", err)
	}

	// create primary key when it doesn't exist yet
	primaryKey := fmt.Sprintf(`
		do $$
		begin
		    if not exists (
		        select 1
		        from   pg_constraint
		        where  conrelid = '%[1]s'::regclass
		        and    contype = 'p'
		    )
		    then
		        alter table %[1]s
		        add constraint %[1]s_pkey primary key (id, collection_id, collection_version);
		    end if;
		end;
		$$;`, index)
	_, err = p.db.Exec(context.Background(), primaryKey)
	if err != nil {
		return fmt.Errorf("error creating primary key: %w", err)
	}

	// create search index metadata table
	_, err = p.db.Exec(context.Background(), fmt.Sprintf(metadataTableDefinition, index))
	if err != nil {
		return fmt.Errorf("error creating search index metadata table: %w", err)
	}

	// create metadata primary key when it doesn't exist yet
	metadataPrimaryKey := fmt.Sprintf(`
		do $$
		begin
		    if not exists (
		        select 1
		        from   pg_constraint
		        where  conrelid = '%[1]s_metadata'::regclass
		        and    contype = 'p'
		    )
		    then
		        alter table %[1]s_metadata
		        add constraint %[1]s_metadata_pkey primary key (collection_id);
		    end if;
		end;
		$$;`, index)
	_, err = p.db.Exec(context.Background(), metadataPrimaryKey)
	if err != nil {
		return fmt.Errorf("error creating metadata primary key: %w", err)
	}

	// create custom collation to correctly handle "numbers in strings" when sorting results
	// see https://www.postgresql.org/docs/12/collation.html#id-1.6.10.4.5.7.5
	collation := fmt.Sprintf(`create collation if not exists custom_numeric (provider = icu, locale = '%s-u-kn-true');`, lang.String())
	_, err = p.db.Exec(context.Background(), collation)
	if err != nil {
		return fmt.Errorf("error creating numeric collation: %w", err)
	}

	if err = p.createIndexes(index, false); err != nil {
		return err
	}
	return err
}

func createExtensions(ctx context.Context, db *pgx.Conn) error {
	for _, ext := range postgresExtensions {
		_, err := db.Exec(ctx, `create extension if not exists `+ext+`;`)
		if err != nil {
			return fmt.Errorf("error creating %s extension: %w", ext, err)
		}
	}
	return nil
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

func (p *Postgres) createIndexes(table string, usePrefix bool) error {
	// GIN indexes are best for text search
	indexName := indexNameFullText
	if usePrefix {
		indexName = fmt.Sprintf("%s_%s", table, indexNameFullText)
	}
	_, err := p.db.Exec(context.Background(), fmt.Sprintf(`create index if not exists %[2]s on only %[1]s using gin(ts);`, table, indexName))
	if err != nil {
		return fmt.Errorf("error creating GIN index: %w", err)
	}

	// GIST indexes for geometry column to support search within a bounding box
	indexName = indexNameGeometry
	if usePrefix {
		indexName = fmt.Sprintf("%s_%s", table, indexNameGeometry)
	}
	_, err = p.db.Exec(context.Background(), fmt.Sprintf(`create index if not exists %[2]s on only %[1]s using gist(geometry);`, table, indexName))
	if err != nil {
		return fmt.Errorf("error creating geometry GIST index: %w", err)
	}

	// GIST indexes for bbox column to support search within a bounding box
	indexName = indexNameBbox
	if usePrefix {
		indexName = fmt.Sprintf("%s_%s", table, indexNameBbox)
	}
	_, err = p.db.Exec(context.Background(), fmt.Sprintf(`create index if not exists %[2]s on only %[1]s using gist(bbox);`, table, indexName))
	if err != nil {
		return fmt.Errorf("error creating bbox GIST index: %w", err)
	}

	// index used to pre-rank results when generic search terms are used
	indexName = indexNamePreRank
	if usePrefix {
		indexName = fmt.Sprintf("%s_%s", table, indexNamePreRank)
	}
	preRankIndex := fmt.Sprintf(`create index if not exists %[2]s on only %[1]s
		(array_length(string_to_array(suggest, ' '), 1) asc, display_name collate "custom_numeric" asc);`, table, indexName)
	_, err = p.db.Exec(context.Background(), preRankIndex)
	if err != nil {
		return fmt.Errorf("error creating pre-rank index: %w", err)
	}
	return nil
}

// CHECK constraint is to avoid ACCESS EXCLUSIVE lock on partition as mentioned on
// https://www.postgresql.org/docs/current/ddl-partitioning.html#DDL-PARTITIONING-DECLARATIVE-MAINTENANCE
func (p *Postgres) createCheck(collectionID string) error {
	dropCheck := fmt.Sprintf(`alter table %[1]s drop constraint if exists %[1]s_col_chk;`, p.partitionToLoad)
	_, err := p.db.Exec(context.Background(), dropCheck)
	if err != nil {
		return fmt.Errorf("error dropping CHECK constraint: %w", err)
	}

	addCheck := fmt.Sprintf(`alter table if exists %[1]s add constraint %[1]s_col_chk check (collection_id = '%[2]s');`,
		p.partitionToLoad, collectionID)
	_, err = p.db.Exec(context.Background(), addCheck)
	if err != nil {
		return fmt.Errorf("error creating CHECK constraint: %w", err)
	}
	return nil
}
