package pkg_test

import (
	"path/filepath"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/values/installer/pkg"
	"github.com/stretchr/testify/assert"
)

func TestFullProtoFiles(t *testing.T) {
	t.Run("Production version", func(t *testing.T) {
		config := &pkg.CapabilityConfig{
			Category:     "data",
			Pkg:          "analytics",
			MajorVersion: 2,
			Files:        []string{"query.proto", "response.proto"},
		}

		expected := []string{
			filepath.Join("capabilities", "data", "analytics", "v2", "query.proto"),
			filepath.Join("capabilities", "data", "analytics", "v2", "response.proto"),
		}

		assert.Equal(t, expected, config.FullProtoFiles())
	})

	t.Run("Prerelease version", func(t *testing.T) {
		config := &pkg.CapabilityConfig{
			Category:      "data",
			Pkg:           "analytics",
			MajorVersion:  2,
			Files:         []string{"query.proto", "response.proto"},
			PreReleaseTag: "alpha",
		}

		expected := []string{
			filepath.Join("capabilities", "data", "analytics", "v2alpha", "query.proto"),
			filepath.Join("capabilities", "data", "analytics", "v2alpha", "response.proto"),
		}

		assert.Equal(t, expected, config.FullProtoFiles())
	})
}

func TestCapabilityConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     pkg.CapabilityConfig
		wantErr string
	}{
		{
			name: "valid config",
			cfg: pkg.CapabilityConfig{
				Category:     "scheduler",
				Pkg:          "cron",
				MajorVersion: 1,
				Files:        []string{"a.proto"},
			},
			wantErr: "",
		},
		{
			name: "missing category",
			cfg: pkg.CapabilityConfig{
				Pkg:          "cron",
				MajorVersion: 1,
				Files:        []string{"a.proto"},
			},
			wantErr: "category must not be empty",
		},
		{
			name: "missing pkg",
			cfg: pkg.CapabilityConfig{
				Category:     "scheduler",
				MajorVersion: 1,
				Files:        []string{"a.proto"},
			},
			wantErr: "pkg must not be empty",
		},
		{
			name: "invalid major version",
			cfg: pkg.CapabilityConfig{
				Category: "scheduler",
				Pkg:      "cron",
				Files:    []string{"a.proto"},
			},
			wantErr: "major-version must be >= 1, got 0",
		},
		{
			name: "missing files",
			cfg: pkg.CapabilityConfig{
				Category:     "scheduler",
				Pkg:          "cron",
				MajorVersion: 1,
			},
			wantErr: "files must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}
