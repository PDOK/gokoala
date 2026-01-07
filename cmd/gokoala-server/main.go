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
	rewritesFileFlag        = "rewrites-file"
	synonymsFileFlag        = "synonyms-file"
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
		&cli.PathFlag{
			Name:     rewritesFileFlag,
			EnvVars:  []string{strcase.ToScreamingSnake(rewritesFileFlag)},
			Usage:    "path to CSV file containing rewrites used to generate suggestions. Only for OGC API Features Search.",
			Required: false,
		},
		&cli.PathFlag{
			Name:     synonymsFileFlag,
			EnvVars:  []string{strcase.ToScreamingSnake(synonymsFileFlag)},
			Usage:    "path to CSV file containing synonyms used to generate suggestions. Only for OGC API Features Search.",
			Required: false,
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

		address := net.JoinHostPort(c.String(hostFlag), strconv.Itoa(c.Int(portFlag)))
		debugPort := c.Int(debugPortFlag)
		shutdownDelay := c.Int(shutdownDelayFlag)
		configFile := c.String(configFileFlag)
		themeFile := c.String(themeFileFlag)
		openAPIFile := c.String(openAPIFileFlag)
		trailingSlash := c.Bool(enableTrailingSlashFlag)
		cors := c.Bool(enableCorsFlag)

		// Engine encapsulates shared non-OGCAPI specific logic
		engine, err := eng.NewEngine(configFile, themeFile, openAPIFile, trailingSlash, cors)
		if err != nil {
			return err
		}
		// Each OGC API building block makes use of said Engine
		err = ogc.SetupBuildingBlocks(engine, c.String(rewritesFileFlag), c.String(synonymsFileFlag))
		if err != nil {
			return err
		}

		return engine.Start(address, debugPort, shutdownDelay)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
