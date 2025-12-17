package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/PDOK/gokoala/internal/etl"
	"github.com/PDOK/gokoala/internal/etl/config"
	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/language"
)

const (
	appName = "gokoala-etl"

	hostFlag                = "host"
	portFlag                = "port"
	debugPortFlag           = "debug-port"
	shutdownDelayFlag       = "shutdown-delay"
	configFileFlag          = "config-file"
	collectionIDFlag        = "collection-id"
	collectionVersionFlag   = "collection-version"
	enableTrailingSlashFlag = "enable-trailing-slash"
	enableCorsFlag          = "enable-cors"
	dbHostFlag              = "db-host"
	dbNameFlag              = "db-name"
	dbPasswordFlag          = "db-password"
	dbPortFlag              = "db-port"
	dbSslModeFlag           = "db-ssl-mode"
	dbUsernameFlag          = "db-username"
	searchIndexFlag         = "search-index"
	sridFlag                = "srid"
	fileFlag                = "file"
	pageSizeFlag            = "page-size"
	skipOptimizeFlag        = "skip-optimize"
	languageFlag            = "lang"
)

var (
	serviceFlags = map[string]cli.Flag{
		hostFlag: &cli.StringFlag{
			Name:     hostFlag,
			Usage:    "bind host",
			Value:    "0.0.0.0",
			Required: false,
			EnvVars:  []string{strcase.ToScreamingSnake(hostFlag)},
		},
		portFlag: &cli.IntFlag{
			Name:     portFlag,
			Usage:    "bind port",
			Value:    8080,
			Required: false,
			EnvVars:  []string{strcase.ToScreamingSnake(portFlag)},
		},
		debugPortFlag: &cli.IntFlag{
			Name:     debugPortFlag,
			Usage:    "bind port for debug server (disabled by default), do not expose this port publicly",
			Value:    -1,
			Required: false,
			EnvVars:  []string{strcase.ToScreamingSnake(debugPortFlag)},
		},
		shutdownDelayFlag: &cli.IntFlag{
			Name:     shutdownDelayFlag,
			Usage:    "delay (in seconds) before initiating graceful shutdown (e.g. useful in k8s to allow ingress controller to update their endpoints list)",
			Value:    0,
			Required: false,
			EnvVars:  []string{strcase.ToScreamingSnake(shutdownDelayFlag)},
		},
		configFileFlag: &cli.StringFlag{
			Name:     configFileFlag,
			Usage:    "reference to YAML configuration file",
			Required: true,
			EnvVars:  []string{strcase.ToScreamingSnake(configFileFlag)},
		},
		collectionIDFlag: &cli.StringFlag{
			Name:     collectionIDFlag,
			Usage:    "reference to collection ID in the config file",
			Required: true,
			EnvVars:  []string{strcase.ToScreamingSnake(collectionIDFlag)},
		},
		collectionVersionFlag: &cli.StringFlag{
			Name:     collectionVersionFlag,
			Usage:    "version reference of the collection",
			Required: true,
			EnvVars:  []string{strcase.ToScreamingSnake(collectionVersionFlag)},
		},
		enableTrailingSlashFlag: &cli.BoolFlag{
			Name:     enableTrailingSlashFlag,
			Usage:    "allow API calls to URLs with a trailing slash.",
			Value:    false, // to satisfy https://gitdocumentatie.logius.nl/publicatie/api/adr/#api-48
			Required: false,
			EnvVars:  []string{strcase.ToScreamingSnake(enableTrailingSlashFlag)},
		},
		enableCorsFlag: &cli.BoolFlag{
			Name:     enableCorsFlag,
			Usage:    "enable Cross-Origin Resource Sharing (CORS) as required by OGC API specs. Disable if you handle CORS elsewhere.",
			Value:    false,
			Required: false,
			EnvVars:  []string{strcase.ToScreamingSnake(enableCorsFlag)},
		},
		sridFlag: &cli.IntFlag{
			Name:     sridFlag,
			EnvVars:  []string{strcase.ToScreamingSnake(sridFlag)},
			Usage:    "SRID search-index bbox column, e.g. 28992 (RD) or 4326 (WSG84). The source geopackage its bbox should be in the same SRID.",
			Required: false,
			Value:    28992,
		},
	}

	commonDBFlags = map[string]cli.Flag{
		dbHostFlag: &cli.StringFlag{
			Name:     dbHostFlag,
			Value:    "localhost",
			Required: false,
			EnvVars:  []string{strcase.ToScreamingSnake(dbHostFlag)},
		},
		dbPortFlag: &cli.IntFlag{
			Name:     dbPortFlag,
			Value:    5432,
			Required: false,
			EnvVars:  []string{strcase.ToScreamingSnake(dbPortFlag)},
		},
		dbNameFlag: &cli.StringFlag{
			Name:     dbNameFlag,
			Usage:    "Connect to this database",
			Value:    "postgres",
			Required: false,
			EnvVars:  []string{strcase.ToScreamingSnake(dbNameFlag)},
		},
		dbSslModeFlag: &cli.StringFlag{
			Name:     dbSslModeFlag,
			Value:    "disable",
			Required: false,
			EnvVars:  []string{strcase.ToScreamingSnake(dbSslModeFlag)},
		},
		dbUsernameFlag: &cli.StringFlag{
			Name:     dbUsernameFlag,
			Value:    "postgres",
			Required: false,
			EnvVars:  []string{strcase.ToScreamingSnake(dbUsernameFlag)},
		},
		dbPasswordFlag: &cli.StringFlag{
			Name:     dbPasswordFlag,
			Value:    "postgres",
			Required: false,
			EnvVars:  []string{strcase.ToScreamingSnake(dbPasswordFlag)},
		},
	}
)

//nolint:funlen
func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Usage = "Run location search and geocoding API, or use as CLI to support the ETL process for this API."
	app.UseShortOptionHandling = true
	app.Commands = []*cli.Command{
		{
			Name:     "create-search-index",
			Category: "etl",
			Usage:    "Create empty search index in database",
			Flags: []cli.Flag{
				commonDBFlags[dbHostFlag],
				commonDBFlags[dbPortFlag],
				commonDBFlags[dbNameFlag],
				commonDBFlags[dbUsernameFlag],
				commonDBFlags[dbPasswordFlag],
				commonDBFlags[dbSslModeFlag],
				&cli.PathFlag{
					Name:     searchIndexFlag,
					EnvVars:  []string{strcase.ToScreamingSnake(searchIndexFlag)},
					Usage:    "Name of search index to create",
					Required: false,
					Value:    "search_index",
				},
				serviceFlags[sridFlag],
				&cli.StringFlag{
					Name:     languageFlag,
					EnvVars:  []string{strcase.ToScreamingSnake(languageFlag)},
					Usage:    "What language will predominantly be used in the search index. Specify as a BCP 47 tag, like 'en', 'nl', 'de'",
					Required: false,
					Value:    "nl",
				},
			},
			Action: func(c *cli.Context) error {
				dbConn := flagsToDBConnStr(c)
				lang, err := language.Parse(c.String(languageFlag))
				if err != nil {
					return err
				}
				return etl.CreateSearchIndex(dbConn, c.String(searchIndexFlag), c.Int(sridFlag), lang)
			},
		},
		{
			Name:     "get-version",
			Category: "etl",
			Usage:    "Get the version of a collection in a search index",
			Flags: []cli.Flag{
				commonDBFlags[dbHostFlag],
				commonDBFlags[dbPortFlag],
				commonDBFlags[dbNameFlag],
				commonDBFlags[dbUsernameFlag],
				commonDBFlags[dbPasswordFlag],
				commonDBFlags[dbSslModeFlag],
				&cli.PathFlag{
					Name:     searchIndexFlag,
					EnvVars:  []string{strcase.ToScreamingSnake(searchIndexFlag)},
					Usage:    "Name of search index",
					Required: false,
					Value:    "search_index",
				},
				serviceFlags[collectionIDFlag],
			},
			Action: func(c *cli.Context) error {
				dbConn := flagsToDBConnStr(c)
				version, err := etl.GetVersion(dbConn, c.String(collectionIDFlag), c.String(searchIndexFlag))
				fmt.Println(version)
				return err
			},
		},
		{
			Name:     "import-file",
			Category: "etl",
			Usage:    "Import file into search index",
			Flags: []cli.Flag{
				commonDBFlags[dbHostFlag],
				commonDBFlags[dbPortFlag],
				commonDBFlags[dbNameFlag],
				commonDBFlags[dbUsernameFlag],
				commonDBFlags[dbPasswordFlag],
				commonDBFlags[dbSslModeFlag],
				serviceFlags[configFileFlag],
				serviceFlags[collectionIDFlag],
				serviceFlags[collectionVersionFlag],
				&cli.PathFlag{
					Name:     searchIndexFlag,
					EnvVars:  []string{strcase.ToScreamingSnake(searchIndexFlag)},
					Usage:    "Name of search index in which to import the given file",
					Required: false,
					Value:    "search_index",
				},
				&cli.PathFlag{
					Name:     fileFlag,
					EnvVars:  []string{strcase.ToScreamingSnake(fileFlag)},
					Usage:    "Path to (e.g GeoPackage) file to import",
					Required: true,
				},
				&cli.IntFlag{
					Name:     pageSizeFlag,
					EnvVars:  []string{strcase.ToScreamingSnake(pageSizeFlag)},
					Usage:    "Page/batch size to use when extracting records from file",
					Required: false,
					Value:    10000,
				},
				&cli.BoolFlag{
					Name:     skipOptimizeFlag,
					EnvVars:  []string{strcase.ToScreamingSnake(skipOptimizeFlag)},
					Usage:    "Skip running VACUUM ANALYZE on the search index after import",
					Required: false,
					Value:    false,
				},
			},
			Action: func(c *cli.Context) error {
				dbConn := flagsToDBConnStr(c)
				cfg, err := config.NewConfig(c.Path(configFileFlag))
				if err != nil {
					return err
				}
				collectionID := c.String(collectionIDFlag)
				collection := cfg.CollectionByID(collectionID)
				if collection == nil {
					return fmt.Errorf("no configured collection found with id: %s", collectionID)
				}
				return etl.ImportFile(*collection, c.String(searchIndexFlag), c.String(collectionVersionFlag),
					c.Path(fileFlag), c.Int(pageSizeFlag), c.Bool(skipOptimizeFlag), dbConn)
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func flagsToDBConnStr(c *cli.Context) string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s&application_name=%s",
		c.String(dbUsernameFlag), c.String(dbPasswordFlag), net.JoinHostPort(c.String(dbHostFlag),
			strconv.Itoa(c.Int(dbPortFlag))), c.String(dbNameFlag), c.String(dbSslModeFlag), appName)
}
