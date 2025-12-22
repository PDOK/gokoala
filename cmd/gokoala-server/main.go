package main

import (
	"log"
	"net"
	"os"
	"strconv"

	"github.com/PDOK/gokoala/config"
	eng "github.com/PDOK/gokoala/internal/engine"
	"github.com/PDOK/gokoala/internal/ogc"
	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"
)

const (
	hostFlag                = "host"
	portFlag                = "port"
	debugPortFlag           = "debug-port"
	shutdownDelayFlag       = "shutdown-delay"
	configFileFlag          = "config-file"
	openAPIFileFlag         = "openapi-file"
	enableTrailingSlashFlag = "enable-trailing-slash"
	enableCorsFlag          = "enable-cors"
	themeFileFlag           = "theme-file"
)

var (
	cliFlags = []cli.Flag{
		&cli.StringFlag{
			Name:     hostFlag,
			Usage:    "bind host for OGC server",
			Value:    "0.0.0.0",
			Required: false,
			EnvVars:  []string{strcase.ToScreamingSnake(hostFlag)},
		},
		&cli.IntFlag{
			Name:     portFlag,
			Usage:    "bind port for OGC server",
			Value:    8080,
			Required: false,
			EnvVars:  []string{strcase.ToScreamingSnake(portFlag)},
		},
		&cli.IntFlag{
			Name:     debugPortFlag,
			Usage:    "bind port for debug server (disabled by default), do not expose this port publicly",
			Value:    -1,
			Required: false,
			EnvVars:  []string{strcase.ToScreamingSnake(debugPortFlag)},
		},
		&cli.IntFlag{
			Name:     shutdownDelayFlag,
			Usage:    "delay (in seconds) before initiating graceful shutdown (e.g. useful in k8s to allow ingress controller to update their endpoints list)",
			Value:    0,
			Required: false,
			EnvVars:  []string{strcase.ToScreamingSnake(shutdownDelayFlag)},
		},
		&cli.StringFlag{
			Name:     configFileFlag,
			Usage:    "reference to YAML configuration file",
			Required: true,
			EnvVars:  []string{strcase.ToScreamingSnake(configFileFlag)},
		},
		&cli.StringFlag{
			Name:     openAPIFileFlag,
			Usage:    "reference to a (customized) OGC OpenAPI spec for the dynamic parts of your OGC API",
			Required: false,
			EnvVars:  []string{strcase.ToScreamingSnake(openAPIFileFlag)},
		},
		&cli.BoolFlag{
			Name:     enableTrailingSlashFlag,
			Usage:    "allow API calls to URLs with a trailing slash.",
			Value:    false, // to satisfy https://gitdocumentatie.logius.nl/publicatie/api/adr/#api-48
			Required: false,
			EnvVars:  []string{strcase.ToScreamingSnake(enableTrailingSlashFlag)},
		},
		&cli.BoolFlag{
			Name:     enableCorsFlag,
			Usage:    "enable Cross-Origin Resource Sharing (CORS) as required by OGC API specs. Disable if you handle CORS elsewhere.",
			Value:    false,
			Required: false,
			EnvVars:  []string{strcase.ToScreamingSnake(enableCorsFlag)},
		},
		&cli.StringFlag{
			Name:     themeFileFlag,
			Usage:    "reference to a (customized) YAML configuration file for the theme",
			Required: false,
			EnvVars:  []string{strcase.ToScreamingSnake(themeFileFlag)},
		},
	}
)

func main() {
	app := cli.NewApp()
	app.Name = config.AppName
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
