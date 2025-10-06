package load

import (
	"context"
	"fmt"

	t "github.com/PDOK/gomagpie/internal/etl/transform"
	"github.com/jackc/pgx/v5"
	pgxgeom "github.com/twpayne/pgx-geom"
	"golang.org/x/text/language"
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
	err = createExtensions(ctx, db)
	if err != nil {
		return nil, err
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

func (p *Postgres) Load(collectionID string, records []t.SearchIndexRecord, index string) (int64, error) {
	partition := fmt.Sprintf(`create table if not exists %[1]s_%[2]s partition of %[1]s for values in ('%[2]s');`, index, collectionID)
	_, err := p.db.Exec(p.ctx, partition)
	if err != nil {
		return -1, fmt.Errorf("error creating partition: %s Error: %w", collectionID, err)
	}

	loaded, err := p.db.CopyFrom(
		p.ctx,
		pgx.Identifier{index},
		[]string{"feature_id", "external_fid", "collection_id", "collection_version", "display_name", "suggest", "geometry_type", "bbox", "geometry"},
		pgx.CopyFromSlice(len(records), func(i int) ([]interface{}, error) {
			r := records[i]
			return []any{r.FeatureID, r.ExternalFid, r.CollectionID, r.CollectionVersion, r.DisplayName, r.Suggest, r.GeometryType, r.Bbox, r.Geometry}, nil
		}),
	)
	if err != nil {
		return -1, fmt.Errorf("unable to copy records: %w", err)
	}
	return loaded, nil
}

func (p *Postgres) Optimize() error {
	_, err := p.db.Exec(p.ctx, `vacuum analyze;`)
	if err != nil {
		return fmt.Errorf("error performing vacuum analyze: %w", err)
	}
	return nil
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
	_, err := p.db.Exec(p.ctx, geometryType)
	if err != nil {
		return fmt.Errorf("error creating geometry type: %w", err)
	}

	textSearchConfig := `
		do $$ begin
		    create text search configuration custom_dict (copy = simple);
		exception
		    when unique_violation then null;
		end $$;`
	_, err = p.db.Exec(p.ctx, textSearchConfig)
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
	_, err = p.db.Exec(p.ctx, alterTextSearchConfig)
	if err != nil {
		return fmt.Errorf("error altering text search configuration: %w", err)
	}

	searchIndexTable := fmt.Sprintf(`
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
	    ts                  tsvector                 generated always as (to_tsvector('custom_dict', suggest)) stored,
		primary key (id, collection_id, collection_version)
	) partition by list(collection_id);`, index, srid)

	_, err = p.db.Exec(p.ctx, searchIndexTable)
	if err != nil {
		return fmt.Errorf("error creating search index table: %w", err)
	}

	// GIN indexes are best for text search
	ginIndex := fmt.Sprintf(`create index if not exists ts_idx on %[1]s using gin(ts);`, index)
	_, err = p.db.Exec(p.ctx, ginIndex)
	if err != nil {
		return fmt.Errorf("error creating GIN index: %w", err)
	}

	// GIST indexes for bbox and geometry columns, to support search within a bounding box
	geometryIndex := fmt.Sprintf(`create index if not exists geometry_idx on %[1]s using gist(geometry);`, index)
	_, err = p.db.Exec(p.ctx, geometryIndex)
	if err != nil {
		return fmt.Errorf("error creating GIST index: %w", err)
	}
	bboxIndex := fmt.Sprintf(`create index if not exists bbox_idx on %[1]s using gist(bbox);`, index)
	_, err = p.db.Exec(p.ctx, bboxIndex)
	if err != nil {
		return fmt.Errorf("error creating GIST index: %w", err)
	}

	// create custom collation to correctly handle "numbers in strings" when sorting results
	// see https://www.postgresql.org/docs/12/collation.html#id-1.6.10.4.5.7.5
	collation := fmt.Sprintf(`create collation if not exists custom_numeric (provider = icu, locale = '%s-u-kn-true');`, lang.String())
	_, err = p.db.Exec(p.ctx, collation)
	if err != nil {
		return fmt.Errorf("error creating numeric collation: %w", err)
	}

	// index used to pre-rank results when generic search terms are used
	preRankIndex := fmt.Sprintf(`create index if not exists pre_rank_idx on %[1]s (array_length(string_to_array(suggest, ' '), 1) asc, display_name collate "custom_numeric" asc);`, index)
	_, err = p.db.Exec(p.ctx, preRankIndex)
	if err != nil {
		return fmt.Errorf("error creating pre-rank index: %w", err)
	}

	return err
}

func createExtensions(ctx context.Context, db *pgx.Conn) error {
	for _, ext := range []string{"postgis", "unaccent"} {
		_, err := db.Exec(ctx, `create extension if not exists `+ext+`;`)
		if err != nil {
			return fmt.Errorf("error creating %s extension: %w", ext, err)
		}
	}
	return nil
}
