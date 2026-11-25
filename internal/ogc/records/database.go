package records

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine/util"
	"github.com/PDOK/gokoala/internal/ogc/common/geospatial"
	"github.com/PDOK/gokoala/internal/ogc/features/cql"
	domain "github.com/PDOK/gokoala/internal/ogc/features/domain"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func openRecordsDatabase(datasource *config.Datasource) (*sqlx.DB, string, error) {
	if datasource == nil {
		return nil, "", fmt.Errorf("Records datasource is required")
	}
	if datasource.Postgres != nil {
		db, err := sqlx.Open("pgx", datasource.Postgres.ConnectionString())
		if err != nil {
			return nil, "", fmt.Errorf("open Records PostgreSQL datasource: %w", err)
		}
		if err := pingRecordsDatabase(db); err != nil {
			_ = db.Close()
			return nil, "", fmt.Errorf("connect to Records PostgreSQL datasource: %w", err)
		}
		return db, "postgres", nil
	}
	if datasource.GeoPackage != nil {
		if datasource.GeoPackage.Local == nil {
			return nil, "", fmt.Errorf("Records currently supports local GeoPackage files; cloud-backed GeoPackages are not supported")
		}
		file, err := filepath.Abs(datasource.GeoPackage.Local.File)
		if err != nil {
			return nil, "", fmt.Errorf("resolve Records GeoPackage path: %w", err)
		}
		fileURL := url.URL{Scheme: "file", Path: filepath.ToSlash(file)}
		query := fileURL.Query()
		query.Set("mode", "ro")
		fileURL.RawQuery = query.Encode()
		db, err := sqlx.Open("sqlite3", fileURL.String())
		if err != nil {
			return nil, "", fmt.Errorf("open Records GeoPackage datasource: %w", err)
		}
		if err := pingRecordsDatabase(db); err != nil {
			_ = db.Close()
			return nil, "", fmt.Errorf("connect to Records GeoPackage datasource: %w", err)
		}
		return db, "sqlite", nil
	}
	return nil, "", fmt.Errorf("Records datasource must specify PostgreSQL or GeoPackage")
}

func pingRecordsDatabase(db *sqlx.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return db.PingContext(ctx)
}

type catalogRow struct {
	ID           string `db:"catalog_id"`
	Title        any    `db:"title"`
	Description  any    `db:"description"`
	ContactPoint any    `db:"contact_point"`
	Publisher    any    `db:"publisher"`
	License      any    `db:"license_uri"`
	Properties   any    `db:"properties"`
}

type datasetRow struct {
	ID             string `db:"dataset_id"`
	Identifier     string `db:"identifier"`
	Type           string `db:"type"`
	AccessRights   string `db:"access_rights_uri"`
	Title          any    `db:"title"`
	Description    any    `db:"description"`
	ContactPoint   any    `db:"contact_point"`
	Creators       any    `db:"creators"`
	Publisher      any    `db:"publisher"`
	Themes         any    `db:"theme_uris"`
	Languages      any    `db:"language_uris"`
	Keywords       any    `db:"keywords"`
	Spatial        any    `db:"spatial"`
	Temporal       any    `db:"temporal"`
	Geometry       any    `db:"geometry_geojson"`
	Bbox           any    `db:"bbox"`
	RecordTime     any    `db:"record_time"`
	Properties     any    `db:"properties"`
	STACProperties any    `db:"stac_properties"`
	STACVersion    string `db:"stac_version"`
	STACExtensions any    `db:"stac_extensions"`
}

type distributionRow struct {
	ID          string         `db:"distribution_id"`
	AssetKey    string         `db:"asset_key"`
	AccessURL   string         `db:"access_url"`
	DownloadURL sql.NullString `db:"download_url"`
	MediaType   sql.NullString `db:"media_type"`
	Format      sql.NullString `db:"format_uri"`
	License     string         `db:"license_uri"`
	Title       any            `db:"title"`
	Description any            `db:"description"`
	ByteSize    any            `db:"byte_size"`
	Checksum    any            `db:"checksum"`
	Properties  any            `db:"properties"`
	STACFields  any            `db:"stac_fields"`
}

func loadRecordsFromDatabase(ctx context.Context, db *sqlx.DB, driver string,
	configured config.RecordsCollections) (config.RecordsCollections, error) {
	configuredByID := make(map[string]config.RecordsCollection, len(configured))
	for _, collection := range configured {
		configuredByID[collection.ID] = collection
	}
	var catalogRows []catalogRow
	catalogQuery := "SELECT catalog_id, title, description, contact_point, publisher, license_uri, properties FROM dcat_catalog ORDER BY catalog_id"
	if err := db.SelectContext(ctx, &catalogRows, catalogQuery); err != nil {
		return nil, fmt.Errorf("read DCAT catalogs: %w", err)
	}
	collections := make(config.RecordsCollections, 0, len(catalogRows))
	for _, row := range catalogRows {
		configCollection := configuredByID[row.ID]
		title := localizedLiteral(row.Title)
		description := localizedLiteral(row.Description)
		if title == "" || description == "" {
			return nil, fmt.Errorf("DCAT catalog %q requires a title and description", row.ID)
		}
		configCollection.ID = row.ID
		configCollection.Metadata = &config.GeoSpatialCollectionMetadata{
			Title:       stringPointer(title),
			Description: stringPointer(description),
		}
		catalogProperties := decodeJSONObject(row.Properties)
		if contactPoint := decodeJSONValue(row.ContactPoint); contactPoint != nil {
			catalogProperties["contacts"] = []any{recordContact(contactPoint)}
		}
		if publisher := decodeJSONValue(row.Publisher); publisher != nil {
			catalogProperties["publisher"] = publisher
		}
		if license := localizedLiteral(row.License); license != "" {
			catalogProperties["license"] = license
		}
		encodedCatalogProperties, err := encodeJSONProperties(catalogProperties)
		if err != nil {
			return nil, fmt.Errorf("encode DCAT catalog properties for %q: %w", row.ID, err)
		}
		configCollection.CatalogProperties = encodedCatalogProperties
		datasets, err := loadDatasetRecords(ctx, db, driver, row.ID)
		if err != nil {
			return nil, err
		}
		configCollection.Records = datasets
		collections = append(collections, configCollection)
	}
	return collections, nil
}

func loadDatasetRecords(ctx context.Context, db *sqlx.DB, driver string, catalogID string) (config.Records, error) {
	query := `SELECT id AS dataset_id, identifier, type, access_rights_uri, title, description, contact_point,
		creators, publisher, theme_uris, language_uris, keywords, spatial, temporal,
		geometry_geojson, bbox, record_time, properties, stac_properties, stac_version, stac_extensions
		FROM dcat_dataset WHERE catalog_id = ? ORDER BY id`
	var datasets []datasetRow
	if err := db.SelectContext(ctx, &datasets, db.Rebind(query), catalogID); err != nil {
		return nil, fmt.Errorf("read DCAT datasets for catalog %q: %w", catalogID, err)
	}
	records := make(config.Records, 0, len(datasets))
	for _, dataset := range datasets {
		if dataset.Identifier == "" || dataset.AccessRights == "" || localizedLiteral(dataset.Title) == "" ||
			localizedLiteral(dataset.Description) == "" || len(decodeJSONArray(dataset.Creators)) == 0 ||
			len(decodeJSONArray(dataset.Themes)) == 0 || dataset.ContactPoint == nil || dataset.Publisher == nil {
			return nil, fmt.Errorf("DCAT dataset %q is missing a required DCAT-AP-NL property", dataset.ID)
		}
		properties := decodeJSONObject(dataset.Properties)
		properties["identifier"] = dataset.Identifier
		properties["type"] = dataset.Type
		properties["accessRights"] = dataset.AccessRights
		properties["title"] = localizedLiteral(dataset.Title)
		properties["description"] = localizedLiteral(dataset.Description)
		properties["contacts"] = []any{recordContact(decodeJSONValue(dataset.ContactPoint))}
		properties["creators"] = decodeJSONArray(dataset.Creators)
		properties["publisher"] = decodeJSONValue(dataset.Publisher)
		properties["themes"] = recordThemes(decodeJSONArray(dataset.Themes))
		properties["keywords"] = decodeJSONArray(dataset.Keywords)
		properties["languages"] = recordLanguages(decodeJSONArray(dataset.Languages))
		properties["externalIds"] = []any{map[string]any{"scheme": "dcat", "value": dataset.Identifier}}
		if spatial := decodeJSONValue(dataset.Spatial); spatial != nil {
			properties["spatial"] = spatial
		}
		if temporal := decodeJSONValue(dataset.Temporal); temporal != nil {
			properties["temporal"] = temporal
		}
		for name, value := range decodeJSONObject(dataset.STACProperties) {
			properties[name] = value
		}
		recordTime := decodeRecordTime(dataset.RecordTime)
		if _, exists := properties["datetime"]; !exists {
			if recordTime == nil {
				return nil, fmt.Errorf("DCAT dataset %q needs record_time or stac_properties.datetime to form a valid STAC Item", dataset.ID)
			}
			applySTACTime(properties, recordTime)
		}
		geometry := decodeJSONValue(dataset.Geometry)
		bbox := decodeFloatSlice(dataset.Bbox)
		assets, err := loadDistributions(ctx, db, dataset.ID)
		if err != nil {
			return nil, err
		}
		if len(assets) == 0 {
			return nil, fmt.Errorf("DCAT dataset %q requires at least one Distribution to form a valid STAC Item", dataset.ID)
		}
		formats := make([]any, 0, len(assets))
		for assetKey, asset := range assets {
			format := map[string]any{"name": assetKey}
			if asset.Type != "" {
				format["mediaType"] = asset.Type
			}
			formats = append(formats, format)
		}
		properties["formats"] = formats
		encodedProperties, err := encodeJSONProperties(properties)
		if err != nil {
			return nil, fmt.Errorf("encode DCAT dataset properties for %q: %w", dataset.ID, err)
		}
		encodedGeometry, err := encodeJSONValue(geometry)
		if err != nil {
			return nil, fmt.Errorf("encode DCAT dataset geometry for %q: %w", dataset.ID, err)
		}
		record := &config.Record{
			ID:             dataset.ID,
			STACVersion:    dataset.STACVersion,
			STACExtensions: decodeJSONStringSlice(dataset.STACExtensions),
			Geometry:       encodedGeometry,
			Bbox:           bbox,
			Time:           recordTime,
			Properties:     encodedProperties,
			Assets:         assets,
		}
		record.STACExtensions = stacExtensionURIs(record)
		records = append(records, record)
	}
	if err := config.ValidateRecords(&config.OgcAPIRecords{Collections: config.RecordsCollections{{ID: catalogID, Records: records}}}); err != nil {
		return nil, err
	}
	return records, nil
}

func loadDistributions(ctx context.Context, db *sqlx.DB, datasetID string) (map[string]config.RecordsAsset, error) {
	query := `SELECT distribution_id, asset_key, access_url, download_url, media_type, format_uri,
		license_uri, title, description, byte_size, checksum, properties, stac_fields
		FROM dcat_distribution WHERE dataset_id = ? ORDER BY asset_key`
	var distributions []distributionRow
	if err := db.SelectContext(ctx, &distributions, db.Rebind(query), datasetID); err != nil {
		return nil, fmt.Errorf("read DCAT distributions for dataset %q: %w", datasetID, err)
	}
	assets := make(map[string]config.RecordsAsset, len(distributions))
	for _, distribution := range distributions {
		href := ""
		if distribution.DownloadURL.Valid {
			href = distribution.DownloadURL.String
		}
		if href == "" {
			href = distribution.AccessURL
		}
		if href == "" || distribution.License == "" {
			return nil, fmt.Errorf("DCAT distribution %q requires access_url and license_uri", distribution.ID)
		}
		fields := decodeJSONObject(distribution.Properties)
		for name, value := range decodeJSONObject(distribution.STACFields) {
			fields[name] = value
		}
		if distribution.ByteSize != nil {
			fields["file:size"] = distribution.ByteSize
		}
		fields["license"] = distribution.License
		assetType := ""
		if distribution.MediaType.Valid {
			assetType = distribution.MediaType.String
		}
		if assetType == "" && distribution.Format.Valid {
			assetType = distribution.Format.String
		}
		encodedFields, err := encodeJSONProperties(fields)
		if err != nil {
			return nil, fmt.Errorf("encode DCAT Distribution %q fields: %w", distribution.ID, err)
		}
		assets[distribution.AssetKey] = config.RecordsAsset{
			Href:        href,
			Title:       localizedLiteral(distribution.Title),
			Description: localizedLiteral(distribution.Description),
			Type:        assetType,
			Roles:       []string{"data"},
			Fields:      encodedFields,
		}
	}
	return assets, nil
}

func decodeJSONValue(raw any) any {
	if raw == nil {
		return nil
	}
	if value, ok := raw.(config.JSONValue); ok {
		decoded, err := value.Any()
		if err == nil {
			return decoded
		}
		return nil
	}
	var encoded []byte
	switch value := raw.(type) {
	case []byte:
		encoded = value
	case string:
		encoded = []byte(value)
	default:
		return raw
	}
	var result any
	if json.Unmarshal(encoded, &result) != nil {
		return raw
	}
	return result
}

func decodeJSONObject(raw any) map[string]any {
	value, _ := decodeJSONValue(raw).(map[string]any)
	if value == nil {
		return make(map[string]any)
	}
	return value
}

func decodeJSONArray(raw any) []any {
	value, _ := decodeJSONValue(raw).([]any)
	return value
}

func decodeJSONStringSlice(raw any) []string {
	values := decodeJSONArray(raw)
	result := make([]string, 0, len(values))
	for _, value := range values {
		if text, ok := value.(string); ok {
			result = append(result, text)
		}
	}
	return result
}

func decodeFloatSlice(raw any) []float64 {
	values := decodeJSONArray(raw)
	result := make([]float64, 0, len(values))
	for _, value := range values {
		switch number := value.(type) {
		case float64:
			result = append(result, number)
		case json.Number:
			if parsed, err := number.Float64(); err == nil {
				result = append(result, parsed)
			}
		}
	}
	return result
}

func localizedLiteral(raw any) string {
	value := decodeJSONValue(raw)
	if text, ok := value.(string); ok {
		return text
	}
	translations, ok := value.(map[string]any)
	if !ok || len(translations) == 0 {
		return ""
	}
	for _, language := range []string{"nl", "en"} {
		if text, ok := translations[language].(string); ok {
			return text
		}
	}
	keys := make([]string, 0, len(translations))
	for language := range translations {
		keys = append(keys, language)
	}
	sort.Strings(keys)
	return fmt.Sprint(translations[keys[0]])
}

func decodeRecordTime(raw any) *config.RecordTime {
	value := decodeJSONValue(raw)
	encoded, err := json.Marshal(value)
	if err != nil || string(encoded) == "null" {
		return nil
	}
	var recordTime config.RecordTime
	if json.Unmarshal(encoded, &recordTime) != nil {
		return nil
	}
	return &recordTime
}

func applySTACTime(properties map[string]any, recordTime *config.RecordTime) {
	switch {
	case recordTime.Timestamp != "":
		properties["datetime"] = recordTime.Timestamp
	case recordTime.Date != "":
		properties["datetime"] = recordTime.Date + "T00:00:00Z"
	case len(recordTime.Interval) == 2:
		properties["datetime"] = nil
		properties["start_datetime"] = recordTime.Interval[0]
		properties["end_datetime"] = recordTime.Interval[1]
	}
}

func stringPointer(value string) *string {
	return &value
}

func recordContact(raw any) map[string]any {
	contact := decodeJSONObject(raw)
	recordContact := make(map[string]any)
	if name := firstString(contact, "name", "fn", "fullName"); name != "" {
		recordContact["name"] = name
	}
	if organization := firstString(contact, "organization", "organization-name"); organization != "" {
		recordContact["organization"] = organization
	}
	if email := firstString(contact, "email", "hasEmail"); email != "" {
		recordContact["emails"] = []any{map[string]any{"value": strings.TrimPrefix(email, "mailto:")}}
	}
	if phone := firstString(contact, "phone", "hasTelephone"); phone != "" {
		recordContact["phones"] = []any{map[string]any{"value": phone}}
	}
	if website := firstString(contact, "url", "hasURL"); website != "" {
		recordContact["links"] = []any{map[string]any{"rel": "about", "type": "text/html", "href": website}}
	}
	if len(recordContact) == 0 {
		recordContact["organization"] = "DCAT contact point"
	}
	return recordContact
}

func recordThemes(uris []any) []any {
	result := make([]any, 0, len(uris))
	for _, rawURI := range uris {
		uri, ok := rawURI.(string)
		if !ok || uri == "" {
			continue
		}
		scheme := uri
		if index := strings.LastIndex(scheme, "/"); index > 0 {
			scheme = scheme[:index]
		}
		result = append(result, map[string]any{
			"concepts": []any{map[string]any{"id": uri}},
			"scheme":   scheme,
		})
	}
	return result
}

func recordLanguages(uris []any) []any {
	result := make([]any, 0, len(uris))
	for _, rawURI := range uris {
		uri, ok := rawURI.(string)
		if !ok {
			continue
		}
		code := uri[strings.LastIndex(uri, "/")+1:]
		switch strings.ToUpper(code) {
		case "NLD":
			code = "nl"
		case "ENG":
			code = "en"
		default:
			code = strings.ToLower(code)
		}
		result = append(result, map[string]any{"code": code})
	}
	return result
}

func firstString(values map[string]any, names ...string) string {
	for _, name := range names {
		if value, ok := values[name].(string); ok && value != "" {
			return value
		}
	}
	return ""
}

func cqlMatchedRecordIDs(ctx context.Context, db *sqlx.DB, driver string, collection config.RecordsCollection, expression string) (map[string]bool, error) {
	if !collection.Filters.CQL.IsEnabled() {
		return nil, fmt.Errorf("CQL filtering is not enabled for records collection %q", collection.ID)
	}
	queryables := make([]domain.Field, 0, len(collection.Filters.Properties))
	for _, queryable := range collection.Filters.Properties {
		queryables = append(queryables, domain.Field{Name: queryable.Name})
	}
	queryables = append(queryables, domain.Field{Name: "id"})
	cqlConfig := recordCQLConfig(collection.Filters.CQL)
	var listener cql.Listener
	if driver == "sqlite" {
		listener = cql.NewGeoPackageListener(util.DefaultRandomizer, queryables, domain.WGS84SRID,
			domain.AxisOrderXY, geospatial.Features, cqlConfig)
	} else {
		listener = cql.NewPostgresListener(util.DefaultRandomizer, queryables, domain.WGS84SRIDPostgis,
			domain.AxisOrderXY, geospatial.Features, cqlConfig)
	}
	parsed, err := cql.ParseToSQL(expression, listener)
	if err != nil {
		return nil, err
	}
	if parsed == nil || parsed.SQL == "" {
		return nil, fmt.Errorf("CQL filter is empty")
	}
	filterSQL := parsed.SQL
	if driver == "postgres" {
		for name := range parsed.Params {
			filterSQL = strings.ReplaceAll(filterSQL, "@"+name, ":"+name)
		}
	}
	params := map[string]any{"catalog_id": collection.ID}
	for name, value := range parsed.Params {
		params[name] = value
	}
	query, args, err := sqlx.Named("SELECT id FROM dcat_dataset WHERE catalog_id = :catalog_id AND ("+filterSQL+")", params)
	if err != nil {
		return nil, fmt.Errorf("bind CQL query parameters: %w", err)
	}
	rows, err := db.QueryxContext(ctx, db.Rebind(query), args...)
	if err != nil {
		return nil, fmt.Errorf("execute CQL query: %w", err)
	}
	defer rows.Close()
	ids := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

func recordCQLConfig(source config.CQL) config.CQL {
	disabled := false
	source.EnableBasicSpatialFunctions = &disabled
	source.EnableBasicSpatialFunctionsPlus = &disabled
	source.EnableSpatialFunctions = &disabled
	source.EnableTemporalFunctions = &disabled
	return source
}

func encodeJSONValue(value any) (config.JSONValue, error) {
	encoded, err := json.Marshal(value)
	return config.JSONValue(encoded), err
}

func encodeJSONProperties(values map[string]any) (map[string]config.JSONValue, error) {
	result := make(map[string]config.JSONValue, len(values))
	for key, value := range values {
		encoded, err := encodeJSONValue(value)
		if err != nil {
			return nil, err
		}
		result[key] = encoded
	}
	return result, nil
}
