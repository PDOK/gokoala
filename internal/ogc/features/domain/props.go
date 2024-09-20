package domain

import (
	"slices"
	"strings"

	"github.com/PDOK/gokoala/internal/engine/util"
	perfjson "github.com/goccy/go-json"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// FeatureProperties the properties of a GeoJSON Feature. Properties are either unordered
// (default, and has the best performance!) or ordered in a specific way as described in the config.
type FeatureProperties struct {
	unordered map[string]any
	ordered   orderedmap.OrderedMap[string, any]
}

func NewFeatureProperties(order bool) FeatureProperties {
	return NewFeaturePropertiesWithData(order, make(map[string]any))
}

func NewFeaturePropertiesWithData(order bool, data map[string]any) FeatureProperties {
	if order {
		// properties are allowed to contain anything, including for example XML/GML
		ordered := *orderedmap.New[string, any](orderedmap.WithDisableHTMLEscape[string, any]())
		for k, v := range data {
			ordered.Set(k, v)
		}
		return FeatureProperties{ordered: ordered}
	}
	return FeatureProperties{unordered: data}
}

// MarshalJSON returns the JSON representation of either the ordered or unordered properties
func (p *FeatureProperties) MarshalJSON() ([]byte, error) {
	if p.unordered != nil {
		// properties are allowed to contain anything, including for example XML/GML.
		return perfjson.MarshalWithOption(p.unordered, perfjson.DisableHTMLEscape())
	}
	return p.ordered.MarshalJSON()
}

func (p *FeatureProperties) Value(key string) any {
	if p.unordered != nil {
		return p.unordered[key]
	}
	return p.ordered.Value(key)
}

func (p *FeatureProperties) Delete(key string) {
	if p.unordered != nil {
		delete(p.unordered, key)
	} else {
		p.ordered.Delete(key)
	}
}

func (p *FeatureProperties) Set(key string, value any) {
	if p.unordered != nil {
		p.unordered[key] = value
	} else {
		p.ordered.Set(key, value)
	}
}

func (p *FeatureProperties) SetRelation(key string, value any, existingKeyPrefix string) {
	p.Set(key, value)
	p.moveKeyBeforePrefix(key, existingKeyPrefix)
}

// moveKeyBeforePrefix best-effort algorithm to place the feature relation BEFORE any similarly named keys.
// For example, places "building.href" before "building_fk" or "building_fid".
func (p *FeatureProperties) moveKeyBeforePrefix(key string, keyPrefix string) {
	if p.unordered != nil {
		return
	}
	var existingKey string
	for pair := p.ordered.Oldest(); pair != nil; pair = pair.Next() {
		if strings.HasPrefix(pair.Key, keyPrefix) {
			existingKey = pair.Key
			break
		}
	}
	if existingKey != "" {
		_ = p.ordered.MoveBefore(key, existingKey)
	}
}

// Keys of the Feature properties.
//
// Note: In the future we might replace this with Go 1.23 iterators (range-over-func) however at the moment this
// isn't supported in Go templates: https://github.com/golang/go/pull/68329
func (p *FeatureProperties) Keys() []string {
	if p.unordered != nil {
		keys := util.Keys(p.unordered)
		slices.Sort(keys) // preserve alphabetical order
		return keys
	}
	result := make([]string, 0, p.ordered.Len())
	for pair := p.ordered.Oldest(); pair != nil; pair = pair.Next() {
		result = append(result, pair.Key)
	}
	return result
}
