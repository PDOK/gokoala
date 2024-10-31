package etl

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func CreateSearchIndex(ctx context.Context, conn string) error {
	db, err := pgx.Connect(ctx, conn)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	defer func(db *pgx.Conn, ctx context.Context) {
		err := db.Close(ctx)
		if err != nil {
			log.Printf("failed to close database connection: %s", err)
		}
	}(db, ctx)

	geometryType := `
    create type geometry_type as enum ('POINT', 'MULTIPOINT', 'LINESTRING', 'MULTILINESTRING', 'POLYGON', 'MULTIPOLYGON');
    `

	_, err = db.Exec(ctx, geometryType)
	if err != nil {
		log.Printf("Error creating geometryType: %v\n", err)
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

	_, err = db.Exec(ctx, searchIndexTable)
	if err != nil {
		log.Printf("Error creating searchIndexTable: %v\n", err)
	}

	fullTextSearchColumn := `
	alter table search_index add column ts TSVECTOR
        generated always as (to_tsvector('dutch', component_thoroughfarename || component_postaldescriptor )) stored ;
    `
	//alter table search_index add column ts TSVECTOR
	//generated always as (to_tsvector('dutch', suggest || display_name )) stored ;
	//`
	_, err = db.Exec(ctx, fullTextSearchColumn)
	if err != nil {
		log.Printf("Error creating fullTextSearchColumn: %v\n", err)
	}

	ginIndex := `
	create index ts_idx on search_index using gin(ts);
    `

	_, err = db.Exec(ctx, ginIndex)
	if err != nil {
		log.Printf("Error creating ginIndex: %v\n", err)
	}

	return err
}
