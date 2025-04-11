package engine

import (
	"log"
	"net/http"
	"net/url"
	"time"
)

func newHealthEndpoint(e *Engine) {
	var target *url.URL
	if tilesConfig := e.Config.OgcAPI.Tiles; tilesConfig != nil {
		var err error
		switch {
		case tilesConfig.DatasetTiles != nil:
			target, err = url.Parse(tilesConfig.DatasetTiles.TileServer.String() + *tilesConfig.DatasetTiles.HealthCheck.TilePath)
		case len(tilesConfig.Collections) > 0 && tilesConfig.Collections[0].Tiles != nil:
			target, err = url.Parse(tilesConfig.Collections[0].Tiles.GeoDataTiles.TileServer.String() + *tilesConfig.Collections[0].Tiles.GeoDataTiles.HealthCheck.TilePath)
		default:
			log.Println("cannot determine health check tilepath, falling back to basic check")
		}
		if err != nil {
			log.Fatalf("invalid health check tilepath: %v", err)
		}
	}
	if target != nil {
		client := &http.Client{Timeout: time.Duration(500) * time.Millisecond}
		e.Router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
			resp, err := client.Head(target.String())
			if err != nil {
				// exact error is irrelevant for health monitoring, but log it for insight
				log.Printf("healthcheck failed: %v", err)
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
