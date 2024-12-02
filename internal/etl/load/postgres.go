package load

import (
	"context"
	"fmt"

	t "github.com/PDOK/gomagpie/internal/etl/transform"
	"github.com/jackc/pgx/v5"
	pgxgeom "github.com/twpayne/pgx-geom"
)

type Postgres struct {
	db  *pgx.Conn
	ctx context.Context
}

func NewPostgres(dbConn string) (*Postgres, error) {
	ctx := context.Background()
	db, err := pgx.Connect(ctx, dbConn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	// add support for Go <-> PostGIS conversions
	if err := pgxgeom.Register(ctx, db); err != nil {
		return nil, err
	}
	return &Postgres{db: db, ctx: ctx}, nil
}

func (p *Postgres) Close() {
	_ = p.db.Close(p.ctx)
}

func (p *Postgres) Load(records []t.SearchIndexRecord, index string) (int64, error) {
	loaded, err := p.db.CopyFrom(
		p.ctx,
		pgx.Identifier{index},
		[]string{"feature_id", "collection_id", "collection_version", "display_name", "suggest", "geometry_type", "bbox"},
		pgx.CopyFromSlice(len(records), func(i int) ([]interface{}, error) {
			r := records[i]
			return []any{r.FeatureID, r.CollectionID, r.CollectionVersion, r.DisplayName, r.Suggest, r.GeometryType, r.Bbox}, nil
		}),
	)
	if err != nil {
		return -1, fmt.Errorf("unable to copy records: %w", err)
	}
	return loaded, nil
}

// Init initialize search index
func (p *Postgres) Init(index string) error {
	geometryType := `create type geometry_type as enum ('POINT', 'MULTIPOINT', 'LINESTRING', 'MULTILINESTRING', 'POLYGON', 'MULTIPOLYGON');`
	_, err := p.db.Exec(p.ctx, geometryType)
	if err != nil {
		return fmt.Errorf("error creating geometry type: %w", err)
	}

	searchIndexTable := fmt.Sprintf(`
	create table if not exists %[1]s (
		id 					serial,
		feature_id 			varchar (8) 			not null ,
		collection_id 		text					not null,
		collection_version 	int 					not null,
		display_name 		text					not null,
		suggest 			text					not null,
		geometry_type 		geometry_type			not null,
		bbox 				geometry(polygon, %[2]d) null,
		primary key (id, collection_id, collection_version)
	) -- partition by list(collection_id);`, index, t.WGS84) // TODO partitioning comes later
	_, err = p.db.Exec(p.ctx, searchIndexTable)
	if err != nil {
		return fmt.Errorf("error creating search index table: %w", err)
	}

	fullTextSearchColumn := fmt.Sprintf(`
		alter table %[1]s add column ts tsvector 
	    generated always as (to_tsvector('dutch', suggest || display_name )) stored;`, index)
	_, err = p.db.Exec(p.ctx, fullTextSearchColumn)
	if err != nil {
		return fmt.Errorf("error creating full-text search column: %w", err)
	}

	ginIndex := fmt.Sprintf(`create index ts_idx on  %[1]s using gin(ts);`, index)
	_, err = p.db.Exec(p.ctx, ginIndex)
	if err != nil {
		return fmt.Errorf("error creating GIN index: %w", err)
	}
	return err
}
