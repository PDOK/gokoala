package etl

func ImportGeoPackage(gpkgPath string, synonymspath string, substitutionsPath string, targetDbConn string) error {
	// determine fields to query, from config
	// query rows (select + rows.next) to slice of structs, with limit+offset
	// transform data
	// copy data to postgres using pgx.copyfromslice
	return nil
}
