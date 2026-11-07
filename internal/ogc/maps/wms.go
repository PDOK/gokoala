package maps

import (
	"errors"
	"fmt"
	"math"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
	"github.com/pdok/ogc-specifications/pkg/wms130"
)

const wmsCRSNamespace = "CRS"

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

	mapResolution, err := toMapResolution(mapRequest.resolved.mmPerPixel)
	if err != nil {
		return nil, err
	}

	return &wms130.GetMapRequest{
		BaseRequest:           wms130.BaseRequest{Service: wms130.Service, Version: wms130.Version},
		StyledLayerDescriptor: *styledLayerDescriptor,
		CRS:                   *crs,
		BoundingBox:           *boundingBox,
		Output:                *getOutput(mapRequest),
		VendorSpecificParameters: map[string]string{
			"map_resolution": fmt.Sprintf("%f", mapResolution),
		},
	}, nil

}

func toMapResolution(mmPerPixel float64) (float64, error) {
	const minMapResolution = 0.00001

	mapResolution := 25.4 / mmPerPixel
	if math.IsNaN(mapResolution) || math.IsInf(mapResolution, 0) || mapResolution < minMapResolution {
		return 0, fmt.Errorf("mm-per-pixel %g results in an invalid map resolution", mmPerPixel)
	}
	return mapResolution, nil
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
	srid, err := mapCRSToSRID(mapRequest.resolved.crs)
	if err != nil {
		return nil, err
	}
	if srid == domain.WGS84SRID {
		return &wms130.CRS{Namespace: wmsCRSNamespace, Code: 84}, nil
	}

	return &wms130.CRS{
		Namespace: "EPSG",
		Code:      int(srid),
	}, nil
}

func getBoundingBox(mapRequest *MapRequest) (*wms130.BoundingBox, error) {
	bbox := mapRequest.resolved.bbox
	if bbox == nil {
		return nil, errors.New("bbox is required")
	}
	return &wms130.BoundingBox{
		LowerCorner: wms130.Position{bbox.Min(0), bbox.Min(1)},
		UpperCorner: wms130.Position{bbox.Max(0), bbox.Max(1)},
	}, nil
}

func getOutput(mapRequest *MapRequest) *wms130.Output {
	return &wms130.Output{
		Size: wms130.Size{
			Width:  mapRequest.resolved.width,
			Height: mapRequest.resolved.height,
		},
		Format:      mapRequest.format,
		Transparent: &mapRequest.transparent,
	}
}
