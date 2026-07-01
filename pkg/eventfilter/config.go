package eventfilter

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// RuleSpec is the YAML shape of a single filter rule. Name is required; rules
// must have at least one criterion. Content rules also require domain and entity.
type RuleSpec struct {
	Name    string        `yaml:"name"`
	Domain  string        `yaml:"domain"`  // exact match on the event source
	Entity  string        `yaml:"entity"`  // exact match on the event type
	CSAKey  string        `yaml:"csaKey"`  // exact match on the originating CSA public key
	Content []ContentSpec `yaml:"content"` // payload conditions; requires domain+entity
	DryRun  bool          `yaml:"dryRun"`  // report-only; never drops
}

// ContentSpec is the YAML shape of one content condition.
type ContentSpec struct {
	// Field is a dot-separated path to a scalar leaf in the decoded message
	// (e.g. "config.rpc_url"). When empty, Pattern is matched against the whole
	// message (all textual leaves, recursively).
	Field   string `yaml:"field"`
	Pattern string `yaml:"pattern"`
}

type fileSpec struct {
	Rules []RuleSpec `yaml:"rules"`
}

// configError marks errors originating from rule configuration so callers can
// distinguish a bad config from a transient failure if needed.
type configError struct {
	msg string
}

func (e *configError) Error() string { return "event filter config: " + e.msg }

// LoadFile reads and parses the YAML rule file at path. Unknown keys are
// rejected. An empty file yields no rules (filtering disabled). The returned
// specs are not yet compiled; pass them to New for validation.
func LoadFile(path string) ([]RuleSpec, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open event filter config %q: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)

	var fs fileSpec
	if err := dec.Decode(&fs); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil // empty file
		}
		return nil, fmt.Errorf("parse event filter config %q: %w", path, err)
	}
	return fs.Rules, nil
}

// compileRule validates a spec and compiles its content regexes. Content rules
// require both domain and entity. domain/entity/csaKey are matched verbatim.
func compileRule(spec *RuleSpec) (*compiledRule, error) {
	if spec.Name == "" {
		return nil, &configError{msg: "rule missing required 'name'"}
	}

	cr := &compiledRule{
		name:   spec.Name,
		dryRun: spec.DryRun,
		domain: spec.Domain,
		entity: spec.Entity,
		csaKey: spec.CSAKey,
	}

	for i := range spec.Content {
		cs := spec.Content[i]
		if cs.Pattern == "" {
			return nil, &configError{msg: fmt.Sprintf("rule %q: content condition %d missing 'pattern'", spec.Name, i)}
		}
		re, cerr := regexp.Compile(cs.Pattern)
		if cerr != nil {
			return nil, &configError{msg: fmt.Sprintf("rule %q: invalid content pattern %q: %v", spec.Name, cs.Pattern, cerr)}
		}
		cr.content = append(cr.content, contentCond{field: cs.Field, pattern: re})
	}

	if len(cr.content) > 0 && (cr.domain == "" || cr.entity == "") {
		return nil, &configError{msg: fmt.Sprintf("rule %q: content rule requires both 'domain' and 'entity' to be defined", spec.Name)}
	}

	if cr.domain == "" && cr.entity == "" && cr.csaKey == "" && len(cr.content) == 0 {
		return nil, &configError{msg: fmt.Sprintf("rule %q has no criteria (would drop everything)", spec.Name)}
	}
	return cr, nil
}
