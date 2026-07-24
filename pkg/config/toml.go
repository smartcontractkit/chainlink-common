package config

import (
	"errors"
	"io"

	"github.com/pelletier/go-toml/v2"
)

// DecodeTOML decodes toml from r in to v.
// Requires strict field matches and returns full toml.StrictMissingError details.
func DecodeTOML(r io.Reader, v any) error {
	d := toml.NewDecoder(r).DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		if strict, ok := errors.AsType[*toml.StrictMissingError](err); ok {
			return errors.New(strict.String())
		}
		return err
	}
	return nil
}
