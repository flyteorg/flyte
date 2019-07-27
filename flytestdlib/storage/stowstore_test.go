package storage

import (
	"testing"

	"github.com/graymeta/stow/google"
	"github.com/stretchr/testify/assert"

	"github.com/lyft/flytestdlib/config"
	"github.com/lyft/flytestdlib/internal/utils"
	"github.com/lyft/flytestdlib/promutils"
)

func Test_newStowRawStore(t *testing.T) {
	type args struct {
		cfg          *Config
		metricsScope promutils.Scope
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"fail", args{&Config{}, promutils.NewTestScope()}, true},
		{"google", args{&Config{
			InitContainer: "flyte",
			Stow: &StowConfig{
				Kind: google.Kind,
				Config: map[string]string{
					google.ConfigProjectId: "x",
					google.ConfigScopes:    "y",
				},
			},
		}, promutils.NewTestScope()}, true},
		{"minio", args{&Config{
			Type:          TypeMinio,
			InitContainer: "some-container",
			Connection: ConnectionConfig{
				Endpoint: config.URL{URL: utils.MustParseURL("http://minio:9000")},
			},
		}, promutils.NewTestScope()}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newStowRawStore(tt.args.cfg, tt.args.metricsScope)
			if tt.wantErr {
				assert.Error(t, err, "newStowRawStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NotNil(t, got, "Expected rawstore, found nil!")
		})
	}
}
