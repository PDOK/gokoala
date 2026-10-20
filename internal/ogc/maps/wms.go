package maps

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/PDOK/gokoala/internal/ogc/features"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/PDOK/gokoala/internal/ogc/features/proj"
	"github.com/pdok/ogc-specifications/pkg/wms130"
	"github.com/twpayne/go-geom"
)

// Build a WMS GetMap request from the given MapRequest.
func buildGetMapRequest(mapRequest *MapRequest, mapsConfig *config.OgcAPIMaps) (*wms130.GetMapRequest, error) {

	if mapRequest.mapType == DatasetMap {
		return nil, errors.New("dataset maps are not supported")
	}

	// Find the current collection
	var collection *config.MapsCollection
	for i := range mapsConfig.Collections {
		if mapsConfig.Collections[i].ID == mapRequest.collection {
			collection = &mapsConfig.Collections[i]
			break
		}
	}
	if collection == nil {
		return nil, fmt.Errorf(
			"collection %q not found in Maps configuration",
			mapRequest.collection,
		)
	}

	styledLayerDescriptor, err := getStyledLayerDescriptor(mapRequest, collection)
	if err != nil {
		return nil, err
	}

	crs, err := getCRS(mapRequest)
	if err != nil {
		return nil, err
	}

	boundingBox, err := getBoundingBox(mapRequest)
	if err != nil {
		return nil, err
	}

	return &wms130.GetMapRequest{
		BaseRequest:           wms130.BaseRequest{Service: wms130.Service, Version: wms130.Version},
		StyledLayerDescriptor: *styledLayerDescriptor,
		CRS:                   *crs,
		BoundingBox:           *boundingBox,
		Output:                *getOutput(mapRequest),
	}, nil

}

func getStyledLayerDescriptor(mapRequest *MapRequest, collection *config.MapsCollection) (*wms130.StyledLayerDescriptor, error) {
	namedLayers := make([]wms130.NamedLayer, 0, len(collection.Layers))
	for _, layer := range collection.Layers {
		namedLayer := wms130.NamedLayer{Name: layer.Name}

		if mapRequest.style != "" {
			foundStyle := false
			for _, style := range layer.Styles {
				if style.ID == mapRequest.style {
					namedLayer.NamedStyle = &wms130.NamedStyle{Name: style.WMSStyle}
					foundStyle = true
					break
				}
			}
			if !foundStyle {
				return nil, fmt.Errorf(
					"style %q not found for layer %q in collection %q",
					mapRequest.style,
					layer.Name,
					mapRequest.collection,
				)
			}
		}

		namedLayers = append(namedLayers, namedLayer)
	}

	return &wms130.StyledLayerDescriptor{
		NamedLayer: namedLayers,
	}, nil
}

func getCRS(mapRequest *MapRequest) (*wms130.CRS, error) {
	crsURI := strings.TrimSpace(string(mapRequest.crs))
	if !strings.HasPrefix(crsURI, domain.CrsURIPrefix+"EPSG/0/") {
		return nil, fmt.Errorf("expected an EPSG CRS URI, got %q", mapRequest.crs)
	}

	srid, err := features.ParseCrsToSRID(
		url.Values{CrsParam: {crsURI}},
		CrsParam,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid CRS %q: %w", mapRequest.crs, err)
	}
	if srid <= 0 {
		return nil, fmt.Errorf("CRS code must be positive, got %q", mapRequest.crs)
	}

	return &wms130.CRS{
		Namespace: "EPSG",
		Code:      int(srid),
	}, nil
}

func getAxisOrder(srs string) (domain.AxisOrder, error) {
	projJSON, err := proj.ExecProjInfo(srs)
	if err != nil {
		return domain.AxisOrderXY, fmt.Errorf("get CRS information for %q: %w", srs, err)
	}

	axisOrder, err := proj.GetAxisOrder(projJSON)
	if err != nil {
		return domain.AxisOrderXY, fmt.Errorf("get axis order for %q: %w", srs, err)
	}

	return axisOrder, nil
}

func getBoundingBox(mapRequest *MapRequest) (*wms130.BoundingBox, error) {
	if mapRequest.bbox == nil {
		return nil, errors.New("bbox is required")
	}
	crs, err := getCRS(mapRequest)
	if err != nil {
		return nil, err
	}

	bbox := [4]float64{
		mapRequest.bbox.Min(0),
		mapRequest.bbox.Min(1),
		mapRequest.bbox.Max(0),
		mapRequest.bbox.Max(1),
	}

	inputSRID := mapRequest.bboxSRID.GetOrDefault()
	// In this case we need to transform the incoming bbox into the requested output CRS
	if crs.Code != inputSRID {
		inputSRS := fmt.Sprintf("EPSG:%d", inputSRID)
		if inputSRID == domain.WGS84SRID {
			inputSRS = "OGC:CRS84"
		}
		outputSRS := fmt.Sprintf("EPSG:%d", crs.Code)

		inputAxisOrder, err := getAxisOrder(inputSRS)
		if err != nil {
			return nil, err
		}
		outputAxisOrder, err := getAxisOrder(outputSRS)
		if err != nil {
			return nil, err
		}

		if inputAxisOrder == domain.AxisOrderYX {
			bbox = [4]float64{bbox[1], bbox[0], bbox[3], bbox[2]}
		}
		transformed, err := util.TransformBoundingBox(
			geom.NewBounds(geom.XY).Set(bbox[:]...), inputSRS, outputSRS,
		)
		if err != nil {
			return nil, fmt.Errorf("transform WMS bbox: %w", err)
		}
		bbox = [4]float64{transformed.Min(0), transformed.Min(1), transformed.Max(0), transformed.Max(1)}
		if outputAxisOrder == domain.AxisOrderYX {
			bbox = [4]float64{bbox[1], bbox[0], bbox[3], bbox[2]}
		}
	}

	minX, minY, maxX, maxY := bbox[0], bbox[1], bbox[2], bbox[3]

	return &wms130.BoundingBox{
		LowerCorner: wms130.Position{minX, minY},
		UpperCorner: wms130.Position{maxX, maxY},
	}, nil
}

func getOutput(mapRequest *MapRequest) *wms130.Output {
	return &wms130.Output{
		Size: wms130.Size{
			Width:  mapRequest.width,
			Height: mapRequest.height,
		},
		Format:      mapRequest.format,
		Transparent: &mapRequest.transparent,
	}
}
