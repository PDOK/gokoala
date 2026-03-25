package load

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"golang.org/x/text/language"
)

const (
	indexNameFullText = "ts_idx"
	indexNameGeometry = "geometry_idx"
	indexNameBbox     = "bbox_idx"
	indexNamePreRank  = "pre_rank_idx"
)

var (
	postgresExtensions = []string{"postgis", "unaccent", "pg_prewarm", "pg_buffercache"}

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
		collection_id       text                      not null,
		revision 			uuid                      not null,
		revision_date       timestamptz default now() not null
	);`
)

// Init initialize search index. Should be idempotent!
//
// Since not all DDL in Postgres support the "if not exists" syntax we use a bit
// of pl/pgsql to make it idempotent.
func (p *Postgres) Init(index string, srid int, lang language.Tag) error {
	log.Printf("initializing search index %s", index)

	if err := p.createGeomType(); err != nil {
		return err
	}
	if err := p.createTextConfig(); err != nil {
		return err
	}
	if err := p.createTables(index, srid); err != nil {
		return err
	}
	if err := p.createCollation(lang); err != nil {
		return err
	}
	if err := p.createIndexes(index, false); err != nil {
		return err
	}
	if err := p.createFunctions(index); err != nil {
		return err
	}
	return nil
}

func (p *Postgres) createGeomType() error {
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
	return nil
}

func (p *Postgres) createTextConfig() error {
	textSearchConfig := `
		do $$ begin
		    create text search configuration custom_dict (copy = simple);
		exception
		    when unique_violation then null;
		end $$;`
	_, err := p.db.Exec(context.Background(), textSearchConfig)
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
	return nil
}

func (p *Postgres) createTables(index string, srid int) error {
	// create search index table
	_, err := p.db.Exec(context.Background(), fmt.Sprintf(tableDefinition, index, srid, "partition by list(collection_id)"))
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
	return nil
}

// create custom collation to correctly handle "numbers in strings" when sorting results
// see https://www.postgresql.org/docs/12/collation.html#id-1.6.10.4.5.7.5
func (p *Postgres) createCollation(lang language.Tag) error {
	collation := fmt.Sprintf(`create collation if not exists custom_numeric (provider = icu, locale = '%s-u-kn-true');`, lang.String())
	_, err := p.db.Exec(context.Background(), collation)
	if err != nil {
		return fmt.Errorf("error creating numeric collation: %w", err)
	}
	return nil
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

// create (utility) pl/pgsql functions
//
//nolint:funlen
func (p *Postgres) createFunctions(index string) error {
	preWarmFunc := fmt.Sprintf(`
-- function to prewarm all partitions of a table into the shared_buffers, including the given indexes
create or replace function gokoala_prewarm_partitions(
    idx_suffixes  	 text[],
    search_idx_table regclass default '%s'::regclass
)
returns void
language plpgsql
as $func$
declare
    part_oid   oid;
    part_name  text;
    idx_name   text;
begin
    for part_oid in
        select relid
        from pg_partition_tree(search_idx_table)
        where isleaf
    loop
        -- resolve partition name
        select relname into part_name
        from pg_class
        where oid = part_oid;

        -- prewarm table
        raise notice 'prewarming partition: %%', part_name;
        perform pg_prewarm(part_oid, 'buffer');

        -- prewarm matching index
        for idx_name in
            select ci.relname
            from pg_index i
            join pg_class ci on ci.oid = i.indexrelid
            where i.indrelid = part_oid
              and ci.relname = any (
                    select part_name || suffix
                    from unnest(idx_suffixes) as suffix
              )
        loop
            raise notice 'prewarming index: %%', idx_name;
            perform pg_prewarm(idx_name, 'buffer');
        end loop;

    end loop;
end;
$func$;
`, index)
	_, err := p.db.Exec(context.Background(), preWarmFunc)
	if err != nil {
		return fmt.Errorf("error creating pre-warm function: %w", err)
	}

	checkBufferCacheFunc := fmt.Sprintf(`
-- function to check shared_buffer cache, which is critical for performance
create or replace function gokoala_inspect_buffercache(
    search_idx_table regclass default '%s'::regclass
)
returns table (
    object_name text,
    kind text,
    total_size text,
    cached_size text,
    percentage_cached numeric
)
language plpgsql stable as
$$
begin
    return query
    with parts as (
        -- all partitions of this table
        select
            c.oid,
            c.relname,
            n.nspname as schema
        from pg_partition_tree(search_idx_table) p
        join pg_class c      on c.oid = p.relid
        join pg_namespace n  on n.oid = c.relnamespace
        where p.isleaf
    ),
    objects as (
        -- each partition + all its indexes
        select
            p.schema,
            p.relname,
            p.oid,
            'table' as kind
        from parts p

        union all

        select
            p.schema,
            ci.relname,
            ci.oid,
            'index' as kind
        from parts p
        join pg_index i  on i.indrelid = p.oid
        join pg_class ci on ci.oid = i.indexrelid
    ),
    sizes as (
        select
            o.schema,
            o.relname,
            o.kind,
            o.oid,
            pg_relation_size(o.oid) as total_bytes,
            pg_relation_filenode(o.oid) as filenode
        from objects o
    ),
	cache as (
		select
			b.relfilenode,
			b.relforknumber,
			count(*) * 8192 as cached_bytes
		from pg_buffercache b
		where b.relforknumber = 0
		group by b.relfilenode, b.relforknumber
	)
    select
        s.relname::text as object_name,
        s.kind::text,
        pg_size_pretty(s.total_bytes) as total_size,
        pg_size_pretty(coalesce(c.cached_bytes, 0)) as cached_size,
        round(
            100.0 * coalesce(c.cached_bytes, 0) / nullif(s.total_bytes, 0),
            2
        ) as percentage_cached
    from sizes s
    left join cache c on c.relfilenode = s.filenode and c.relforknumber = 0
    order by s.relname, percentage_cached desc;
end;
$$;
`, index)
	_, err = p.db.Exec(context.Background(), checkBufferCacheFunc)
	if err != nil {
		return fmt.Errorf("error creating inspect buffercache function: %w", err)
	}

	return nil
}
