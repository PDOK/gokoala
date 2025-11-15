package geospatial

import (
	"net/url"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/PDOK/gomagpie/config"
	"golang.org/x/text/language"

	"github.com/PDOK/gomagpie/internal/engine"
	"github.com/stretchr/testify/assert"
)

func init() {
	// change working dir to root, to mimic behavior of 'go run' in order to resolve template files.
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../../../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestNewCollections(t *testing.T) {
	type args struct {
		e *engine.Engine
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test render templates with Collections (using OGC GeoVolumes config, since that contains collections)",
			args: args{
				e: engine.NewEngineWithConfig(&config.Config{
					Version:            "1.0.0",
					Title:              "Test API",
					Abstract:           "Test API description",
					AvailableLanguages: []config.Language{{Tag: language.Dutch}},
					BaseURL:            config.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
				}, false, true),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			collections := NewCollections(test.args.e)
			assert.NotEmpty(t, collections.engine.Templates.RenderedTemplates)
		})
	}
}
