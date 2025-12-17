package config

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// change working dir to root, to mimic behavior of 'go run' in order to resolve template/config files.
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestNewConfig(t *testing.T) {
	type args struct {
		configFile string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "read valid config file",
			args: args{
				configFile: "internal/etl/testdata/config.yaml",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewConfig(tt.args.configFile)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Len(t, cfg.Collections, 1)
				assert.Equal(t, "addresses", cfg.Collections[0].ID)
				assert.Equal(t, "addresses", cfg.Collections[0].Tables[0].Table)
				assert.Equal(t, "fid", cfg.Collections[0].Tables[0].FID)
			}
		})
	}
}
