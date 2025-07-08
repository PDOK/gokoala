package features

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/PDOK/gokoala/internal/engine"
	"github.com/stretchr/testify/assert"
)

func init() {
	// change working dir to root, to mimic behavior of 'go run' in order to resolve template files.
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

// Run the benchmark with the following command:
//
//	go test -bench=BenchmarkFeatures -run=^# -benchmem -count=10 > bench1_run1.txt
//
// Install "benchstat": go install golang.org/x/perf/cmd/benchstat@latest
// Now compare the results for each benchmark before and after making a change, e.g:
//
//	benchstat bench1_run1.txt bench1_run2.txt
//
// This will summarize the difference in performance between the runs.
// To profile CPU and Memory usage run as:
//
//	go test -bench=BenchmarkFeatures -run=^# -benchmem -count=10 -cpuprofile cpu.pprof -memprofile mem.pprof
//
// Now analyse the pprof files using:
//
//	go tool pprof -web cpu.pprof
//	go tool pprof -web mem.pprof
//
// ----
func BenchmarkFeatures(b *testing.B) {
	type fields struct {
		configFile string
		url        string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "1", // output is WGS84 json, input is WGS84 bbox
			fields: fields{
				configFile: "internal/ogc/features/testdata/config_benchmark.yaml",
				url:        "http://localhost:8080/collections/dutch-addresses/items?bbox=4.651476%2C52.962408%2C4.979398%2C53.074282&f=json&limit=1000",
			},
		},
		{
			name: "2", // same as benchmark 1 above, but now the next page
			fields: fields{
				configFile: "internal/ogc/features/testdata/config_benchmark.yaml",
				url:        "http://localhost:8080/collections/dutch-addresses/items?bbox=4.651476%2C52.962408%2C4.979398%2C53.074282&cursor=Cpc%7CwXkQbQ&f=json&limit=1000",
			},
		},
		{
			name: "3", // output is WGS84 json, input is RD bbox
			fields: fields{
				configFile: "internal/ogc/features/testdata/config_benchmark.yaml",
				url:        "http://localhost:8080/collections/dutch-addresses/items?bbox=105564.79055389616405591%2C553072.85584054281935096%2C127668.63754775881534442%2C565347.87356295716017485&bbox-crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&f=json&limit=1000",
			},
		},
		{
			name: "4", // same as benchmark 3 above, but now the next page
			fields: fields{
				configFile: "internal/ogc/features/testdata/config_benchmark.yaml",
				url:        "http://localhost:8080/collections/dutch-addresses/items?bbox=105564.79055389616405591%2C553072.85584054281935096%2C127668.63754775881534442%2C565347.87356295716017485&bbox-crs=http%3A%2F%2Fwww.opengis.net%2Fdef%2Fcrs%2FEPSG%2F0%2F28992&cursor=Cyo%7CiLD6Iw&f=json&limit=1000",
			},
		},
	}
	for _, tt := range tests {
		req, err := createRequest(tt.fields.url, "dutch-addresses", "", "json")
		if err != nil {
			assert.Fail(b, err.Error())
		}
		rr, ts := createMockServer()

		newEngine, err := engine.NewEngine(tt.fields.configFile, "", "", false, true)
		assert.NoError(b, err)
		features := NewFeatures(newEngine)
		handler := features.Features()

		// Start benchmark
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				handler.ServeHTTP(rr, req)

				assert.Equal(b, 200, rr.Code)
			}
		})

		ts.Close()
	}
}
