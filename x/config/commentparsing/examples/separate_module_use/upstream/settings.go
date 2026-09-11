// Package upstream is a dependency in its own module, and generates its own DocComments: a
// consumer in another module has no path to this source, only to what was compiled from it.
package upstream

//go:generate go run ./gen

// Settings is embedded into a consumer's config, so its fields sit in the enclosing section.
type Settings struct {
	// Region is the deployment region.
	Region string `toml:"region" validate:"required,oneof=us eu ap"`

	// Tenant scopes every request. Leave empty for the shared tenant.
	Tenant string `toml:"tenant"`
}

// Retry is a field rather than an embed, so it renders as a section of its own.
type Retry struct {
	// MaxAttempts is the total tries, including the first.
	MaxAttempts int `toml:"max_attempts" validate:"gte=1"`

	// Backoff multiplies the delay after each failed attempt.
	Backoff float64 `toml:"backoff" validate:"gt=1"`
}
