package maps

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/PDOK/gokoala/config"
	"github.com/pdok/ogc-specifications/pkg/wms130"
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
	// TODO CRS handling needs to support URIs.
	namespace, codeText, ok := strings.Cut(strings.TrimSpace(string(mapRequest.crs)), ":")
	if !ok || codeText == "" {
		return nil, fmt.Errorf("invalid CRS %q", mapRequest.crs)
	}

	code, err := strconv.Atoi(codeText)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid CRS code in %q: %w", mapRequest.crs, err,
		)
	}

	return &wms130.CRS{
		Namespace: namespace,
		Code:      code,
	}, nil
}

func getBoundingBox(mapRequest *MapRequest) (*wms130.BoundingBox, error) {
	crs, err := getCRS(mapRequest)
	if err != nil {
		return nil, err
	}

	// In this case we need to transform the incoming bbox into the requested output CRS
	if crs.Code != mapRequest.bboxCRS.GetOrDefault() {

		// TODO Implement reprojection logic
		return nil, errors.New(
			"CRS reprojection is not supported",
		)
	}

	// TODO Handle coordinate order for CRSes in AxisOrder YX
	minX, minY, maxX, maxY := mapRequest.bbox[0], mapRequest.bbox[1], mapRequest.bbox[2], mapRequest.bbox[3]

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
