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
