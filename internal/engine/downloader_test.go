package engine

import (
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	// change working dir to root, to mimic behavior of 'go run' in order to resolve template/config files.
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestDownload(t *testing.T) {
	type args struct {
		url           string
		parallelism   int
		tlsSkipVerify bool
		timeout       time.Duration
		retryDelay    time.Duration
		retryMaxDelay time.Duration
		maxRetries    int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test concurrent download",
			args: args{
				url:           "http://localhost:8989/bag.gpkg",
				parallelism:   2,
				tlsSkipVerify: false,
				timeout:       60 * time.Second,
				retryDelay:    1 * time.Second,
				retryMaxDelay: 60 * time.Second,
				maxRetries:    2,
			},
		},
		{
			name: "Test regular download",
			args: args{
				url:           "http://localhost:8989/bag.gpkg",
				parallelism:   1,
				tlsSkipVerify: false,
				timeout:       60 * time.Second,
				retryDelay:    1 * time.Second,
				retryMaxDelay: 60 * time.Second,
				maxRetries:    2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputFile, err := os.CreateTemp("", "gpkg")
			assert.NoError(t, err)
			defer os.Remove(outputFile.Name())

			ts := createMockServer()
			defer ts.Close()

			parsedURL, err := url.Parse(tt.args.url)
			assert.NoError(t, err)

			downloadTime, err := Download(*parsedURL, outputFile.Name(), tt.args.parallelism, tt.args.tlsSkipVerify,
				tt.args.timeout, tt.args.retryDelay, tt.args.retryMaxDelay, tt.args.maxRetries)
			assert.NoError(t, err)
			assert.Greater(t, *downloadTime, 100*time.Nanosecond)
			assert.FileExists(t, outputFile.Name())
			stat, err := outputFile.Stat()
			assert.NoError(t, err)
			assert.Greater(t, stat.Size(), int64(1*1024))
		})
	}
}

func createMockServer() *httptest.Server {
	l, err := net.Listen("tcp", "localhost:8989")
	if err != nil {
		log.Fatal(err)
	}
	ts := httptest.NewUnstartedServer(http.FileServer(http.Dir("internal/ogc/features/datasources/geopackage/testdata")))
	err = ts.Listener.Close()
	if err != nil {
		log.Fatal(err)
	}
	ts.Listener = l
	ts.Start()
	return ts
}
