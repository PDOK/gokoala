package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/PDOK/gokoala/internal/ogc/features_search/etl"
	"github.com/PDOK/gokoala/internal/ogc/features_search/etl/config"
	"github.com/google/uuid"
	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/language"
)

const (
	appName = "gokoala-etl"

	configFileFlag   = "config-file"
	collectionIDFlag = "collection-id"
	revisionFlag     = "revision"
	dbHostFlag       = "db-host"
	dbNameFlag       = "db-name"
	dbPasswordFlag   = "db-password"
	dbPortFlag       = "db-port"
	dbSslModeFlag    = "db-ssl-mode"
	dbUsernameFlag   = "db-username"
	searchIndexFlag  = "search-index"
	sridFlag         = "srid"
	fileFlag         = "file"
	pageSizeFlag     = "page-size"
	skipOptimizeFlag = "skip-optimize"
	languageFlag     = "lang"
)

var (
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
	app.Usage = "Run an ETL (Extract-Transform-Load) process to populate the Features Search API."
	app.UseShortOptionHandling = true
	app.Commands = []*cli.Command{
		{
			Name:  "create-search-index",
			Usage: "Create an empty search index in the database",
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
				&cli.IntFlag{
					Name:     sridFlag,
					EnvVars:  []string{strcase.ToScreamingSnake(sridFlag)},
					Usage:    "SRID search-index bbox column, e.g. 28992 (RD) or 4326 (WSG84). The source geopackage its bbox should be in the same SRID.",
					Required: false,
					Value:    28992,
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
			Name:  "get-revision",
			Usage: "Get the revision (UUID) of a collection in the search index",
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
				&cli.StringFlag{
					Name:     collectionIDFlag,
					Usage:    "ID/name of the collection in the search index to get the version of",
					Required: true,
					EnvVars:  []string{strcase.ToScreamingSnake(collectionIDFlag)},
				},
			},
			Action: func(c *cli.Context) error {
				dbConn := flagsToDBConnStr(c)
				revision, err := etl.GetRevision(dbConn, c.String(collectionIDFlag), c.String(searchIndexFlag))
				fmt.Println(revision)
				return err
			},
		},
		{
			Name:  "import-file",
			Usage: "Import a file (e.g. GeoPackage) into the search index",
			Flags: []cli.Flag{
				commonDBFlags[dbHostFlag],
				commonDBFlags[dbPortFlag],
				commonDBFlags[dbNameFlag],
				commonDBFlags[dbUsernameFlag],
				commonDBFlags[dbPasswordFlag],
				commonDBFlags[dbSslModeFlag],
				&cli.StringFlag{
					Name:     configFileFlag,
					Usage:    "Reference to YAML configuration file",
					Required: true,
					EnvVars:  []string{strcase.ToScreamingSnake(configFileFlag)},
				},
				&cli.StringFlag{
					Name:     revisionFlag,
					Usage:    "Revision number of the data in the collection, should be a UUID",
					Required: true,
					EnvVars:  []string{strcase.ToScreamingSnake(revisionFlag)},
				},
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
				revision := c.String(revisionFlag)
				_, err = uuid.Parse(revision)
				if err != nil {
					return fmt.Errorf("invalid revision %s, must be a UUID", revision)
				}
				file := c.Path(fileFlag)
				for _, coll := range cfg.Collections {
					err = etl.ImportFile(coll, c.String(searchIndexFlag), revision,
						file, c.Int(pageSizeFlag), c.Bool(skipOptimizeFlag), dbConn)

					if err != nil {
						return fmt.Errorf("failed to import collection %s from file %s, error: %w",
							coll.ID, file, err)
					}
				}
				return nil
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
