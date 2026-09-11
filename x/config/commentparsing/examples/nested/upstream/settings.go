// Package upstream is a second package the consumer's config tree reaches into, so Load has to be
// given its source directory too.
package upstream

// Settings is embedded into the consumer's config, so its fields sit in the enclosing section.
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
