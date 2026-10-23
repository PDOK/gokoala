package records

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PDOK/gokoala/config"
	"github.com/PDOK/gokoala/internal/engine"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

const (
	pluralItemsPath   = "/collections/{collectionId}/items"
	singularItemPath  = "/collection/{collectionId}/items/{recordId}"
	pluralItemPath    = pluralItemsPath + "/{recordId}"
	sortablesPath     = "/collections/{collectionId}/sortables"
	queryablesPath    = "/collections/{collectionId}/queryables"
	sortablesRelation = "http://www.opengis.net/def/rel/ogc/1.0/sortables"
	mediaTypeGeoJSON  = "application/geo+json"
	mediaTypeHTML     = "text/html"
	mediaTypeJSON     = "application/json"
	mediaTypeMarkdown = "text/markdown"
	defaultLimit      = 10
	maximumLimit      = 10000
	templatesDir      = "internal/ogc/records/templates/"
)

var (
	itemsHTMLKey = engine.NewTemplateKey(templatesDir + "items.go.html")
	itemsMDKey   = engine.NewTemplateKey(templatesDir + "items.go.md")
	itemHTMLKey  = engine.NewTemplateKey(templatesDir + "item.go.html")
	itemMDKey    = engine.NewTemplateKey(templatesDir + "item.go.md")
)

type Records struct {
	engine         *engine.Engine
	collections    map[string]config.RecordsCollection
	database       *sqlx.DB
	databaseDriver string
}

func NewRecords(e *engine.Engine) (*Records, error) {
	if e.Templates != nil {
		e.ParseTemplate(itemsHTMLKey)
		e.ParseTemplate(itemsMDKey)
		e.ParseTemplate(itemHTMLKey)
		e.ParseTemplate(itemMDKey)
	}
	collections := make(map[string]config.RecordsCollection)
	if e.Config.OgcAPI.Records != nil {
		for _, collection := range e.Config.OgcAPI.Records.Collections {
			collections[collection.ID] = collection
		}
	}

	result := &Records{engine: e, collections: collections}
	if e.Config.OgcAPI.Records != nil && e.Config.OgcAPI.Records.Datasource != nil {
		database, driver, err := openRecordsDatabase(e.Config.OgcAPI.Records.Datasource)
		if err != nil {
			return nil, err
		}
		loadedCollections, err := loadRecordsFromDatabase(context.Background(), database, driver, e.Config.OgcAPI.Records.Collections)
		if err != nil {
			_ = database.Close()
			return nil, err
		}
		e.Config.OgcAPI.Records.Collections = loadedCollections
		for _, collection := range loadedCollections {
			collections[collection.ID] = collection
		}
		if e.OpenAPI != nil {
			e.RebuildOpenAPI(nil)
		}
		result.database = database
		result.databaseDriver = driver
	}
	return result, nil
}

func (rs *Records) Close() error {
	if rs.database == nil {
		return nil
	}
	return rs.database.Close()
}

func (rs *Records) HasCollection(id string) bool {
	_, ok := rs.collections[id]
	return ok
}

func (rs *Records) RegisterItemRoutes() {
	rs.engine.Router.Get(pluralItemsPath, rs.Items())
	rs.engine.Router.Get(pluralItemPath, rs.Item())
	rs.RegisterAuxiliaryRoutes()
}

func (rs *Records) RegisterAuxiliaryRoutes() {
	rs.engine.Router.Get(singularItemPath, rs.Item())
	rs.engine.Router.Get(sortablesPath, rs.Sortables())
	rs.engine.Router.Get(queryablesPath, rs.Queryables())
}

func (rs *Records) Items() http.HandlerFunc {
	return rs.handleItems
}

func (rs *Records) handleItems(w http.ResponseWriter, r *http.Request) {
	if !rs.validateRequest(w, r) {
		return
	}
	collectionID := chi.URLParam(r, "collectionId")
	collection, ok := rs.collections[collectionID]
	if !ok {
		engine.RenderProblem(engine.ProblemNotFound, w)
		return
	}

	matchedRecords, err := selectRecords(collection.Records, r.URL.Query())
	if err != nil {
		engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
		return
	}
	if filter := r.URL.Query().Get("filter"); filter != "" {
		if r.URL.Query().Get("filter-lang") != "" && r.URL.Query().Get("filter-lang") != "cql2-text" {
			engine.RenderProblem(engine.ProblemBadRequest, w, "Records filtering supports filter-lang=cql2-text")
			return
		}
		if rs.database == nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, "CQL filtering requires a database-backed Records catalog")
			return
		}
		matchedIDs, err := cqlMatchedRecordIDs(r.Context(), rs.database, rs.databaseDriver, collection, filter)
		if err != nil {
			engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
			return
		}
		matchedRecords = retainRecordIDs(matchedRecords, matchedIDs)
	}
	matchedRecords, err = sortRecords(matchedRecords, collection, r.URL.Query())
	if err != nil {
		engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
		return
	}
	records, nextOffset, err := pageRecords(matchedRecords, r.URL.Query())
	if err != nil {
		engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
		return
	}

	features := make([]map[string]any, 0, len(records))
	for _, record := range records {
		features = append(features, rs.item(collectionID, record))
	}
	links := rs.collectionLinks(r, collectionID)
	if nextOffset != nil {
		nextQuery := r.URL.Query()
		nextQuery.Set("offset", strconv.Itoa(*nextOffset))
		nextHref := strings.TrimRight(rs.engine.Config.BaseURL.String(), "/") +
			"/collections/" + url.PathEscape(collectionID) + "/items?" + nextQuery.Encode()
		links = append(links, config.RecordsLink{Rel: "next", Type: mediaTypeGeoJSON, Href: nextHref})
	}
	response := map[string]any{
		"type":           "FeatureCollection",
		"features":       features,
		"numberReturned": len(features),
		"numberMatched":  len(matchedRecords),
		"timeStamp":      time.Now().UTC().Format(time.RFC3339Nano),
		"links":          links,
	}
	if rs.renderItemsPage(w, r, collection, records, len(matchedRecords)) {
		return
	}
	rs.serveJSON(w, r, response)
}

func (rs *Records) Item() http.HandlerFunc {
	return rs.handleItem
}

func (rs *Records) handleItem(w http.ResponseWriter, r *http.Request) {
	if !rs.validateRequest(w, r) {
		return
	}
	collectionID := chi.URLParam(r, "collectionId")
	recordID := chi.URLParam(r, "recordId")
	if recordID == "" {
		recordID = chi.URLParam(r, "featureId")
	}
	collection, ok := rs.collections[collectionID]
	if !ok {
		engine.RenderProblem(engine.ProblemNotFound, w)
		return
	}

	for _, record := range collection.Records {
		if record.ID == recordID {
			if rs.renderItemPage(w, r, collectionID, record) {
				return
			}
			rs.serveJSON(w, r, rs.item(collectionID, record))
			return
		}
	}
	engine.RenderProblem(engine.ProblemNotFound, w)
}

func (rs *Records) Sortables() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !rs.validateRequest(w, r) {
			return
		}
		collectionID := chi.URLParam(r, "collectionId")
		collection, ok := rs.collections[collectionID]
		if !ok {
			engine.RenderProblem(engine.ProblemNotFound, w)
			return
		}
		properties := make(map[string]any)
		for _, name := range sortableNames(collection) {
			properties[name] = map[string]any{"type": sortableType(collection, name)}
		}
		baseURL := strings.TrimRight(rs.engine.Config.BaseURL.String(), "/")
		response := map[string]any{
			"$schema":              "https://json-schema.org/draft/2020-12/schema",
			"$id":                  baseURL + "/collections/" + url.PathEscape(collectionID) + "/sortables",
			"type":                 "object",
			"properties":           properties,
			"additionalProperties": false,
		}
		rs.serveJSONAs(w, r, response, "application/schema+json")
	}
}

func (rs *Records) item(collectionID string, record *config.Record) map[string]any {
	baseURL := strings.TrimRight(rs.engine.Config.BaseURL.String(), "/")
	collectionPath := "/collections/" + url.PathEscape(collectionID)
	itemPath := collectionPath + "/items/" + url.PathEscape(record.ID)
	links := []config.RecordsLink{
		{Rel: "self", Type: mediaTypeGeoJSON, Href: baseURL + itemPath + "?f=json"},
		{Rel: "alternate", Type: mediaTypeHTML, Href: baseURL + itemPath + "?f=html"},
		{Rel: "alternate", Type: mediaTypeMarkdown, Href: baseURL + itemPath + "?f=md"},
		{Rel: "collection", Type: mediaTypeJSON, Href: baseURL + collectionPath + "?f=json"},
	}
	links = append(links, record.Links...)
	assets := make(map[string]map[string]any, len(record.Assets))
	for id, configuredAsset := range record.Assets {
		asset := make(map[string]any, 5+len(configuredAsset.Fields))
		for key, value := range configuredAsset.Fields {
			asset[key] = value
		}
		asset["href"] = configuredAsset.Href
		if configuredAsset.Title != "" {
			asset["title"] = configuredAsset.Title
		}
		if configuredAsset.Description != "" {
			asset["description"] = configuredAsset.Description
		}
		if configuredAsset.Type != "" {
			asset["type"] = configuredAsset.Type
		}
		if len(configuredAsset.Roles) > 0 {
			asset["roles"] = configuredAsset.Roles
		}
		assets[id] = asset
	}
	item := map[string]any{
		"id":           record.ID,
		"type":         "Feature",
		"geometry":     record.Geometry,
		"properties":   record.Properties,
		"stac_version": record.STACVersion,
		"assets":       assets,
		"links":        links,
	}
	stacExtensions := stacExtensionURIs(record)
	if len(stacExtensions) > 0 {
		item["stac_extensions"] = stacExtensions
	}
	if len(record.Bbox) > 0 {
		item["bbox"] = record.Bbox
	}
	if record.Time != nil {
		item["time"] = record.Time
	}
	conformsTo := append([]string(nil), record.ConformsTo...)
	appendUnique(&conformsTo, "http://www.opengis.net/spec/ogcapi-records-1/1.0/conf/record-core")
	if len(stacExtensions) > 0 {
		appendUnique(&conformsTo, "https://www.opengis.net/spec/ogcapi-records-5/1.0/conf/stac-extensions")
	}
	item["conformsTo"] = conformsTo
	if record.Collection != "" {
		item["collection"] = record.Collection
	}
	return item
}

func stacExtensionURIs(record *config.Record) []string {
	result := append([]string(nil), record.STACExtensions...)
	knownExtensions := []struct {
		prefix string
		name   string
		uri    string
	}{
		{prefix: "file", name: "file", uri: "https://stac-extensions.github.io/file/v2.1.0/schema.json"},
		{prefix: "table", name: "table", uri: "https://stac-extensions.github.io/table/v1.3.0/schema.json"},
		{prefix: "raster", name: "raster", uri: "https://stac-extensions.github.io/raster/v2.0.0/schema.json"},
		{prefix: "proj", name: "projection", uri: "https://stac-extensions.github.io/projection/v2.1.0/schema.json"},
	}
	usedPrefixes := make(map[string]bool)
	collect := func(fields map[string]config.JSONValue) {
		for name := range fields {
			prefix, _, hasPrefix := strings.Cut(name, ":")
			if hasPrefix {
				usedPrefixes[prefix] = true
			}
		}
	}
	collect(record.Properties)
	for _, asset := range record.Assets {
		collect(asset.Fields)
	}
	for _, extension := range knownExtensions {
		if !usedPrefixes[extension.prefix] {
			continue
		}
		found := false
		for _, uri := range result {
			if strings.Contains(uri, "/"+extension.name+"/") {
				found = true
				break
			}
		}
		if !found {
			result = append(result, extension.uri)
		}
	}
	return result
}

func appendUnique(values *[]string, value string) {
	for _, existing := range *values {
		if existing == value {
			return
		}
	}
	*values = append(*values, value)
}

func (rs *Records) collectionLinks(r *http.Request, collectionID string) []config.RecordsLink {
	baseURL := strings.TrimRight(rs.engine.Config.BaseURL.String(), "/")
	itemsURL := baseURL + "/collections/" + url.PathEscape(collectionID) + "/items"
	selfQuery := r.URL.Query()
	if selfQuery.Get("f") == "" {
		selfQuery.Set("f", "json")
	}
	alternateQuery := r.URL.Query()
	alternateQuery.Set("f", "html")
	return []config.RecordsLink{
		{Rel: "self", Type: mediaTypeGeoJSON, Href: itemsURL + "?" + selfQuery.Encode()},
		{Rel: "alternate", Type: mediaTypeHTML, Href: itemsURL + "?" + alternateQuery.Encode()},
		{Rel: "alternate", Type: mediaTypeMarkdown, Href: itemsURL + "?f=md"},
		{Rel: "collection", Type: mediaTypeJSON, Href: baseURL + "/collections/" + url.PathEscape(collectionID) + "?f=json"},
	}
}

func (rs *Records) renderItemsPage(w http.ResponseWriter, r *http.Request, collection config.RecordsCollection,
	records config.Records, numberMatched int) bool {
	format := rs.negotiatedFormat(r)
	key, ok := map[string]engine.TemplateKey{engine.FormatHTML: itemsHTMLKey, engine.FormatMarkdown: itemsMDKey}[format]
	if !ok {
		return false
	}
	items := make([]recordPageItem, 0, len(records))
	for _, record := range records {
		items = append(items, rs.pageItem(collection.ID, record))
	}
	page := recordsPage{CollectionID: collection.ID, Items: items, NumberMatched: numberMatched}
	rs.engine.RenderAndServe(w, r, engine.ExpandTemplateKey(key, rs.engine.CN.NegotiateLanguage(w, r)), page,
		[]engine.Breadcrumb{{Name: "Collections", Path: "collections"}, {Name: collection.ID, Path: "collections/" + collection.ID}},
		[]engine.OutputFormat{{Key: engine.FormatJSON, Name: "GeoJSON"}, {Key: engine.FormatHTML, Name: "HTML"}, {Key: engine.FormatMarkdown, Name: "Markdown"}})
	return true
}

func (rs *Records) renderItemPage(w http.ResponseWriter, r *http.Request, collectionID string, record *config.Record) bool {
	format := rs.negotiatedFormat(r)
	key, ok := map[string]engine.TemplateKey{engine.FormatHTML: itemHTMLKey, engine.FormatMarkdown: itemMDKey}[format]
	if !ok {
		return false
	}
	page := rs.pageItem(collectionID, record)
	rs.engine.RenderAndServe(w, r, engine.ExpandTemplateKey(key, rs.engine.CN.NegotiateLanguage(w, r)), page,
		[]engine.Breadcrumb{{Name: "Collections", Path: "collections"}, {Name: collectionID, Path: "collections/" + collectionID}, {Name: page.Title, Path: "collections/" + collectionID + "/items/" + record.ID}},
		[]engine.OutputFormat{{Key: engine.FormatGeoJSON, Name: "GeoJSON"}, {Key: engine.FormatHTML, Name: "HTML"}, {Key: engine.FormatMarkdown, Name: "Markdown"}})
	return true
}

func (rs *Records) negotiatedFormat(r *http.Request) string {
	if rs.engine.CN == nil {
		return engine.FormatJSON
	}
	return rs.engine.CN.NegotiateFormat(r)
}

func (rs *Records) pageItem(collectionID string, record *config.Record) recordPageItem {
	title := propertyString(record.Properties["title"])
	if title == "" {
		title = record.ID
	}
	baseURL := strings.TrimRight(rs.engine.Config.BaseURL.String(), "/")
	properties := make([]recordPageProperty, 0, len(record.Properties))
	keys := make([]string, 0, len(record.Properties))
	for key := range record.Properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := record.Properties[key]
		formatted := string(value)
		if decoded, err := value.Any(); err == nil {
			if text, ok := decoded.(string); ok {
				formatted = text
			}
		}
		properties = append(properties, recordPageProperty{Name: key, Value: formatted})
	}
	assets := make([]recordPageAsset, 0, len(record.Assets))
	assetKeys := make([]string, 0, len(record.Assets))
	for key := range record.Assets {
		assetKeys = append(assetKeys, key)
	}
	sort.Strings(assetKeys)
	for _, key := range assetKeys {
		asset := record.Assets[key]
		assets = append(assets, recordPageAsset{Key: key, Href: asset.Href, Title: asset.Title, MediaType: asset.Type})
	}
	return recordPageItem{
		ID:          record.ID,
		Title:       title,
		Description: propertyString(record.Properties["description"]),
		Href:        baseURL + "/collections/" + url.PathEscape(collectionID) + "/items/" + url.PathEscape(record.ID),
		Properties:  properties,
		Assets:      assets,
	}
}

type recordsPage struct {
	CollectionID  string
	Items         []recordPageItem
	NumberMatched int
}

type recordPageItem struct {
	ID          string
	Title       string
	Description string
	Href        string
	Properties  []recordPageProperty
	Assets      []recordPageAsset
}

type recordPageProperty struct {
	Name  string
	Value string
}

type recordPageAsset struct {
	Key       string
	Href      string
	Title     string
	MediaType string
}

func (rs *Records) serveJSON(w http.ResponseWriter, r *http.Request, value any) {
	rs.serveJSONAs(w, r, value, mediaTypeGeoJSON)
}

func (rs *Records) serveJSONAs(w http.ResponseWriter, r *http.Request, value any, contentType string) {
	rs.engine.Serve(w, r,
		engine.ServeValidation(false, rs.engine.OpenAPI != nil),
		engine.ServeContentType(contentType),
		engine.ServeJSON(value),
	)
}

func (rs *Records) validateRequest(w http.ResponseWriter, r *http.Request) bool {
	if rs.engine.OpenAPI == nil {
		return true
	}
	if err := rs.engine.OpenAPI.ValidateRequest(r); err != nil {
		engine.RenderProblem(engine.ProblemBadRequest, w, err.Error())
		return false
	}
	return true
}

func selectRecords(records config.Records, query url.Values) (config.Records, error) {
	selected := make(config.Records, 0, len(records))
	var requestedBbox []float64
	var requestedTime *timeRange
	if query.Has("bbox") {
		var err error
		requestedBbox, err = parseBbox(query.Get("bbox"))
		if err != nil {
			return nil, err
		}
	}
	if query.Has("datetime") {
		parsed, err := parseDateTimeRange(query.Get("datetime"))
		if err != nil {
			return nil, err
		}
		requestedTime = &parsed
	}
	ids, hasIDs := commaSeparated(query, "ids")
	types, hasTypes := commaSeparated(query, "type")
	externalIDs, hasExternalIDs := commaSeparated(query, "externalIds")
	searchTerms, hasSearchTerms := commaSeparated(query, "q")
	queryableNames := make([]string, 0, len(query))
	for name := range query {
		if !isReservedParameter(name) {
			queryableNames = append(queryableNames, name)
		}
	}
	for _, record := range records {
		if query.Has("bbox") && !matchesBbox(record, requestedBbox) {
			continue
		}
		if requestedTime != nil {
			matches, err := matchesDateTime(record.Time, *requestedTime)
			if err != nil {
				return nil, err
			}
			if !matches {
				continue
			}
		}
		if hasIDs && !contains(ids, record.ID) {
			continue
		}
		if hasTypes && !contains(types, propertyString(record.Properties["type"])) {
			continue
		}
		if hasExternalIDs && !matchesExternalIDs(record.Properties["externalIds"], externalIDs) {
			continue
		}
		if hasSearchTerms && !matchesSearchTerms(record.Properties, searchTerms) {
			continue
		}
		matchesQueryables := true
		for _, name := range queryableNames {
			value := query.Get(name)
			var propertyValue any = record.Properties[name]
			if name == "id" {
				propertyValue = record.ID
			}
			if propertyString(propertyValue) != value {
				matchesQueryables = false
				break
			}
		}
		if !matchesQueryables {
			continue
		}
		selected = append(selected, record)
	}
	return selected, nil
}

func commaSeparated(query url.Values, name string) ([]string, bool) {
	values, exists := query[name]
	if !exists {
		return nil, false
	}
	var result []string
	for _, value := range values {
		result = append(result, strings.Split(value, ",")...)
	}
	return result, true
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func matchesExternalIDs(raw any, queries []string) bool {
	if encoded, ok := raw.(config.JSONValue); ok {
		decoded, err := encoded.Any()
		if err != nil {
			return false
		}
		raw = decoded
	}
	values, ok := raw.([]any)
	if !ok {
		return false
	}
	for _, rawIdentifier := range values {
		identifier, ok := rawIdentifier.(map[string]any)
		if !ok {
			continue
		}
		value := propertyString(identifier["value"])
		scheme := propertyString(identifier["scheme"])
		for _, query := range queries {
			queryScheme, queryValue, hasScheme := strings.Cut(query, ":")
			if queryValue == "" || (!hasScheme && value == query) || (hasScheme && scheme == queryScheme && value == queryValue) {
				return true
			}
		}
	}
	return false
}

func matchesSearchTerms(properties map[string]config.JSONValue, terms []string) bool {
	textFields := []string{"title", "description", "keywords"}
	for _, term := range terms {
		words := strings.Fields(term)
		if len(words) == 0 {
			continue
		}
		pattern := "(?i)" + regexp.QuoteMeta(words[0])
		for _, word := range words[1:] {
			pattern += `\s+` + regexp.QuoteMeta(word)
		}
		matcher, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		for _, field := range textFields {
			for _, value := range textValues(properties[field]) {
				if matcher.MatchString(value) {
					return true
				}
			}
		}
	}
	return false
}

func textValues(value any) []string {
	if raw, ok := value.(config.JSONValue); ok {
		decoded, err := raw.Any()
		if err == nil {
			return textValues(decoded)
		}
	}
	switch value := value.(type) {
	case string:
		return []string{value}
	case []string:
		return value
	case []any:
		result := make([]string, 0, len(value))
		for _, item := range value {
			if text, ok := item.(string); ok {
				result = append(result, text)
			}
		}
		return result
	default:
		return nil
	}
}

func propertyString(value any) string {
	if raw, ok := value.(config.JSONValue); ok {
		return raw.StringValue()
	}
	if text, ok := value.(string); ok {
		return text
	}
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func isReservedParameter(name string) bool {
	switch name {
	case "f", "bbox", "datetime", "limit", "offset", "q", "type", "ids", "externalIds", "sortby", "language", "profile", "filter", "filter-lang", "filter-crs":
		return true
	default:
		return false
	}
}

type timeRange struct {
	start *time.Time
	end   *time.Time
}

func parseBbox(raw string) ([]float64, error) {
	parts := strings.Split(raw, ",")
	if len(parts) != 4 && len(parts) != 6 {
		return nil, fmt.Errorf("bbox must contain four or six coordinates")
	}
	bbox := make([]float64, len(parts))
	for index, part := range parts {
		value, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid bbox coordinate %q", part)
		}
		bbox[index] = value
	}
	maxXIndex, maxYIndex := 2, 3
	if len(bbox) == 6 {
		maxXIndex, maxYIndex = 3, 4
		if bbox[2] > bbox[5] {
			return nil, fmt.Errorf("bbox vertical minimum exceeds maximum")
		}
	}
	if bbox[1] > bbox[maxYIndex] {
		return nil, fmt.Errorf("bbox south coordinate exceeds north coordinate")
	}
	if bbox[1] < -90 || bbox[1] > 90 || bbox[maxYIndex] < -90 || bbox[maxYIndex] > 90 {
		return nil, fmt.Errorf("bbox latitude must be between -90 and 90")
	}
	_ = maxXIndex
	return bbox, nil
}

func matchesBbox(record *config.Record, requested []float64) bool {
	if !recordHasGeometry(record.Geometry) {
		return true
	}
	if len(record.Bbox) != 4 && len(record.Bbox) != 6 {
		return false
	}
	if !longitudeIntersects(record.Bbox[0], record.Bbox[bboxMaxXIndex(record.Bbox)], requested[0], requested[bboxMaxXIndex(requested)]) {
		return false
	}
	if record.Bbox[1] > requested[bboxMaxYIndex(requested)] || record.Bbox[bboxMaxYIndex(record.Bbox)] < requested[1] {
		return false
	}
	if len(record.Bbox) == 6 && len(requested) == 6 {
		if record.Bbox[2] > requested[5] || record.Bbox[5] < requested[2] {
			return false
		}
	}
	return true
}

func bboxMaxXIndex(bbox []float64) int {
	if len(bbox) == 6 {
		return 3
	}
	return 2
}

func bboxMaxYIndex(bbox []float64) int {
	if len(bbox) == 6 {
		return 4
	}
	return 3
}

func longitudeIntersects(firstWest, firstEast, secondWest, secondEast float64) bool {
	for _, first := range longitudeIntervals(firstWest, firstEast) {
		for _, second := range longitudeIntervals(secondWest, secondEast) {
			if first[0] <= second[1] && first[1] >= second[0] {
				return true
			}
		}
	}
	return false
}

func longitudeIntervals(west, east float64) [][2]float64 {
	if west <= east {
		return [][2]float64{{west, east}}
	}
	return [][2]float64{{west, 180}, {-180, east}}
}

func parseDateTimeRange(raw string) (timeRange, error) {
	if !strings.Contains(raw, "/") {
		instant, err := parseTimeValue(raw)
		if err != nil {
			return timeRange{}, fmt.Errorf("invalid datetime %q", raw)
		}
		return timeRange{start: &instant, end: &instant}, nil
	}
	parts := strings.Split(raw, "/")
	if len(parts) != 2 {
		return timeRange{}, fmt.Errorf("datetime interval must contain one slash")
	}
	start, err := parseTimeBound(parts[0])
	if err != nil {
		return timeRange{}, fmt.Errorf("invalid datetime interval start %q", parts[0])
	}
	end, err := parseTimeBound(parts[1])
	if err != nil {
		return timeRange{}, fmt.Errorf("invalid datetime interval end %q", parts[1])
	}
	if start != nil && end != nil && start.After(*end) {
		return timeRange{}, fmt.Errorf("datetime interval start is after its end")
	}
	return timeRange{start: start, end: end}, nil
}

func matchesDateTime(recordTime *config.RecordTime, requested timeRange) (bool, error) {
	recordRange, hasTemporal, err := recordTimeRange(recordTime)
	if err != nil {
		return false, err
	}
	if !hasTemporal {
		return true, nil
	}
	if requested.start != nil && recordRange.end != nil && recordRange.end.Before(*requested.start) {
		return false, nil
	}
	if requested.end != nil && recordRange.start != nil && recordRange.start.After(*requested.end) {
		return false, nil
	}
	return true, nil
}

func recordTimeRange(recordTime *config.RecordTime) (timeRange, bool, error) {
	if recordTime == nil {
		return timeRange{}, false, nil
	}
	if recordTime.Timestamp != "" {
		instant, err := parseTimeValue(recordTime.Timestamp)
		return timeRange{start: &instant, end: &instant}, true, err
	}
	if recordTime.Date != "" {
		day, err := time.Parse("2006-01-02", recordTime.Date)
		if err != nil {
			return timeRange{}, false, fmt.Errorf("invalid record date %q", recordTime.Date)
		}
		start := day.UTC()
		end := start.Add(24*time.Hour - time.Nanosecond)
		return timeRange{start: &start, end: &end}, true, nil
	}
	if len(recordTime.Interval) == 2 {
		start, err := parseTimeBound(recordTime.Interval[0])
		if err != nil {
			return timeRange{}, false, fmt.Errorf("invalid record interval start %q", recordTime.Interval[0])
		}
		end, err := parseTimeBound(recordTime.Interval[1])
		if err != nil {
			return timeRange{}, false, fmt.Errorf("invalid record interval end %q", recordTime.Interval[1])
		}
		return timeRange{start: start, end: end}, true, nil
	}
	return timeRange{}, false, nil
}

func parseTimeBound(value string) (*time.Time, error) {
	if value == "" || value == ".." {
		return nil, nil
	}
	timestamp, err := parseTimeValue(value)
	if err != nil {
		return nil, err
	}
	return &timestamp, nil
}

func parseTimeValue(value string) (time.Time, error) {
	if day, err := time.Parse("2006-01-02", value); err == nil {
		return day.UTC(), nil
	}
	timestamp, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, err
	}
	return timestamp.UTC(), nil
}

func pageRecords(records config.Records, query url.Values) (config.Records, *int, error) {
	limit := defaultLimit
	if rawLimit := query.Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit < 1 {
			return nil, nil, fmt.Errorf("invalid limit %q", rawLimit)
		}
		limit = parsedLimit
	}
	if limit > maximumLimit {
		limit = maximumLimit
	}
	offset := 0
	if rawOffset := query.Get("offset"); rawOffset != "" {
		parsedOffset, err := strconv.Atoi(rawOffset)
		if err != nil || parsedOffset < 0 {
			return nil, nil, fmt.Errorf("invalid offset %q", rawOffset)
		}
		offset = parsedOffset
	}
	if offset >= len(records) {
		return config.Records{}, nil, nil
	}
	end := offset + limit
	if end >= len(records) {
		return records[offset:], nil, nil
	}
	nextOffset := end
	return records[offset:end], &nextOffset, nil
}

func retainRecordIDs(records config.Records, matchedIDs map[string]bool) config.Records {
	selected := make(config.Records, 0, len(records))
	for _, record := range records {
		if matchedIDs[record.ID] {
			selected = append(selected, record)
		}
	}
	return selected
}

func (rs *Records) Queryables() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !rs.validateRequest(w, r) {
			return
		}
		collectionID := chi.URLParam(r, "collectionId")
		collection, ok := rs.collections[collectionID]
		if !ok || !collection.SupportsCQL() {
			engine.RenderProblem(engine.ProblemNotFound, w)
			return
		}
		properties := map[string]any{
			"id": map[string]any{"type": "string"},
		}
		for _, queryable := range collection.Filters.Properties {
			properties[queryable.Name] = map[string]any{"type": sortableType(collection, queryable.Name)}
		}
		baseURL := strings.TrimRight(rs.engine.Config.BaseURL.String(), "/")
		response := map[string]any{
			"$schema":              "https://json-schema.org/draft/2020-12/schema",
			"$id":                  baseURL + "/collections/" + url.PathEscape(collectionID) + "/queryables",
			"type":                 "object",
			"properties":           properties,
			"additionalProperties": false,
		}
		rs.serveJSONAs(w, r, response, "application/schema+json")
	}
}

func recordHasGeometry(value config.JSONValue) bool {
	if len(value) == 0 {
		return false
	}
	geometry, err := value.Any()
	return err != nil || geometry != nil
}
