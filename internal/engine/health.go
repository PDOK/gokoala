package engine

import (
	"net/http"
	"net/url"
	"time"
)

func newHealthEndpoint(e *Engine) {
	if tilesConfig := e.Config.OgcAPI.Tiles; tilesConfig != nil {
		client := &http.Client{Timeout: time.Duration(500) * time.Millisecond}
		var target *url.URL
		if tilesConfig.DatasetTiles != nil {
			target, _ = url.Parse(tilesConfig.DatasetTiles.TileServer.String() + *tilesConfig.DatasetTiles.HealthCheck.TilePath)
		} else if tilesConfig.Collections != nil {
			target, _ = url.Parse(tilesConfig.Collections[0].Tiles.GeoDataTiles.TileServer.String() + *tilesConfig.Collections[0].Tiles.GeoDataTiles.HealthCheck.TilePath)
		}

		e.Router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
			resp, err := client.Head(target.String())
			if err != nil {
				// exact error is irrelevant for health monitoring
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(resp.StatusCode)
				resp.Body.Close()
			}
		})
	} else {
		e.Router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
			SafeWrite(w.Write, []byte("OK"))
		})
	}
}
