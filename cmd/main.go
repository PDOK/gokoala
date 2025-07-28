package main

import (
	"log"
	"net"
	"os"
	"strconv"

	eng "github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc"
	"github.com/urfave/cli/v2"

	_ "go.uber.org/automaxprocs"
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
		&cli.StringFlag{
			Name:     "theme-file",
			Usage:    "reference to a (customized) YAML configuration file for the theme",
			Required: false,
			EnvVars:  []string{"THEME_FILE"},
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
		themeFile := c.String("theme-file")
		openAPIFile := c.String("openapi-file")
		trailingSlash := c.Bool("enable-trailing-slash")
		cors := c.Bool("enable-cors")

		// Engine encapsulates shared non-OGC API specific logic
		engine, err := eng.NewEngine(configFile, themeFile, openAPIFile, trailingSlash, cors)
		if err != nil {
			return err
		}
		// Each OGC API building block makes use of said Engine
		ogc.SetupBuildingBlocks(engine)

		return engine.Start(address, debugPort, shutdownDelay)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
