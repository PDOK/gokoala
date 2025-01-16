package transform

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"text/template"

	"github.com/PDOK/gomagpie/config"
	"github.com/PDOK/gomagpie/internal/engine"
	"github.com/go-spatial/geom"
	pggeom "github.com/twpayne/go-geom" // this lib has a large overlap with github.com/go-spatial/geom but we need it to integrate with postgres
)

const WGS84 = 4326

type RawRecord struct {
	FeatureID    int64
	FieldValues  []any
	Bbox         *geom.Extent
	GeometryType string
}

type SearchIndexRecord struct {
	FeatureID         string
	CollectionID      string
	CollectionVersion int
	DisplayName       string
	Suggest           string
	GeometryType      string
	Bbox              *pggeom.Polygon
}

type Transformer struct {
	substAndSynonyms *SubstAndSynonyms
}

func NewTransformer(substitutionsFile string, synonymsFile string) (*Transformer, error) {
	substAndSynonyms, err := NewSubstAndSynonyms(substitutionsFile, synonymsFile)
	return &Transformer{substAndSynonyms}, err
}

func (t Transformer) Transform(records []RawRecord, collection config.GeoSpatialCollection) ([]SearchIndexRecord, error) {
	result := make([]SearchIndexRecord, 0, len(records))
	for _, r := range records {
		fieldValuesByName, err := slicesToStringMap(collection.Search.Fields, r.FieldValues)
		if err != nil {
			return nil, err
		}
		displayName, err := t.renderTemplate(collection.Search.DisplayNameTemplate, fieldValuesByName)
		if err != nil {
			return nil, err
		}
		allFieldValuesByName := t.substAndSynonyms.generate(fieldValuesByName)
		suggestions := make([]string, 0, len(collection.Search.ETL.SuggestTemplates))
		for i := range allFieldValuesByName {
			for _, suggestTemplate := range collection.Search.ETL.SuggestTemplates {
				suggestion, err := t.renderTemplate(suggestTemplate, allFieldValuesByName[i])
				if err != nil {
					return nil, err
				}
				suggestions = append(suggestions, suggestion)
			}
		}
		suggestions = slices.Compact(suggestions)

		bbox, err := r.transformBbox()
		if err != nil {
			return nil, err
		}

		// create target record(s)
		for _, suggestion := range suggestions {
			resultRecord := SearchIndexRecord{
				FeatureID:         strconv.FormatInt(r.FeatureID, 10),
				CollectionID:      collection.ID,
				CollectionVersion: collection.Search.Version,
				DisplayName:       displayName,
				Suggest:           suggestion,
				GeometryType:      r.GeometryType,
				Bbox:              bbox,
			}
			result = append(result, resultRecord)
		}
	}
	return result, nil
}

func (t Transformer) renderTemplate(templateFromConfig string, fieldValuesByName map[string]string) (string, error) {
	parsedTemplate, err := template.New("").
		Funcs(engine.GlobalTemplateFuncs).
		Option("missingkey=zero").
		Parse(templateFromConfig)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	if err = parsedTemplate.Execute(&b, fieldValuesByName); err != nil {
		return "", err
	}
	return strings.TrimSpace(b.String()), err
}

func (r RawRecord) transformBbox() (*pggeom.Polygon, error) {
	if strings.EqualFold(r.GeometryType, "POINT") {
		r.Bbox = r.Bbox.ExpandBy(0.1) // create valid bbox in case original geom is a point by expanding it a bit
	}
	if r.Bbox.Area() <= 0 {
		return nil, errors.New("bbox area must be greater than zero")
	}
	// convert bbox to polygon type supported by Postgres db driver
	polygon, err := pggeom.NewPolygon(pggeom.XY).SetCoords([][]pggeom.Coord{{
		{r.Bbox.MinX(), r.Bbox.MinY()},
		{r.Bbox.MaxX(), r.Bbox.MinY()},
		{r.Bbox.MaxX(), r.Bbox.MaxY()},
		{r.Bbox.MinX(), r.Bbox.MaxY()},
		{r.Bbox.MinX(), r.Bbox.MinY()},
	}})
	if err != nil {
		return nil, err
	}
	polygon = polygon.SetSRID(WGS84)
	if polygon.Area() <= 0 {
		return nil, errors.New("polygon area must be greater than zero")
	}
	return polygon, nil
}

func slicesToStringMap(keys []string, values []any) (map[string]string, error) {
	if len(keys) != len(values) {
		return nil, fmt.Errorf("slices must be of the same length, got %d keys and %d values", len(keys), len(values))
	}
	result := make(map[string]string, len(keys))
	for i := range keys {
		value := values[i]
		if value != nil {
			stringValue := fmt.Sprintf("%v", value)
			result[keys[i]] = stringValue
		}
	}
	return result, nil
}
