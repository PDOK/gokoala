package maps

import (
	"net/url"
)

type mapsCollectionURL struct {
	baseURL url.URL
	params  url.Values
}
