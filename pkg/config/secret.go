package config

import (
	"encoding"
	"fmt"
	"net/url"
)

const redacted = "xxxxx"

var (
	_ fmt.Stringer           = (*SecretString)(nil)
	_ encoding.TextMarshaler = (*SecretString)(nil)
)

// SecretString is a string that formats and encodes redacted, as "xxxxx".
type SecretString string

func NewSecretString(s string) *SecretString { return (*SecretString)(&s) }

func (s SecretString) String() string { return redacted }

func (s SecretString) GoString() string { return redacted }

func (s SecretString) MarshalText() ([]byte, error) { return []byte(redacted), nil }

var (
	_ fmt.Stringer             = (*SecretURL)(nil)
	_ encoding.TextMarshaler   = (*SecretURL)(nil)
	_ encoding.TextUnmarshaler = (*SecretURL)(nil)
)

// SecretURL is a URL that formats and encodes redacted, as "xxxxx".
type SecretURL URL

func NewSecretURL(u *URL) *SecretURL { return (*SecretURL)(u) }

func MustSecretURL(u string) *SecretURL { return NewSecretURL(MustParseURL(u)) }

func (s *SecretURL) String() string { return redacted }

func (s *SecretURL) GoString() string { return redacted }

func (s *SecretURL) URL() *url.URL { return (*URL)(s).URL() }

func (s *SecretURL) MarshalText() ([]byte, error) { return []byte(redacted), nil }

func (s *SecretURL) UnmarshalText(text []byte) error {
	if err := (*URL)(s).UnmarshalText(text); err != nil {
		//opt: if errors.Is(url.Error), just redact the err.URL field?
		return fmt.Errorf("failed to parse url: %s", redacted)
	}
	return nil
}
