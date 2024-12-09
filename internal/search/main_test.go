package search

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/PDOK/gomagpie/config"
	"github.com/PDOK/gomagpie/internal/engine"
	"github.com/PDOK/gomagpie/internal/etl"
	"github.com/docker/go-connections/nat"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const testSearchIndex = "search_index"

func init() {
	// change working dir to root
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestSuggest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	ctx := context.Background()

	// given available postgres
	dbPort, postgisContainer, err := setupPostgis(ctx, t)
	if err != nil {
		t.Error(err)
	}
	defer terminateContainer(ctx, t, postgisContainer)

	dbConn := fmt.Sprintf("postgres://postgres:postgres@127.0.0.1:%d/%s?sslmode=disable", dbPort.Int(), "test_db")

	// given available engine
	eng, err := engine.NewEngine("internal/etl/testdata/config.yaml", false, false)
	assert.NoError(t, err)

	// given search endpoint
	searchEndpoint := NewSearch(eng, dbConn, testSearchIndex)

	// given empty search index
	err = etl.CreateSearchIndex(dbConn, testSearchIndex)
	assert.NoError(t, err)

	// given imported geopackage
	conf, err := config.NewConfig("internal/etl/testdata/config.yaml")
	assert.NoError(t, err)
	collection := config.CollectionByID(conf, "addresses")
	table := config.FeatureTable{Name: "addresses", FID: "fid", Geom: "geom"}
	err = etl.ImportFile(*collection, testSearchIndex, "internal/etl/testdata/addresses-crs84.gpkg", table, 1000, dbConn)
	assert.NoError(t, err)

	// run test cases
	type fields struct {
		url string
	}
	type want struct {
		body       string
		statusCode int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "Suggest: Oudeschild",
			fields: fields{
				url: "http://localhost:8080/search/suggest?q=\"Oudeschild\"&limit=50",
			},
			want: want{
				body: `
{"type":"FeatureCollection","timeStamp":"2000-01-01T00:00:00Z","features":[{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"Barentszstraat - Oudeschild","highlight":"Barentszstraat, 1792AD <b>Oudeschild</b>","href":"<todo>","score":0.14426951110363007},"geometry":{"type":"Polygon","coordinates":[[[4.748384044242354,52.93901709012591],[4.948384044242354,52.93901709012591],[4.948384044242354,53.13901709012591],[4.748384044242354,53.13901709012591],[4.748384044242354,52.93901709012591]]]},"id":"548"},{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"Bolwerk - Oudeschild","highlight":"Bolwerk, 1792AS <b>Oudeschild</b>","href":"<todo>","score":0.14426951110363007},"geometry":{"type":"Polygon","coordinates":[[[4.75002232386939,52.93847294238573],[4.95002232386939,52.93847294238573],[4.95002232386939,53.13847294238573],[4.75002232386939,53.13847294238573],[4.75002232386939,52.93847294238573]]]},"id":"1050"},{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"Commandeurssingel - Oudeschild","highlight":"Commandeurssingel, 1792AV <b>Oudeschild</b>","href":"<todo>","score":0.14426951110363007},"geometry":{"type":"Polygon","coordinates":[[[4.7451477245429015,52.93967814281323],[4.945147724542901,52.93967814281323],[4.945147724542901,53.13967814281323],[4.7451477245429015,53.13967814281323],[4.7451477245429015,52.93967814281323]]]},"id":"2725"},{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"De Houtmanstraat - Oudeschild","highlight":"De Houtmanstraat, 1792BC <b>Oudeschild</b>","href":"<todo>","score":0.12426698952913284},"geometry":{"type":"Polygon","coordinates":[[[4.748360166368449,52.93815392755542],[4.948360166368448,52.93815392755542],[4.948360166368448,53.13815392755542],[4.748360166368449,53.13815392755542],[4.748360166368449,52.93815392755542]]]},"id":"2921"},{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"De Ruyterstraat - Oudeschild","highlight":"De Ruyterstraat, 1792AP <b>Oudeschild</b>","href":"<todo>","score":0.12426698952913284},"geometry":{"type":"Polygon","coordinates":[[[4.747714279539418,52.93617309495475],[4.947714279539417,52.93617309495475],[4.947714279539417,53.136173094954756],[4.747714279539418,53.136173094954756],[4.747714279539418,52.93617309495475]]]},"id":"3049"},{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"De Wittstraat - Oudeschild","highlight":"De Wittstraat, 1792BP <b>Oudeschild</b>","href":"<todo>","score":0.12426698952913284},"geometry":{"type":"Polygon","coordinates":[[[4.745616492688666,52.93705261983951],[4.945616492688665,52.93705261983951],[4.945616492688665,53.137052619839515],[4.745616492688666,53.137052619839515],[4.745616492688666,52.93705261983951]]]},"id":"3041"}],"numberReturned":6}
`,
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Suggest: Den ",
			fields: fields{
				url: "http://localhost:8080/search/suggest?q=\"Den\"&limit=50",
			},
			want: want{
				body: `
{"type":"FeatureCollection","timeStamp":"2000-01-01T00:00:00Z","features":[{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"Abbewaal - Den Burg","highlight":"Abbewaal, 1791WZ <b>Den</b> Burg","href":"<todo>","score":0.11162212491035461},"geometry":{"type":"Polygon","coordinates":[[[4.701721422439945,52.9619223105808],[4.901721422439945,52.9619223105808],[4.901721422439945,53.161922310580806],[4.701721422439945,53.161922310580806],[4.701721422439945,52.9619223105808]]]},"id":"99"},{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"Achterom - Den Burg","highlight":"Achterom, 1791AN <b>Den</b> Burg","href":"<todo>","score":0.11162212491035461},"geometry":{"type":"Polygon","coordinates":[[[4.699813158490893,52.95463219709524],[4.899813158490892,52.95463219709524],[4.899813158490892,53.154632197095246],[4.699813158490893,53.154632197095246],[4.699813158490893,52.95463219709524]]]},"id":"114"},{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"Akenbuurt - Den Burg","highlight":"Akenbuurt, 1791PJ <b>Den</b> Burg","href":"<todo>","score":0.11162212491035461},"geometry":{"type":"Polygon","coordinates":[[[4.680059099895046,52.95346592050607],[4.880059099895045,52.95346592050607],[4.880059099895045,53.15346592050607],[4.680059099895046,53.15346592050607],[4.680059099895046,52.95346592050607]]]},"id":"46"},{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"Amaliaweg - Den Hoorn","highlight":"Amaliaweg, 1797SW <b>Den</b> Hoorn","href":"<todo>","score":0.11162212491035461},"geometry":{"type":"Polygon","coordinates":[[[4.68911630577304,52.92449928128154],[4.889116305773039,52.92449928128154],[4.889116305773039,53.124499281281544],[4.68911630577304,53.124499281281544],[4.68911630577304,52.92449928128154]]]},"id":"50"},{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"Bakkenweg - Den Hoorn","highlight":"Bakkenweg, 1797RJ <b>Den</b> Hoorn","href":"<todo>","score":0.11162212491035461},"geometry":{"type":"Polygon","coordinates":[[[4.6548723261037095,52.94811743920973],[4.854872326103709,52.94811743920973],[4.854872326103709,53.148117439209734],[4.6548723261037095,53.148117439209734],[4.6548723261037095,52.94811743920973]]]},"id":"520"},{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"Beatrixlaan - Den Burg","highlight":"Beatrixlaan, 1791GE <b>Den</b> Burg","href":"<todo>","score":0.11162212491035461},"geometry":{"type":"Polygon","coordinates":[[[4.690892824472019,52.95558352001795],[4.890892824472019,52.95558352001795],[4.890892824472019,53.155583520017956],[4.690892824472019,53.155583520017956],[4.690892824472019,52.95558352001795]]]},"id":"591"},{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"Ada van Hollandstraat - Den Burg","highlight":"Ada van Hollandstraat, 1791DH <b>Den</b> Burg","href":"<todo>","score":0.09617967158555984},"geometry":{"type":"Polygon","coordinates":[[[4.696235388824104,52.95196001510249],[4.8962353888241035,52.95196001510249],[4.8962353888241035,53.151960015102496],[4.696235388824104,53.151960015102496],[4.696235388824104,52.95196001510249]]]},"id":"26"},{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"Anne Frankstraat - Den Burg","highlight":"Anne Frankstraat, 1791DT <b>Den</b> Burg","href":"<todo>","score":0.09617967158555984},"geometry":{"type":"Polygon","coordinates":[[[4.692873779103581,52.950932925919574],[4.892873779103581,52.950932925919574],[4.892873779103581,53.15093292591958],[4.692873779103581,53.15093292591958],[4.692873779103581,52.950932925919574]]]},"id":"474"}],"numberReturned":8}
`,
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Suggest: Den. With deepCopy params",
			fields: fields{
				url: "http://localhost:8080/search/suggest?q=\"Den\"&weg[version]=2&weg[relevance]=0.8&adres[version]=1&adres[relevance]=1&limit=10&f=json&crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992",
			},
			want: want{
				body: `
{"type":"FeatureCollection","timeStamp":"2000-01-01T00:00:00Z","features":[{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"Abbewaal - Den Burg","highlight":"Abbewaal, 1791WZ <b>Den</b> Burg","href":"<todo>","score":0.11162212491035461},"geometry":{"type":"Polygon","coordinates":[[[4.701721422439945,52.9619223105808],[4.901721422439945,52.9619223105808],[4.901721422439945,53.161922310580806],[4.701721422439945,53.161922310580806],[4.701721422439945,52.9619223105808]]]},"id":"99"},{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"Achterom - Den Burg","highlight":"Achterom, 1791AN <b>Den</b> Burg","href":"<todo>","score":0.11162212491035461},"geometry":{"type":"Polygon","coordinates":[[[4.699813158490893,52.95463219709524],[4.899813158490892,52.95463219709524],[4.899813158490892,53.154632197095246],[4.699813158490893,53.154632197095246],[4.699813158490893,52.95463219709524]]]},"id":"114"},{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"Akenbuurt - Den Burg","highlight":"Akenbuurt, 1791PJ <b>Den</b> Burg","href":"<todo>","score":0.11162212491035461},"geometry":{"type":"Polygon","coordinates":[[[4.680059099895046,52.95346592050607],[4.880059099895045,52.95346592050607],[4.880059099895045,53.15346592050607],[4.680059099895046,53.15346592050607],[4.680059099895046,52.95346592050607]]]},"id":"46"},{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"Amaliaweg - Den Hoorn","highlight":"Amaliaweg, 1797SW <b>Den</b> Hoorn","href":"<todo>","score":0.11162212491035461},"geometry":{"type":"Polygon","coordinates":[[[4.68911630577304,52.92449928128154],[4.889116305773039,52.92449928128154],[4.889116305773039,53.124499281281544],[4.68911630577304,53.124499281281544],[4.68911630577304,52.92449928128154]]]},"id":"50"},{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"Bakkenweg - Den Hoorn","highlight":"Bakkenweg, 1797RJ <b>Den</b> Hoorn","href":"<todo>","score":0.11162212491035461},"geometry":{"type":"Polygon","coordinates":[[[4.6548723261037095,52.94811743920973],[4.854872326103709,52.94811743920973],[4.854872326103709,53.148117439209734],[4.6548723261037095,53.148117439209734],[4.6548723261037095,52.94811743920973]]]},"id":"520"},{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"Beatrixlaan - Den Burg","highlight":"Beatrixlaan, 1791GE <b>Den</b> Burg","href":"<todo>","score":0.11162212491035461},"geometry":{"type":"Polygon","coordinates":[[[4.690892824472019,52.95558352001795],[4.890892824472019,52.95558352001795],[4.890892824472019,53.155583520017956],[4.690892824472019,53.155583520017956],[4.690892824472019,52.95558352001795]]]},"id":"591"},{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"Ada van Hollandstraat - Den Burg","highlight":"Ada van Hollandstraat, 1791DH <b>Den</b> Burg","href":"<todo>","score":0.09617967158555984},"geometry":{"type":"Polygon","coordinates":[[[4.696235388824104,52.95196001510249],[4.8962353888241035,52.95196001510249],[4.8962353888241035,53.151960015102496],[4.696235388824104,53.151960015102496],[4.696235388824104,52.95196001510249]]]},"id":"26"},{"type":"Feature","properties":{"collectionId":"addresses","collectionVersion":"1","displayName":"Anne Frankstraat - Den Burg","highlight":"Anne Frankstraat, 1791DT <b>Den</b> Burg","href":"<todo>","score":0.09617967158555984},"geometry":{"type":"Polygon","coordinates":[[[4.692873779103581,52.950932925919574],[4.892873779103581,52.950932925919574],[4.892873779103581,53.15093292591958],[4.692873779103581,53.15093292591958],[4.692873779103581,52.950932925919574]]]},"id":"474"}],"numberReturned":8}
`,
				statusCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock time
			now = func() time.Time { return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC) }

			// given available server
			rr, ts := createMockServer()
			defer ts.Close()

			// when
			handler := searchEndpoint.Suggest()
			req, err := createRequest(tt.fields.url)
			assert.NoError(t, err)
			handler.ServeHTTP(rr, req)

			// then
			assert.Equal(t, tt.want.statusCode, rr.Code)

			log.Printf("============ ACTUAL:\n %s", rr.Body.String())
			assert.JSONEq(t, tt.want.body, rr.Body.String())
		})
	}
}

func setupPostgis(ctx context.Context, t *testing.T) (nat.Port, testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image: "docker.io/postgis/postgis:16-3.5-alpine",
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "postgres",
		},
		ExposedPorts: []string{"5432/tcp"},
		Cmd:          []string{"postgres", "-c", "fsync=off"},
		WaitingFor:   wait.ForLog("PostgreSQL init process complete; ready for start up."),
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      "tests/testdata/sql/init-db.sql",
				ContainerFilePath: "/docker-entrypoint-initdb.d/" + filepath.Base("testdata/init-db.sql"),
				FileMode:          0755,
			},
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Error(err)
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Error(err)
	}

	log.Println("Giving postgres a few extra seconds to fully start")
	time.Sleep(2 * time.Second)

	return port, container, err
}

func terminateContainer(ctx context.Context, t *testing.T, container testcontainers.Container) {
	if err := container.Terminate(ctx); err != nil {
		t.Fatalf("Failed to terminate container: %s", err.Error())
	}
}

func createMockServer() (*httptest.ResponseRecorder, *httptest.Server) {
	rr := httptest.NewRecorder()
	l, err := net.Listen("tcp", "localhost:9095")
	if err != nil {
		log.Fatal(err)
	}
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		engine.SafeWrite(w.Write, []byte(r.URL.String()))
	}))
	err = ts.Listener.Close()
	if err != nil {
		log.Fatal(err)
	}
	ts.Listener = l
	ts.Start()
	return rr, ts
}

func createRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if req == nil || err != nil {
		return req, err
	}
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chi.NewRouteContext()))
	return req, err
}
