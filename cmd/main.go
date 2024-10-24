package main

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"log"
	"net"
	"os"
	"strconv"

	eng "github.com/PDOK/gomagpie/internal/engine"
	"github.com/PDOK/gomagpie/internal/etl"
	"github.com/PDOK/gomagpie/internal/ogc"
	"github.com/urfave/cli/v2"

	_ "go.uber.org/automaxprocs"
)

var (
	serverFlags = []cli.Flag{
		&cli.StringFlag{
			Name:     "host",
			Usage:    "bind host",
			Value:    "0.0.0.0",
			Required: false,
			EnvVars:  []string{"HOST"},
		},
		&cli.IntFlag{
			Name:     "port",
			Usage:    "bind port",
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
	commonDBFlags = []cli.Flag{
		&cli.StringFlag{
			Name:    "db-host",
			Value:   "localhost",
			EnvVars: []string{strcase.ToScreamingSnake("db-host")},
		},
		&cli.IntFlag{
			Name:    "db-port",
			Value:   5432,
			EnvVars: []string{strcase.ToScreamingSnake("db-port")},
		},
		&cli.StringFlag{
			Name:    "db-name",
			Usage:   "Connect to this database",
			EnvVars: []string{strcase.ToScreamingSnake("db-name")},
		},
		&cli.StringFlag{
			Name:    "db-ssl-mode",
			Value:   "disable",
			EnvVars: []string{strcase.ToScreamingSnake("db-ssl-mode")},
		},
		&cli.StringFlag{
			Name:    "db-username",
			Value:   "postgres",
			EnvVars: []string{strcase.ToScreamingSnake("db-username")},
		},
		&cli.StringFlag{
			Name:    "db-password",
			Value:   "postgres",
			EnvVars: []string{strcase.ToScreamingSnake("db-password")},
		},
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "gomagpie"
	app.Usage = "Run location search and geocoding API service, or use as CLI to support the ETL process for this service."
	app.UseShortOptionHandling = true
	app.Commands = []*cli.Command{
		{
			Name:  "run",
			Usage: "Run location search and geocoding API server",
			Description: `
Run location search and geocoding API server.
`,
			Action: func(c *cli.Context) error {
				log.Println(c.Command.Usage)

				address := net.JoinHostPort(c.String("host"), strconv.Itoa(c.Int("port")))
				debugPort := c.Int("debug-port")
				shutdownDelay := c.Int("shutdown-delay")
				configFile := c.String("config-file")
				trailingSlash := c.Bool("enable-trailing-slash")
				cors := c.Bool("enable-cors")

				// Engine encapsulates shared logic
				engine, err := eng.NewEngine(configFile, trailingSlash, cors)
				if err != nil {
					return err
				}
				// Each OGC API building block makes use of said Engine
				ogc.SetupBuildingBlocks(engine)

				return engine.Start(address, debugPort, shutdownDelay)
			},
			Flags: serverFlags,
		},
		{
			Name:  "create-search-index",
			Usage: "Create search index",
			Description: `
Create search index in database. This exists of a search table "zoek_index" with full text indices and prepared for partitioning
`,
			Action: func(c *cli.Context) error {
				dbConn := flagsToDBConnStr(c)
				return etl.CreateSearchIndex(c.Context, dbConn)
			},
			Flags: commonDBFlags,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func flagsToDBConnStr(c *cli.Context) string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s&application_name=%s",
		c.String("db-username"), c.String("db-password"), net.JoinHostPort(c.String("db-host"),
			strconv.Itoa(c.Int("db-port"))), c.String("db-name"), c.String("db-ssl-mode"),
		"gomagpie")
}
