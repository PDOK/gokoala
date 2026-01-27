package config

import (
	"log"
	"sort"

	"dario.cat/mergo"
	"github.com/PDOK/gokoala/internal/engine/types"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// GeoSpatialCollection Configuration for a collection of geospatial data.
//
// Interface/abstraction for common collection properties regardless of the specific
// type (e.g., tiles, features, 3dgeovolumes, etc.).
//
// +kubebuilder:object:generate:false
type GeoSpatialCollection interface {

	// GetID Unique ID of the collection
	GetID() string

	// GetMetadata Metadata describing the collection contents
	GetMetadata() *GeoSpatialCollectionMetadata

	// GetLinks Links pertaining to this collection (e.g., downloads, documentation)
	GetLinks() *CollectionLinks

	// Merge the (metadata and links) of the given collection with this collection. Return the merged collection.
	Merge(collection GeoSpatialCollection) GeoSpatialCollection
}

// GeoSpatialCollections All collections configured for this OGC API. Can contain a mix of tiles/features/etc.
// +kubebuilder:object:generate:false
type GeoSpatialCollections []GeoSpatialCollection

// HasCollections does this API offer collections, for example, with features, tiles, 3d tiles, etc.
func (c *Config) HasCollections() bool {
	return c.AllCollections() != nil
}

// AllCollections get all collections - for example, with features, tiles, 3d tiles - offered through this OGC API.
// Results are returned in alphabetic or literal order.
func (c *Config) AllCollections() GeoSpatialCollections {
	var result []GeoSpatialCollection
	if c.OgcAPI.GeoVolumes != nil {
		geoVolumes := types.ToInterfaceSlice[Collection3dGeoVolumes, GeoSpatialCollection](c.OgcAPI.GeoVolumes.Collections)
		result = append(result, geoVolumes...)
	}
	if c.OgcAPI.Tiles != nil {
		tiles := types.ToInterfaceSlice[CollectionTiles, GeoSpatialCollection](c.OgcAPI.Tiles.Collections)
		result = append(result, tiles...)
	}
	if c.OgcAPI.Features != nil {
		features := types.ToInterfaceSlice[CollectionFeatures, GeoSpatialCollection](c.OgcAPI.Features.Collections)
		result = append(result, features...)
	}
	if c.OgcAPI.FeaturesSearch != nil {
		featuresSearch := types.ToInterfaceSlice[CollectionFeaturesSearch, GeoSpatialCollection](c.OgcAPI.FeaturesSearch.Collections)
		result = append(result, featuresSearch...)
	}

	// sort
	if len(c.OgcAPICollectionOrder) > 0 {
		sortByLiteralOrder(result, c.OgcAPICollectionOrder)
	} else {
		sortByAlphabet(result)
	}

	return result
}

// Unique lists all unique GeoSpatialCollections (no duplicate IDs).
// Don't use in the hot path (creates a map on every invocation).
func (g GeoSpatialCollections) Unique() []GeoSpatialCollection {
	collectionsByID := g.toMap()
	result := make([]GeoSpatialCollection, 0, collectionsByID.Len())
	for pair := collectionsByID.Oldest(); pair != nil; pair = pair.Next() {
		result = append(result, pair.Value)
	}

	return result
}

// ContainsID check if a given collection - by ID - exists.
// Don't use in the hot path (creates a map on every invocation).
func (g GeoSpatialCollections) ContainsID(id string) bool {
	collectionsByID := g.toMap()
	_, ok := collectionsByID.Get(id)

	return ok
}

func (g GeoSpatialCollections) toMap() orderedmap.OrderedMap[string, GeoSpatialCollection] {
	collectionsByID := orderedmap.New[string, GeoSpatialCollection]()
	for _, current := range g {
		existing, ok := collectionsByID.Get(current.GetID())
		if ok {
			existing = existing.Merge(current)
			collectionsByID.Set(current.GetID(), existing)
		} else {
			collectionsByID.Set(current.GetID(), current)
		}
	}

	return *collectionsByID
}

func sortByAlphabet(collection []GeoSpatialCollection) {
	sort.Slice(collection, func(i, j int) bool {
		iName := collection[i].GetID()
		jName := collection[j].GetID()
		// prefer to sort by title when available, collection ID otherwise
		if collection[i].GetMetadata() != nil && collection[i].GetMetadata().Title != nil {
			iName = *collection[i].GetMetadata().Title
		}
		if collection[j].GetMetadata() != nil && collection[j].GetMetadata().Title != nil {
			jName = *collection[j].GetMetadata().Title
		}

		return iName < jName
	})
}

func sortByLiteralOrder(collections []GeoSpatialCollection, literalOrder []string) {
	collectionOrderIndex := make(map[string]int)
	for i, id := range literalOrder {
		collectionOrderIndex[id] = i
	}
	sort.Slice(collections, func(i, j int) bool {
		// sort, according to the explicit/literal order specified in OgcAPICollectionOrder
		return collectionOrderIndex[collections[i].GetID()] < collectionOrderIndex[collections[j].GetID()]
	})
}

func mergeMetadata(this GeoSpatialCollection, other GeoSpatialCollection) *GeoSpatialCollectionMetadata {
	return mergeField(this.GetID(), this.GetMetadata(), other.GetMetadata(), false)
}

func mergeLinks(this GeoSpatialCollection, other GeoSpatialCollection) *CollectionLinks {
	return mergeField(this.GetID(), this.GetLinks(), other.GetLinks(), true)
}

func mergeField[T any](id string, this *T, other *T, shouldAppend bool) *T {
	switch {
	case this == nil && other == nil:
		return nil
	case this == nil:
		return other
	case other == nil:
		return this
	}

	existing := *this
	var err error
	if shouldAppend {
		err = mergo.Merge(&existing, other, mergo.WithAppendSlice)
	} else {
		err = mergo.Merge(&existing, other)
	}
	if err != nil {
		log.Fatalf("failed to merge fields from 2 collections "+
			"with the same name '%s': %v", id, err)
		return nil
	}
	return &existing
}
