package config

import (
	"log"
	"sort"

	"dario.cat/mergo"
)

type GeoSpatialCollections []GeoSpatialCollection

// Unique lists all unique GeoSpatialCollections (no duplicate IDs),
// return results in alphabetic order
func (g GeoSpatialCollections) Unique() []GeoSpatialCollection {
	collectionsByID := g.toMap()
	flattened := make([]GeoSpatialCollection, 0, len(collectionsByID))
	for _, v := range collectionsByID {
		flattened = append(flattened, v)
	}
	sort.Slice(flattened, func(i, j int) bool {
		icomp := flattened[i].ID
		jcomp := flattened[j].ID
		// prefer to sort by title when available, collection ID otherwise
		if flattened[i].Metadata != nil && flattened[i].Metadata.Title != nil {
			icomp = *flattened[i].Metadata.Title
		}
		if flattened[j].Metadata != nil && flattened[j].Metadata.Title != nil {
			jcomp = *flattened[j].Metadata.Title
		}
		return icomp < jcomp
	})
	return flattened
}

// ContainsID check if given collection - by ID - exists
func (g GeoSpatialCollections) ContainsID(id string) bool {
	_, ok := g.toMap()[id]
	return ok
}

func (g GeoSpatialCollections) toMap() map[string]GeoSpatialCollection {
	collectionsByID := make(map[string]GeoSpatialCollection)
	for _, current := range g {
		existing, ok := collectionsByID[current.ID]
		if ok {
			err := mergo.Merge(&existing, current)
			if err != nil {
				log.Fatalf("failed to merge 2 collections with the same name '%s': %v", current.ID, err)
			}
			collectionsByID[current.ID] = existing
		} else {
			collectionsByID[current.ID] = current
		}
	}
	return collectionsByID
}
