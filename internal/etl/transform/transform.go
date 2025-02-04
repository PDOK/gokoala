package transform

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"text/template"

	"github.com/PDOK/gomagpie/config"
	"github.com/PDOK/gomagpie/internal/engine"
	"github.com/twpayne/go-geom"
)

const WGS84 = 4326

type RawRecord struct {
	FeatureID    int64
	FieldValues  []any
	Bbox         *geom.Bounds
	GeometryType string
}

type SearchIndexRecord struct {
	FeatureID         string
	CollectionID      string
	CollectionVersion int
	DisplayName       string
	Suggest           string
	GeometryType      string
	Bbox              *geom.Polygon
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

func (r RawRecord) transformBbox() (*geom.Polygon, error) {
	if strings.EqualFold(r.GeometryType, "POINT") {
		// create valid bbox in case original geom is a point by expanding it a bit (eventually we'll replace this with something better)
		minx := r.Bbox.Min(0) - 0.1
		miny := r.Bbox.Min(1) - 0.1
		maxx := r.Bbox.Max(0) + 0.1
		maxy := r.Bbox.Max(1) + 0.1
		r.Bbox = geom.NewBounds(geom.XY).Set(minx, miny, maxx, maxy)
	}
	if surfaceArea(r.Bbox) <= 0 {
		return nil, errors.New("bbox area must be greater than zero")
	}
	return r.Bbox.Polygon().SetSRID(WGS84), nil
}

// Copied from https://github.com/PDOK/gokoala/blob/070ec77b2249553959330ff8029bfdf48d7e5d86/internal/ogc/features/url.go#L264
func surfaceArea(bbox *geom.Bounds) float64 {
	// Use the same logic as bbox.Area() in https://github.com/go-spatial/geom to calculate surface area.
	// The bounds.Area() in github.com/twpayne/go-geom behaves differently and is not what we're looking for.
	return math.Abs((bbox.Max(1) - bbox.Min(1)) * (bbox.Max(0) - bbox.Min(0)))
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
