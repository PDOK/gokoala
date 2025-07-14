package features

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/docker/go-connections/nat"
	"github.com/go-chi/chi/v5"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	postgresPortEnv = "DB_PORT"
	postgresCompose = "internal/ogc/features/datasources/postgres/testdata/docker-compose.yaml"
)

func init() {
	// change working dir to root, to mimic behavior of 'go run' in order to resolve template files.
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../../")
	if err := os.Chdir(dir); err != nil {
		panic(err)
	}
}

// TestMain package-wide setup/teardown and test utils
func TestMain(m *testing.M) {
	ctx := context.Background()

	stack := setup(ctx)
	exitCode := m.Run()
	teardown(ctx, stack)
	os.Exit(exitCode)
}

func setup(ctx context.Context) *compose.DockerCompose {
	port, stack, err := setupPostgres(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if err = os.Setenv(postgresPortEnv, port.Port()); err != nil {
		log.Fatal(err)
	}
	return stack
}

func teardown(ctx context.Context, stack *compose.DockerCompose) {
	// We would rather use t.Setenv() but this isn't possible in TestMain.
	// Therefore, it's important to unset the env variable ourselves since this isn't done automatically
	if err := os.Unsetenv(postgresPortEnv); err != nil {
		log.Fatal(err)
	}
	if err := terminateStack(ctx, stack); err != nil {
		log.Fatal(err)
	}
}

// setupPostgres start PostgreSQL and fill with testdata derived from GeoPackages.
func setupPostgres(ctx context.Context) (nat.Port, *compose.DockerCompose, error) {
	log.Println("Setting up postgres")
	stack, err := compose.NewDockerComposeWith(compose.WithStackFiles(postgresCompose))
	if err != nil {
		return "", nil, err
	}

	err = stack.
		WaitForService("postgres", wait.ForListeningPort("5432/tcp")).
		WaitForService("postgres-init-data", wait.ForExit()).
		Up(ctx, compose.Wait(true))
	if err != nil {
		return "", nil, err
	}

	container, err := stack.ServiceContainer(ctx, "postgres")
	if err != nil {
		return "", nil, err
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return "", nil, err
	}

	log.Println("Giving postgres a few extra seconds to fully start")
	time.Sleep(2 * time.Second)
	log.Printf("Postgres running at port %s", port.Port())

	return port, stack, err
}

func terminateStack(ctx context.Context, stack *compose.DockerCompose) error {
	log.Println("Terminate postgres stack")
	return stack.Down(
		ctx,
		compose.RemoveOrphans(true),
		compose.RemoveVolumes(true),
		compose.RemoveImagesLocal,
	)
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

func createRequest(url string, collectionID string, featureID string, format string) (*http.Request, error) {
	url = strings.ReplaceAll(url, ":collectionId", collectionID)
	url = strings.ReplaceAll(url, ":featureId", featureID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if req == nil || err != nil {
		return req, err
	}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("collectionId", collectionID)
	rctx.URLParams.Add("featureId", featureID)

	queryString := req.URL.Query()
	queryString.Add(engine.FormatParam, format)
	req.URL.RawQuery = queryString.Encode()

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req, err
}

func normalize(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), ""))
}

func printActual(rr *httptest.ResponseRecorder) {
	log.Print("\n===========================\n")
	log.Print("\n==> ACTUAL RESPONSE BELOW. Copy/paste and compare with response in file. " +
		"Note that in the case of HTML output we only compare relevant snippets instead of the whole file.")
	log.Print("\n===========================\n")
	log.Print(rr.Body.String()) // to ease debugging & updating expected results
	log.Print("\n===========================\n")
}
