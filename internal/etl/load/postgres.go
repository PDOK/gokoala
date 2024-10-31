package load

import (
	"context"
	"fmt"

	"github.com/PDOK/gomagpie/config"
	t "github.com/PDOK/gomagpie/internal/etl/transform"
	"github.com/jackc/pgx/v5"
)

type Postgis struct {
	db  *pgx.Conn
	ctx context.Context
}

func NewPostgis(dbConn string) (*Postgis, error) {
	ctx := context.Background()
	db, err := pgx.Connect(ctx, dbConn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	return &Postgis{db: db, ctx: ctx}, nil
}

func (p *Postgis) Close() error {
	return p.db.Close(p.ctx)
}

func (p *Postgis) Load(records []t.RawRecord, collection config.GeoSpatialCollection) (int64, error) {
	loaded, err := p.db.CopyFrom(
		context.Background(),
		pgx.Identifier{"search_index"},
		[]string{"component_thoroughfarename",
			"component_postaldescriptor",
			"component_addressareaname"},
		pgx.CopyFromSlice(len(records), func(i int) ([]interface{}, error) {
			//searchIndexRecord, err := records[i].Transform()
			//if err != nil {
			//	return nil, err
			//}
			return records[i].FieldValues, nil
		}),
	)
	if err != nil {
		return -1, fmt.Errorf("unable to copy records: %w", err)
	}
	return loaded, nil
}

// Init initialize search index
func (p *Postgis) Init() error {
	geometryType := `
    create type geometry_type as enum ('POINT', 'MULTIPOINT', 'LINESTRING', 'MULTILINESTRING', 'POLYGON', 'MULTIPOLYGON');
    `
	_, err := p.db.Exec(p.ctx, geometryType)
	if err != nil {
		return fmt.Errorf("error creating geometry type: %w", err)
	}

	searchIndexTable := `
    create table if not exists search_index (
		id 					serial,
	    component_thoroughfarename varchar,
        component_postaldescriptor varchar,
        component_addressareaname varchar,
	    primary key (id)
	);
    `
	//create table if not exists search_index (
	//	id 					serial,
	//	feature_id 			varchar (8) 			not null ,
	//	collection_id 		text					not null,
	//	collection_version 	int 					not null,
	//	display_name 		text					not null,
	//	suggest 			text					not null,
	//	geometry_type 		geometry_type			not null,
	//	bbox 				geometry(POLYGON,4326)	not null,
	//	primary key (id, collection_id, collection_version)
	//) partition by list(collection_id);

	_, err = p.db.Exec(p.ctx, searchIndexTable)
	if err != nil {
		return fmt.Errorf("error creating search index table: %w", err)
	}

	fullTextSearchColumn := `
	alter table search_index add column ts TSVECTOR
        generated always as (to_tsvector('dutch', component_thoroughfarename || component_postaldescriptor )) stored ;
    `
	//alter table search_index add column ts TSVECTOR
	//generated always as (to_tsvector('dutch', suggest || display_name )) stored ;
	//`
	_, err = p.db.Exec(p.ctx, fullTextSearchColumn)
	if err != nil {
		return fmt.Errorf("error creating full-text search column: %w", err)
	}

	ginIndex := `
	create index ts_idx on search_index using gin(ts);
    `
	_, err = p.db.Exec(p.ctx, ginIndex)
	if err != nil {
		return fmt.Errorf("error creating GIN index: %w", err)
	}
	return err
}
