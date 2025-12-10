package common

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/PDOK/gokoala/internal/ogc/common/geospatial"
	"github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

const (
	EnvLogSQL = "LOG_SQL"

	selectAll = "*"
)

// DatasourceCommon shared data and logic between data sources.
type DatasourceCommon struct {
	TransformOnTheFly bool
	QueryTimeout      time.Duration
	FidColumn         string
	ExternalFidColumn string
	MaxDecimals       int
	ForceUTC          bool

	TableByCollectionID           map[string]*Table
	PropertyFiltersByCollectionID map[string]datasources.PropertyFiltersWithAllowedValues
	PropertiesByCollectionID      map[string]*config.FeatureProperties
	RelationsByCollectionID       map[string][]config.Relation
}

// Table metadata about a table containing features or attributes in a data source.
type Table struct {
	Name               string
	Type               geospatial.CollectionType
	GeometryColumnName string
	GeometryType       string

	Schema *domain.Schema // required
}

func (dc *DatasourceCommon) GetSchema(collection string) (*domain.Schema, error) {
	table, err := dc.CollectionToTable(collection)
	if err != nil {
		return nil, err
	}

	return table.Schema, nil
}

func (dc *DatasourceCommon) GetCollectionType(collection string) (geospatial.CollectionType, error) {
	table, err := dc.CollectionToTable(collection)
	if err != nil {
		return "", err
	}

	return table.Type, nil
}

func (dc *DatasourceCommon) GetPropertyFiltersWithAllowedValues(collection string) datasources.PropertyFiltersWithAllowedValues {
	return dc.PropertyFiltersByCollectionID[collection]
}

func (dc *DatasourceCommon) SupportsOnTheFlyTransformation() bool {
	return dc.TransformOnTheFly
}

func (dc *DatasourceCommon) CollectionToTable(collection string) (*Table, error) {
	table, ok := dc.TableByCollectionID[collection]
	if !ok {
		return nil, fmt.Errorf("can't query collection '%s' since it doesn't exist in "+
			"datasource, available in datasource: %v", collection, util.Keys(dc.TableByCollectionID))
	}

	return table, nil
}

// SelectGeom function signature to select geometry from a table while taking axis order into account.
type SelectGeom func(order domain.AxisOrder, table *Table) string

// SelectColumns build select clause.
func (dc *DatasourceCommon) SelectColumns(table *Table, axisOrder domain.AxisOrder,
	selectGeom SelectGeom, selectRelation SelectRelation,
	propConfig *config.FeatureProperties, relationsConfig []config.Relation,
	includePrevNext bool) string {

	columns := orderedmap.New[string, struct{}]() // map (actually a set) to prevent accidental duplicate columns
	switch {
	case propConfig != nil:
		// select columns in a specific order (we need an ordered map for this purpose!)
		for _, prop := range propConfig.Properties {
			if prop != table.GeometryColumnName {
				columns.Set(prop, struct{}{})
			}
		}
		if !propConfig.PropertiesExcludeUnknown {
			// select missing columns according to the table schema
			for _, field := range table.Schema.Fields {
				if field.Name != table.GeometryColumnName {
					_, ok := columns.Get(field.Name)
					if !ok {
						columns.Set(field.Name, struct{}{})
					}
				}
			}
		}
	case table.Schema != nil:
		// select all columns according to the table schema
		for _, field := range table.Schema.Fields {
			if field.Name != table.GeometryColumnName {
				columns.Set(field.Name, struct{}{})
			}
		}
	default:
		log.Println("Warning: table doesn't have a schema. Can't select columns by name, selecting all")

		return selectAll
	}

	columns.Set(dc.FidColumn, struct{}{})
	if includePrevNext {
		columns.Set(domain.PrevFid, struct{}{})
		columns.Set(domain.NextFid, struct{}{})
	}

	// turn columns and subqueries into SQL string
	result := ColumnsToSQL(slices.Collect(columns.KeysFromOldest()), true)
	if includePrevNext {
		result += dc.relationsToSQL(relationsConfig, selectRelation, "nextprevfeat")
	} else {
		result += dc.relationsToSQL(relationsConfig, selectRelation, table.Name)
	}
	result += selectGeom(axisOrder, table)

	return result
}

func PropertyFiltersToSQL(pf map[string]string, symbol string) (sql string, namedParams map[string]any) {
	namedParams = make(map[string]any)
	var sqlBuilder strings.Builder
	sqlBuilder.WriteString(sql)

	if len(pf) > 0 {
		position := 0

		for k, v := range pf {
			position++
			namedParam := fmt.Sprintf("pf%d", position)
			// column name in double quotes in case it is a reserved keyword
			// also: we don't currently support LIKE since wildcard searches don't use the index
			sqlBuilder.WriteString(fmt.Sprintf(" and \"%s\" = %s%s", k, symbol, namedParam))
			namedParams[namedParam] = v
		}
	}

	return sqlBuilder.String(), namedParams
}

func TemporalCriteriaToSQL(temporalCriteria datasources.TemporalCriteria, symbol string) (sql string, namedParams map[string]any) {
	namedParams = make(map[string]any)
	if !temporalCriteria.ReferenceDate.IsZero() {
		namedParams["referenceDate"] = temporalCriteria.ReferenceDate
		startDate := temporalCriteria.StartDateProperty
		endDate := temporalCriteria.EndDateProperty
		sql = fmt.Sprintf(" and \"%[1]s\" <= %[3]sreferenceDate and (\"%[2]s\" >= %[3]sreferenceDate or \"%[2]s\" is null)",
			startDate, endDate, symbol)
	}

	return sql, namedParams
}

// ColumnsToSQL converts a slice of column names to a comma-separated string of column names.
//
// Beware: Always set escape=true to get the column names wrapped in double quotes, the only
// exception is when using subselects.
func ColumnsToSQL(columns []string, escape bool) string {
	if escape {
		return fmt.Sprintf("\"%s\"", strings.Join(columns, `", "`))
	}
	return strings.Join(columns, `", "`)
}

// SelectRelation function signature to select related features (using a many-to-many table).
// The resulting SQL query should return a string of comma-separated FIDs to the related features.
type SelectRelation func(relation config.Relation, relationName string, targetFID string, sourceTableAlias string) string

func (dc *DatasourceCommon) relationsToSQL(relations []config.Relation, selectRelation SelectRelation, sourceTableAlias string) string {
	result := ""
	if len(relations) == 0 {
		return result
	}

	subQueries := make([]string, 0)
	for _, relation := range relations {
		relationName := relation.RelatedCollection
		if relation.Prefix != "" {
			relationName += "_" + relation.Prefix
		}

		targetFID := relation.Columns.Target
		if dc.ExternalFidColumn != "" {
			targetFID = dc.ExternalFidColumn
			relationName += "_" + dc.ExternalFidColumn
		}

		subQueries = append(subQueries, selectRelation(relation, relationName, targetFID, sourceTableAlias))
	}

	if len(subQueries) > 0 {
		result += ", "
		result += ColumnsToSQL(subQueries, false)
	}
	return result
}

func ValidateUniqueness(result map[string]*Table) {
	uniqueTables := make(map[string]struct{})
	for _, table := range result {
		uniqueTables[table.Name] = struct{}{}
	}
	if len(uniqueTables) != len(result) {
		log.Printf("Warning: found %d unique table names for %d collections, "+
			"usually each collection is backed by its own unique table\n", len(uniqueTables), len(result))
	}
}
