package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	gokoalaEngine "github.com/PDOK/gokoala/engine"
	"github.com/PDOK/gokoala/ogc/common/core"
	"github.com/PDOK/gokoala/ogc/common/geospatial"
	"github.com/PDOK/gokoala/ogc/features"
	"github.com/PDOK/gokoala/ogc/geovolumes"
	"github.com/PDOK/gokoala/ogc/processes"
	"github.com/PDOK/gokoala/ogc/styles"
	"github.com/PDOK/gokoala/ogc/tiles"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/urfave/cli/v2"
)

var (
	cliFlags = []cli.Flag{
		&cli.StringFlag{
			Name:     "host",
			Usage:    "bind host for OGC server",
			Value:    "0.0.0.0",
			Required: false,
			EnvVars:  []string{"HOST"},
		},
		&cli.IntFlag{
			Name:     "port",
			Usage:    "bind port for OGC server",
			Value:    8080,
			Required: false,
			EnvVars:  []string{"PORT"},
		},
		&cli.IntFlag{
			Name:     "debug-port",
			Usage:    "bind port for debug server (disabled by default), do not expose this port publicly",
			Value:    -1,
			Required: false,
			EnvVars:  []string{"DEBUG_PORT"},
		},
		&cli.IntFlag{
			Name:     "shutdown-delay",
			Usage:    "delay (in seconds) before initiating graceful shutdown (e.g. useful in k8s to allow ingress controller to update their endpoints list)",
			Value:    0,
			Required: false,
			EnvVars:  []string{"SHUTDOWN_DELAY"},
		},
		&cli.StringFlag{
			Name:     "config-file",
			Usage:    "reference to YAML configuration file",
			Required: true,
			EnvVars:  []string{"CONFIG_FILE"},
		},
		&cli.StringFlag{
			Name:     "openapi-file",
			Usage:    "reference to a (customized) OGC OpenAPI spec for the dynamic parts of your OGC API",
			Required: false,
			EnvVars:  []string{"OPENAPI_FILE"},
		},
		&cli.BoolFlag{
			Name:     "enable-trailing-slash",
			Usage:    "allow API calls to URLs with a trailing slash.",
			Value:    false, // to satisfy https://gitdocumentatie.logius.nl/publicatie/api/adr/#api-48
			Required: false,
			EnvVars:  []string{"ALLOW_TRAILING_SLASH"},
		},
		&cli.BoolFlag{
			Name:     "enable-cors",
			Usage:    "enable Cross-Origin Resource Sharing (CORS) as required by OGC API specs. Disable if you handle CORS elsewhere.",
			Value:    false,
			Required: false,
			EnvVars:  []string{"ENABLE_CORS"},
		},
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "GoKoala"
	app.Usage = "Cloud Native OGC APIs server, written in Go"
	app.Flags = cliFlags
	app.Action = func(c *cli.Context) error {
		log.Printf("%s - %s\n", app.Name, app.Usage)

		address := net.JoinHostPort(c.String("host"), strconv.Itoa(c.Int("port")))
		debugPort := c.Int("debug-port")
		shutdownDelay := c.Int("shutdown-delay")
		configFile := c.String("config-file")
		openAPIFile := c.String("openapi-file")

		// Engine encapsulates shared non-OGC API specific logic
		engine := gokoalaEngine.NewEngine(configFile, openAPIFile)

		router := newRouter(engine, c.Bool("enable-trailing-slash"), c.Bool("enable-cors"))

		return engine.Start(address, router, debugPort, shutdownDelay)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func newRouter(engine *gokoalaEngine.Engine, allowTrailingSlash bool, enableCORS bool) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RealIP)
	router.Use(middleware.GetHead) // support HEAD requests https://docs.ogc.org/is/17-069r4/17-069r4.html#_http_1_1
	if allowTrailingSlash {
		router.Use(middleware.StripSlashes)
	}
	if enableCORS {
		router.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{http.MethodGet, http.MethodHead, http.MethodOptions},
			AllowedHeaders:   []string{gokoalaEngine.HeaderRequestedWith},
			ExposedHeaders:   []string{gokoalaEngine.HeaderContentCrs, gokoalaEngine.HeaderLink},
			AllowCredentials: false,
			MaxAge:           int((time.Hour * 24).Seconds()),
		}))
	}
	// implements https://gitdocumentatie.logius.nl/publicatie/api/adr/#api-57
	router.Use(middleware.SetHeader(gokoalaEngine.HeaderAPIVersion, engine.Config.Version))
	router.Use(middleware.Compress(5, gokoalaEngine.CompressibleMediaTypes...)) // enable gzip responses

	// OGC Common Part 1, will always be started
	core.NewCommonCore(engine, router)

	// OGC Common part 2
	if engine.Config.HasCollections() {
		geospatial.NewCollections(engine, router)
	}
	// OGC 3D GeoVolumes API
	if engine.Config.OgcAPI.GeoVolumes != nil {
		geovolumes.NewThreeDimensionalGeoVolumes(engine, router)
	}
	// OGC Tiles API
	if engine.Config.OgcAPI.Tiles != nil {
		tiles.NewTiles(engine, router)
	}
	// OGC Styles API
	if engine.Config.OgcAPI.Styles != nil {
		styles.NewStyles(engine, router)
	}
	// OGC Features API
	if engine.Config.OgcAPI.Features != nil {
		features.NewFeatures(engine, router)
	}
	// OGC Processes API
	if engine.Config.OgcAPI.Processes != nil {
		processes.NewProcesses(engine, router)
	}

	// Resources endpoint to serve static assets
	if engine.Config.Resources != nil {
		gokoalaEngine.NewResourcesEndpoint(engine, router)
	}
	// Health endpoint
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		gokoalaEngine.SafeWrite(w.Write, []byte("OK"))
	})

	return router
}
