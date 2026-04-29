package geopackage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path"
	"sync"

	"github.com/mattn/go-sqlite3"
	"github.com/qustavo/sqlhooks/v2"
)

const (
	SqliteDriverName = "sqlite3_with_extensions"

	// IgnoreAccentCollation custom collation
	IgnoreAccentCollation = "NOACCENT"
)

var once sync.Once

// LoadDriver Load sqlite (with extensions) once.
//
// Extensions are by default expected in /usr/lib. For spatialite you can
// alternatively/optionally set SPATIALITE_LIBRARY_PATH.
func LoadDriver() {
	once.Do(func() {
		spatialite := path.Join(os.Getenv("SPATIALITE_LIBRARY_PATH"), "mod_spatialite")

		driver := &sqlite3.SQLiteDriver{
			Extensions: []string{spatialite},

			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				// The 'und@colStrength=primary' Unicode collation allows accent/diacritics to be ignored.
				// https://sqlite.org/src/dir/ext/icu
				query := fmt.Sprintf("select icu_load_collation('und@colStrength=primary', '%s');", IgnoreAccentCollation)
				_, err := conn.Exec(query, nil)
				if err != nil {
					log.Fatalf(errICUNotEnabled+" - %v", err)
				}
				return err
			},
		}

		sql.Register(SqliteDriverName, sqlhooks.Wrap(driver, NewSQLLogFromEnv())) // add support for SQL logging
	})
}
