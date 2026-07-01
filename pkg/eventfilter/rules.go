package eventfilter

import (
	"context"
	"regexp"

	"go.uber.org/zap"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// contentCond is one compiled content condition. An empty field matches against
// the whole decoded message (all textual leaves); otherwise field is a
// dot-separated path to a scalar leaf whose stringified value is matched.
type contentCond struct {
	field   string
	pattern *regexp.Regexp
}

// compiledRule is a validated rule. Empty string metadata fields are "not a criterion";
// a non-empty content slice makes this a content (stage-2) rule.
type compiledRule struct {
	name    string
	domain  string // exact match; "" means any
	entity  string // exact match; "" means any
	csaKey  string // exact match; "" means any
	content []contentCond
	dryRun  bool
}

// RuleSet evaluates an event against a list of rules. Rules are ORed; the parts
// within a rule are ANDed. It is immutable after construction and safe for
// concurrent use.
type RuleSet struct {
	metaRules    []*compiledRule // content == nil: evaluated pre-encode (stage 1)
	contentRules []*compiledRule // content != nil: evaluated post-decode (stage 2)
	custom       []Filter        // code-registered metadata-stage filters (always enforced)
	log          *zap.SugaredLogger
}

// New compiles and validates the rule specs into a RuleSet, failing fast on any
// invalid content pattern regex, missing/duplicate name, empty content pattern,
// or zero-criteria rule. When no rule and no custom filter is configured the RuleSet is a no-op
// (see Enabled), so callers may construct one unconditionally.
func New(specs []RuleSpec, opts ...Option) (*RuleSet, error) {
	rs := &RuleSet{log: zap.NewNop().Sugar()}

	seen := make(map[string]bool, len(specs))
	for i := range specs {
		cr, err := compileRule(&specs[i])
		if err != nil {
			return nil, err
		}
		if seen[cr.name] {
			return nil, &configError{msg: "duplicate filter rule name " + cr.name}
		}
		seen[cr.name] = true

		if len(cr.content) > 0 {
			rs.contentRules = append(rs.contentRules, cr)
		} else {
			rs.metaRules = append(rs.metaRules, cr)
		}
	}

	for _, opt := range opts {
		opt(rs)
	}
	return rs, nil
}

// Enabled reports whether the RuleSet can ever drop an event.
func (rs *RuleSet) Enabled() bool {
	return len(rs.metaRules) > 0 || len(rs.contentRules) > 0 || len(rs.custom) > 0
}

// HasContentRules reports whether any content (stage-2) rules are configured.
func (rs *RuleSet) HasContentRules() bool { return len(rs.contentRules) > 0 }

// MatchMetadata evaluates the metadata-only rules and custom filters against e.
// It returns whether a rule matched, the matched rule name, and whether that
// match is enforced (drop) versus dry-run (report only). An enforced match wins
// over a dry-run match.
func (rs *RuleSet) MatchMetadata(ctx context.Context, e *Event) (matched bool, rule string, enforced bool) {
	dryRule := ""
	for _, r := range rs.metaRules {
		if !matchMeta(r, e) {
			continue
		}
		if !r.dryRun {
			return true, r.name, true
		}
		if dryRule == "" {
			dryRule = r.name
		}
	}
	for _, f := range rs.custom {
		if f.ShouldDrop(ctx, e) {
			return true, f.Name(), true
		}
	}
	if dryRule != "" {
		return true, dryRule, false
	}
	return false, "", false
}

// ContentCandidate reports whether any content rule's metadata precondition
// matches e. When false, the caller can skip decoding the payload entirely.
func (rs *RuleSet) ContentCandidate(e *Event) bool {
	for _, r := range rs.contentRules {
		if matchMeta(r, e) {
			return true
		}
	}
	return false
}

// MatchContent evaluates the content rules against e (metadata precondition) and
// the decoded message. Same return contract and enforced-wins behavior as
// MatchMetadata.
func (rs *RuleSet) MatchContent(e *Event, msg protoreflect.Message) (matched bool, rule string, enforced bool) {
	dryRule := ""
	for _, r := range rs.contentRules {
		if matchMeta(r, e) && contentMatches(r.content, msg) {
			if !r.dryRun {
				return true, r.name, true
			}
			if dryRule == "" {
				dryRule = r.name
			}
		}
	}
	return dryRule != "", dryRule, false
}

// matchMeta reports whether every set metadata criterion matches its field (AND). A
// rule with no metadata criteria matches vacuously (only valid for content
// rules; metadata-only rules are required to set at least one criterion).
func matchMeta(r *compiledRule, e *Event) bool {
	if r.domain != "" && r.domain != e.Domain {
		return false
	}
	if r.entity != "" && r.entity != e.Entity {
		return false
	}
	if r.csaKey != "" && r.csaKey != e.CSAKey {
		return false
	}
	return true
}

// contentMatches reports whether every content condition matches the message
// (AND). An unresolvable field path never matches, so it cannot cause a drop.
func contentMatches(conds []contentCond, msg protoreflect.Message) bool {
	for _, c := range conds {
		var (
			val string
			ok  bool
		)
		if c.field == "" {
			val, ok = allTextValues(msg), true
		} else {
			val, ok = resolveFieldValue(msg, c.field)
		}
		if !ok || !c.pattern.MatchString(val) {
			return false
		}
	}
	return true
}
