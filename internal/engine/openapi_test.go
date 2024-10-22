package engine

import (
	"net/url"
	"testing"

	gokoalaconfig "github.com/PDOK/gokoala/config"

	"github.com/stretchr/testify/assert"
)

func Test_newOpenAPI(t *testing.T) {
	type args struct {
		config      *gokoalaconfig.Config
		openAPIFile string
	}
	tests := []struct {
		name                         string
		args                         args
		expectedStringsInOpenAPISpec []string
	}{
		{
			name: "Test render OpenAPI spec with MINIMAL config",
			args: args{
				config: &gokoalaconfig.Config{
					Version:  "2.3.0",
					Title:    "Test API",
					Abstract: "Test API description",
					BaseURL:  gokoalaconfig.URL{URL: &url.URL{Scheme: "https", Host: "api.foobar.example", Path: "/"}},
				},
			},
			expectedStringsInOpenAPISpec: []string{
				"Landing page",
				"/conformance",
				"/api",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			openAPI := newOpenAPI(test.args.config, []string{test.args.openAPIFile}, nil)
			assert.NotNil(t, openAPI)

			// verify resulting OpenAPI spec contains expected strings (keywords, paths, etc)
			for _, expectedStr := range test.expectedStringsInOpenAPISpec {
				assert.Contains(t, string(openAPI.SpecJSON), expectedStr,
					"\"%s\" not found in spec: %s", expectedStr, string(openAPI.SpecJSON))
			}
		})
	}
}
