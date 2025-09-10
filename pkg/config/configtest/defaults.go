package configtest

import (
	"io"

	"github.com/smartcontractkit/chainlink-common/pkg/config/configdoc"
)

// DocDefaultsOnly reads only the default values from a docs TOML file and decodes in to cfg.
// Fields without defaults will set to zero values.
// Arrays of tables are ignored.
// Deprecated: use configdoc.DefaultsOnly
func DocDefaultsOnly(r io.Reader, cfg any, decode func(io.Reader, any) error) error {
	return configdoc.DefaultsOnly(r, cfg, decode)
}
