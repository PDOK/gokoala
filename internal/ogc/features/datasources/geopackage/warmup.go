package geopackage

import (
	"errors"
	"fmt"
	"log"

	"github.com/PDOK/gokoala/config"

	"github.com/jmoiron/sqlx"
)

// warmUpFeatureTables executes a warmup query to speedup subsequent queries.
// This encompasses traversing index(es) to fill the local cache.
func warmUpFeatureTables(
	configuredCollections config.GeoSpatialCollections,
	featureTableByCollectionID map[string]*featureTable,
	db *sqlx.DB) error {

	for collID, table := range featureTableByCollectionID {
		if table == nil {
			return errors.New("given table can't be nil")
		}
		for _, coll := range configuredCollections {
			if coll.ID == collID && coll.Features != nil {
				if err := warmUpFeatureTable(table.TableName, db); err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
}

func warmUpFeatureTable(tableName string, db *sqlx.DB) error {
	query := fmt.Sprintf(`
select minx,maxx,miny,maxy from %[1]s where minx <= 0 and maxx >= 0 and miny <= 0 and maxy >= 0
`, tableName)

	log.Printf("start warm-up of feature table '%s'", tableName)
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to warm-up feature table '%s': %w", tableName, err)
	}
	log.Printf("end warm-up of feature table '%s'", tableName)
	return nil
}
