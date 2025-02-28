package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/PDOK/gomagpie/config"
	"github.com/PDOK/gomagpie/internal/search"
	"github.com/iancoleman/strcase"
	"golang.org/x/text/language"

	eng "github.com/PDOK/gomagpie/internal/engine"
	"github.com/PDOK/gomagpie/internal/etl"
	"github.com/PDOK/gomagpie/internal/ogc"
	"github.com/urfave/cli/v2"

	_ "go.uber.org/automaxprocs"
)

const (
	appName = "gomagpie"

	hostFlag                 = "host"
	portFlag                 = "port"
	debugPortFlag            = "debug-port"
	shutdownDelayFlag        = "shutdown-delay"
	configFileFlag           = "config-file"
	collectionIDFlag         = "collection-id"
	enableTrailingSlashFlag  = "enable-trailing-slash"
	enableCorsFlag           = "enable-cors"
	dbHostFlag               = "db-host"
	dbNameFlag               = "db-name"
	dbPasswordFlag           = "db-password"
	dbPortFlag               = "db-port"
	dbSslModeFlag            = "db-ssl-mode"
	dbUsernameFlag           = "db-username"
	searchIndexFlag          = "search-index"
	fileFlag                 = "file"
	featureTableFlag         = "feature-table"
	featureTableFidFlag      = "fid"
	featureTableGeomFlag     = "geom"
	pageSizeFlag             = "page-size"
	rewritesFileFlag         = "rewrites-file"
	synonymsFileFlag         = "synonyms-file"
	languageFlag             = "lang"
	rankNormalization        = "rank-normalization"
	exactMatchMultiplier     = "exact-match-multiplier"
	primarySuggestMultiplier = "primary-suggest-multiplier"
	rankThreshold            = "rank-threshold"
	preRankLimit             = "pre-rank-limit"
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
			Name:  "start-service",
			Usage: "Start service to serve location API",
			Flags: []cli.Flag{
				serviceFlags[hostFlag],
				serviceFlags[portFlag],
				serviceFlags[debugPortFlag],
				serviceFlags[shutdownDelayFlag],
				serviceFlags[configFileFlag],
				serviceFlags[enableTrailingSlashFlag],
				serviceFlags[enableCorsFlag],
				commonDBFlags[dbHostFlag],
				commonDBFlags[dbPortFlag],
				commonDBFlags[dbNameFlag],
				commonDBFlags[dbUsernameFlag],
				commonDBFlags[dbPasswordFlag],
				commonDBFlags[dbSslModeFlag],
				&cli.PathFlag{
					Name:    searchIndexFlag,
					EnvVars: []string{strcase.ToScreamingSnake(searchIndexFlag)},
					Usage:   "Name of search index to use",
					Value:   "search_index",
				},
				&cli.PathFlag{
					Name:     rewritesFileFlag,
					EnvVars:  []string{strcase.ToScreamingSnake(rewritesFileFlag)},
					Usage:    "Path to csv file containing rewrites.csv used to generate suggestions",
					Required: true,
				},
				&cli.PathFlag{
					Name:     synonymsFileFlag,
					EnvVars:  []string{strcase.ToScreamingSnake(synonymsFileFlag)},
					Usage:    "Path to csv file containing synonyms used to generate suggestions",
					Required: true,
				},
				&cli.IntFlag{
					Name:     rankNormalization,
					EnvVars:  []string{strcase.ToScreamingSnake(rankNormalization)},
					Usage:    "Normalization specifies whether and how a document's length should impact its rank. Possible values are 0, 1, 2, 4, 8, 16 and 32. For more information see https://www.postgresql.org/docs/current/textsearch-controls.html",
					Required: false,
					Value:    1,
				},
				&cli.Float64Flag{
					Name:     exactMatchMultiplier,
					EnvVars:  []string{strcase.ToScreamingSnake(exactMatchMultiplier)},
					Usage:    "Multiply the exact match rank to boost it above the wildcard matches",
					Required: false,
					Value:    3.0,
				},
				&cli.Float64Flag{
					Name:     primarySuggestMultiplier,
					EnvVars:  []string{strcase.ToScreamingSnake(primarySuggestMultiplier)},
					Usage:    "The primary suggest is equal to the display name. With this multiplier you can boost it above other suggests",
					Required: false,
					Value:    1.01,
				},
				&cli.IntFlag{
					Name:     rankThreshold,
					EnvVars:  []string{strcase.ToScreamingSnake(rankThreshold)},
					Usage:    "The threshold above which results are pre-ranked instead ranked exactly",
					Required: false,
					Value:    40000,
				},
				&cli.IntFlag{
					Name:     preRankLimit,
					EnvVars:  []string{strcase.ToScreamingSnake(preRankLimit)},
					Usage:    "The number of results which are pre-ranked when the rank threshold is hit",
					Required: false,
					Value:    400,
				},
			},
			Action: func(c *cli.Context) error {
				log.Println(c.Command.Usage)

				address := net.JoinHostPort(c.String(hostFlag), strconv.Itoa(c.Int(portFlag)))
				debugPort := c.Int(debugPortFlag)
				shutdownDelay := c.Int(shutdownDelayFlag)
				configFile := c.String(configFileFlag)
				trailingSlash := c.Bool(enableTrailingSlashFlag)
				cors := c.Bool(enableCorsFlag)

				dbConn := flagsToDBConnStr(c)

				// Engine encapsulates shared logic
				engine, err := eng.NewEngine(configFile, trailingSlash, cors)
				if err != nil {
					return err
				}
				// Each OGC API building block makes use of said Engine
				ogc.SetupBuildingBlocks(engine, dbConn)
				// Create search endpoint
				_, err = search.NewSearch(
					engine,
					dbConn,
					c.String(searchIndexFlag),
					c.Path(rewritesFileFlag),
					c.Path(synonymsFileFlag),
					c.Int(rankNormalization),
					c.Float64(exactMatchMultiplier),
					c.Float64(primarySuggestMultiplier),
					c.Int(rankThreshold),
					c.Int(preRankLimit),
				)
				if err != nil {
					return err
				}
				return engine.Start(address, debugPort, shutdownDelay)
			},
		},
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
				return etl.CreateSearchIndex(dbConn, c.String(searchIndexFlag), lang)
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
				&cli.StringFlag{
					Name:     featureTableFidFlag,
					EnvVars:  []string{strcase.ToScreamingSnake(featureTableFidFlag)},
					Usage:    "Name of feature ID field in file",
					Required: false,
					Value:    "fid",
				},
				&cli.StringFlag{
					Name:     featureTableGeomFlag,
					EnvVars:  []string{strcase.ToScreamingSnake(featureTableGeomFlag)},
					Usage:    "Name of geometry field in file",
					Required: false,
					Value:    "geom",
				},
				&cli.StringFlag{
					Name:     featureTableFlag,
					EnvVars:  []string{strcase.ToScreamingSnake(featureTableFlag)},
					Usage:    "Name of the table in given file to import",
					Required: true,
				},
				&cli.IntFlag{
					Name:     pageSizeFlag,
					EnvVars:  []string{strcase.ToScreamingSnake(pageSizeFlag)},
					Usage:    "Page/batch size to use when extracting records from file",
					Required: false,
					Value:    10000,
				},
			},
			Action: func(c *cli.Context) error {
				dbConn := flagsToDBConnStr(c)
				cfg, err := config.NewConfig(c.Path(configFileFlag))
				if err != nil {
					return err
				}
				featureTable := config.FeatureTable{
					Name: c.String(featureTableFlag),
					FID:  c.String(featureTableFidFlag),
					Geom: c.String(featureTableGeomFlag),
				}
				collectionID := c.String(collectionIDFlag)
				collection := config.CollectionByID(cfg, collectionID)
				if collection == nil {
					return fmt.Errorf("no configured collection found with id: %s", collectionID)
				}
				return etl.ImportFile(*collection, c.String(searchIndexFlag), c.Path(fileFlag), featureTable,
					c.Int(pageSizeFlag), dbConn)
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
