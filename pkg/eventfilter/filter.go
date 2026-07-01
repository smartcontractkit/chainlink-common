// Package eventfilter drops events at ingress before they are forwarded
// downstream, to keep sensitive data out of the pipeline.
//
// Rules are loaded from a YAML file (see config.go) into a RuleSet. Each rule
// has up to four optional parts — csaKey, domain, entity (exact metadata
// matches) and content (payload regexes) — that are ANDed within the rule;
// rules are ORed across the set. Evaluation is two-stage: metadata-only rules
// run before the payload is decoded (cheap), while rules with a content part
// run after the payload has been decoded to a proto message. Content rules
// require both domain and entity to constrain the payload scan.
package eventfilter

import (
	"context"

	"go.uber.org/zap"
)

// Event is the metadata view of an event. It feeds the metadata stage and
// serves as the precondition check for content rules.
type Event struct {
	Domain string // resolved event source
	Entity string // event type
	CSAKey string // CSA public key of the originating node, or "" if not CSA-authenticated
}

// Filter is a programmatic drop predicate evaluated in the metadata stage.
// Implementations must be safe for concurrent use; Name is used for log/metric
// attribution. Custom filters are always enforced (they have no dry-run mode).
type Filter interface {
	ShouldDrop(ctx context.Context, e *Event) bool
	Name() string
}

// Option configures a RuleSet.
type Option func(*RuleSet)

// WithLogger sets the logger used by the RuleSet.
func WithLogger(log *zap.SugaredLogger) Option {
	return func(rs *RuleSet) {
		if log != nil {
			rs.log = log
		}
	}
}

// WithCustomFilters appends code-registered metadata-stage filters.
func WithCustomFilters(filters ...Filter) Option {
	return func(rs *RuleSet) { rs.custom = append(rs.custom, filters...) }
}
