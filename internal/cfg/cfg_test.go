package cfg_test

import (
	"path/filepath"
	"testing"

	"github.com/bxrne/darwin/internal/cfg"
)

func TestLoadConfigFiles(t *testing.T) {
	cases := []struct {
		name    string
		path    string
		wantErr bool
	}{
		// Samples
		{"valid_bitstring", "../testdata/config/valid_bitstring.toml", true},
		{"valid_tree", "../testdata/config/valid_tree.toml", true},
		{"valid_action_tree", "../testdata/config/valid_action_tree.toml", true},
		{"both_enabled", "../testdata/config/both_enabled.toml", true},
		{"invalid_evolution", "../testdata/config/invalid_evolution.toml", true},
		{"valid_metrics_enabled", "../testdata/config/valid_metrics_enabled.toml", true},
		{"invalid_metrics", "../testdata/config/invalid_metrics.toml", true},
		// Examples
		{"valid_default", "../config/default.toml", false},
		{"valid_small", "../config/small.toml", true},
		{"valid_medium", "../config/medium.toml", true},
		{"valid_large", "../config/large.toml", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := filepath.Join("..", c.path) // adjust if test lives elsewhere
			_, err := cfg.LoadConfig(p)
			if (err != nil) != c.wantErr {
				t.Fatalf("LoadConfig(%s) err=%v wantErr=%v", p, err, c.wantErr)
			}
		})
	}
}
