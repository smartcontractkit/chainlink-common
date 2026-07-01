package eventfilter

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

// eventPayloadDesc is a small proto schema used only in tests to exercise the
// filter against a proto message. It mirrors:
//
//	message EventPayload {
//	  string string_val = 1;
//	  EventMetadata metadata = 2;
//	}
//	message EventMetadata { string string_val = 1; }
//
// We build the descriptor at init time via dynamicpb so tests don't require
// protoc or generated .pb.go fixtures.
var eventPayloadDesc protoreflect.MessageDescriptor

func init() {
	fdp := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("eventfilter/testdata.proto"),
		Package: proto.String("eventfilter.test"),
		Syntax:  proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("EventMetadata"),
				Field: []*descriptorpb.FieldDescriptorProto{{
					Name:   proto.String("string_val"),
					Number: proto.Int32(1),
					Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
					Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				}},
			},
			{
				Name: proto.String("EventPayload"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   proto.String("string_val"),
						Number: proto.Int32(1),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
					{
						Name:     proto.String("metadata"),
						Number:   proto.Int32(2),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".eventfilter.test.EventMetadata"),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
					},
				},
			},
		},
	}
	fd, err := protodesc.NewFile(fdp, nil)
	if err != nil {
		panic(err)
	}
	eventPayloadDesc = fd.Messages().ByName("EventPayload")
}

func simpleMsg(value string) protoreflect.Message {
	m := dynamicpb.NewMessage(eventPayloadDesc)
	m.Set(eventPayloadDesc.Fields().ByName("string_val"), protoreflect.ValueOfString(value))
	return m
}

func nestedMsg(value string) protoreflect.Message {
	metaFD := eventPayloadDesc.Fields().ByName("metadata")
	metaMsg := dynamicpb.NewMessage(metaFD.Message())
	metaMsg.Set(metaFD.Message().Fields().ByName("string_val"), protoreflect.ValueOfString(value))

	m := dynamicpb.NewMessage(eventPayloadDesc)
	m.Set(metaFD, protoreflect.ValueOfMessage(metaMsg))
	return m
}

func mustNewSingle(t *testing.T, spec RuleSpec, opts ...Option) *RuleSet {
	t.Helper()
	rs, err := New([]RuleSpec{spec}, opts...)
	require.NoError(t, err)
	return rs
}

func TestNew_InvalidRegex(t *testing.T) {
	for _, tt := range []struct {
		name string
		spec RuleSpec
	}{
		{"content invalid pattern", RuleSpec{Name: "r", Content: []ContentSpec{{Pattern: "("}}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New([]RuleSpec{tt.spec})
			require.Error(t, err)
		})
	}
}

func TestNew_DuplicateName(t *testing.T) {
	specs := []RuleSpec{
		{Name: "dup", Entity: "Foo"},
		{Name: "dup", Domain: "bar"},
	}
	_, err := New(specs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestNew_ZeroCriteriaRejected(t *testing.T) {
	_, err := New([]RuleSpec{{Name: "empty"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no criteria")
}

func TestNew_ContentMissingPattern(t *testing.T) {
	_, err := New([]RuleSpec{{
		Name:    "r",
		Entity:  "Foo",
		Content: []ContentSpec{{Field: "x"}},
	}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing 'pattern'")
}

func TestNew_ContentInvalidPattern(t *testing.T) {
	_, err := New([]RuleSpec{{
		Name:    "r",
		Content: []ContentSpec{{Pattern: "(invalid"}},
	}})
	require.Error(t, err)
}

func TestNew_ContentRule_RequiresDomainAndEntity(t *testing.T) {
	for _, tt := range []struct {
		name string
		spec RuleSpec
	}{
		{"missing domain", RuleSpec{Name: "r", Entity: "E", Content: []ContentSpec{{Pattern: "x"}}}},
		{"missing entity", RuleSpec{Name: "r", Domain: "d", Content: []ContentSpec{{Pattern: "x"}}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New([]RuleSpec{tt.spec})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "content rule")
		})
	}
}

func TestMatchMetadata_Basic(t *testing.T) {
	ctx := context.Background()

	for _, tt := range []struct {
		name         string
		spec         RuleSpec
		event        *Event
		wantMatched  bool
		wantEnforced bool
	}{
		{
			name:         "entity match enforced",
			spec:         RuleSpec{Name: "r", Entity: "NodeTOMLConfig"},
			event:        &Event{Entity: "NodeTOMLConfig", Domain: "cre"},
			wantMatched:  true,
			wantEnforced: true,
		},
		{
			name:        "entity no match",
			spec:        RuleSpec{Name: "r", Entity: "NodeTOMLConfig"},
			event:       &Event{Entity: "Observation", Domain: "cre"},
			wantMatched: false,
		},
		{
			name:         "domain exact match",
			spec:         RuleSpec{Name: "r", Domain: "cre"},
			event:        &Event{Domain: "cre"},
			wantMatched:  true,
			wantEnforced: true,
		},
		{
			name:        "domain no match on different value",
			spec:        RuleSpec{Name: "r", Domain: "cre"},
			event:       &Event{Domain: "cre-extra"},
			wantMatched: false,
		},
		{
			name:         "csa_key match",
			spec:         RuleSpec{Name: "r", CSAKey: "abc123"},
			event:        &Event{CSAKey: "abc123"},
			wantMatched:  true,
			wantEnforced: true,
		},
		{
			name:        "AND: entity matches but domain does not",
			spec:        RuleSpec{Name: "r", Entity: "NodeConfig", Domain: "cre"},
			event:       &Event{Entity: "NodeConfig", Domain: "platform"},
			wantMatched: false,
		},
		{
			name:         "AND: all criteria satisfied",
			spec:         RuleSpec{Name: "r", Entity: "NodeConfig", Domain: "cre"},
			event:        &Event{Entity: "NodeConfig", Domain: "cre"},
			wantMatched:  true,
			wantEnforced: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rs := mustNewSingle(t, tt.spec)
			matched, rule, enforced := rs.MatchMetadata(ctx, tt.event)
			assert.Equal(t, tt.wantMatched, matched)
			if tt.wantMatched {
				assert.Equal(t, "r", rule)
				assert.Equal(t, tt.wantEnforced, enforced)
			}
		})
	}
}

func TestMatchMetadata_OR(t *testing.T) {
	ctx := context.Background()
	rs, err := New([]RuleSpec{
		{Name: "rule-a", Entity: "TypeA"},
		{Name: "rule-b", Entity: "TypeB"},
	})
	require.NoError(t, err)

	matched, rule, enforced := rs.MatchMetadata(ctx, &Event{Entity: "TypeB"})
	assert.True(t, matched)
	assert.Equal(t, "rule-b", rule)
	assert.True(t, enforced)

	matched, _, _ = rs.MatchMetadata(ctx, &Event{Entity: "TypeC"})
	assert.False(t, matched)
}

func TestMatchMetadata_EnforcedWinsOverDryRun(t *testing.T) {
	ctx := context.Background()
	rs, err := New([]RuleSpec{
		{Name: "dry", Entity: "Anything", DryRun: true},
		{Name: "hard", Domain: "cre"},
	})
	require.NoError(t, err)

	matched, rule, enforced := rs.MatchMetadata(ctx, &Event{Entity: "Anything", Domain: "cre"})
	assert.True(t, matched)
	assert.Equal(t, "hard", rule)
	assert.True(t, enforced)
}

func TestMatchMetadata_DryRun(t *testing.T) {
	ctx := context.Background()
	rs := mustNewSingle(t, RuleSpec{Name: "dry-rule", Entity: "NodeConfig", DryRun: true})

	matched, rule, enforced := rs.MatchMetadata(ctx, &Event{Entity: "NodeConfig"})
	assert.True(t, matched)
	assert.Equal(t, "dry-rule", rule)
	assert.False(t, enforced)
}

func TestRuleSet_Enabled(t *testing.T) {
	disabled, err := New(nil)
	require.NoError(t, err)
	assert.False(t, disabled.Enabled())

	enabled := mustNewSingle(t, RuleSpec{Name: "r", Domain: "cre"})
	assert.True(t, enabled.Enabled())
}

func TestRuleSet_HasContentRules(t *testing.T) {
	metaOnly := mustNewSingle(t, RuleSpec{Name: "r", Entity: "Foo"})
	assert.False(t, metaOnly.HasContentRules())

	withContent := mustNewSingle(t, RuleSpec{
		Name:    "c",
		Domain:  "d",
		Entity:  "e",
		Content: []ContentSpec{{Pattern: "secret"}},
	})
	assert.True(t, withContent.HasContentRules())
}

type staticFilter struct {
	name string
	drop bool
}

func (f *staticFilter) ShouldDrop(_ context.Context, _ *Event) bool { return f.drop }
func (f *staticFilter) Name() string                                { return f.name }

func TestRuleSet_CustomFilters(t *testing.T) {
	ctx := context.Background()

	t.Run("custom filter drops and is enabled", func(t *testing.T) {
		rs, err := New(nil, WithCustomFilters(&staticFilter{name: "custom-1", drop: true}))
		require.NoError(t, err)
		assert.True(t, rs.Enabled())

		matched, rule, enforced := rs.MatchMetadata(ctx, &Event{Entity: "Anything"})
		assert.True(t, matched)
		assert.Equal(t, "custom-1", rule)
		assert.True(t, enforced)
	})

	t.Run("metadata rule evaluated before custom filter", func(t *testing.T) {
		rs, err := New(
			[]RuleSpec{{Name: "meta-rule", Entity: "Foo"}},
			WithCustomFilters(&staticFilter{name: "custom-1", drop: true}),
		)
		require.NoError(t, err)

		matched, rule, enforced := rs.MatchMetadata(ctx, &Event{Entity: "Foo"})
		assert.True(t, matched)
		assert.Equal(t, "meta-rule", rule)
		assert.True(t, enforced)
	})

	t.Run("custom filter passes", func(t *testing.T) {
		rs, err := New(nil, WithCustomFilters(&staticFilter{name: "custom-1", drop: false}))
		require.NoError(t, err)

		matched, _, _ := rs.MatchMetadata(ctx, &Event{Entity: "Anything"})
		assert.False(t, matched)
	})
}

func TestContentCandidate(t *testing.T) {
	rs, err := New([]RuleSpec{
		{Name: "meta-only", Entity: "MetaOnly"},
		{Name: "content-rule", Domain: "cre", Entity: "Anything", Content: []ContentSpec{{Pattern: "secret"}}},
	})
	require.NoError(t, err)

	assert.True(t, rs.ContentCandidate(&Event{Domain: "cre", Entity: "Anything"}))
	assert.False(t, rs.ContentCandidate(&Event{Domain: "other", Entity: "MetaOnly"}))
}

func TestMatchContent_FieldPath(t *testing.T) {
	rs := mustNewSingle(t, RuleSpec{
		Name:    "rpc-key",
		Domain:  "d",
		Entity:  "e",
		Content: []ContentSpec{{Field: "string_val", Pattern: "secret"}},
	})

	msg := simpleMsg("my-secret-value")
	matched, rule, enforced := rs.MatchContent(&Event{Domain: "d", Entity: "e"}, msg)
	assert.True(t, matched)
	assert.Equal(t, "rpc-key", rule)
	assert.True(t, enforced)

	msgNoMatch := simpleMsg("safe-value")
	matched, _, _ = rs.MatchContent(&Event{Domain: "d", Entity: "e"}, msgNoMatch)
	assert.False(t, matched)
}

func TestMatchContent_WholeMsgMatch(t *testing.T) {
	rs := mustNewSingle(t, RuleSpec{
		Name:    "whole-msg",
		Domain:  "d",
		Entity:  "e",
		Content: []ContentSpec{{Pattern: "hidden"}},
	})

	msg := simpleMsg("contains-hidden-text")
	matched, _, _ := rs.MatchContent(&Event{Domain: "d", Entity: "e"}, msg)
	assert.True(t, matched)
}

func TestMatchContent_NestedField(t *testing.T) {
	rs := mustNewSingle(t, RuleSpec{
		Name:    "nested",
		Domain:  "d",
		Entity:  "e",
		Content: []ContentSpec{{Field: "metadata.string_val", Pattern: "^sensitive$"}},
	})

	msg := nestedMsg("sensitive")
	matched, _, enforced := rs.MatchContent(&Event{Domain: "d", Entity: "e"}, msg)
	assert.True(t, matched)
	assert.True(t, enforced)

	msgNoMatch := nestedMsg("safe")
	matched, _, _ = rs.MatchContent(&Event{Domain: "d", Entity: "e"}, msgNoMatch)
	assert.False(t, matched)
}

func TestMatchContent_MissingFieldNotDropped(t *testing.T) {
	rs := mustNewSingle(t, RuleSpec{
		Name:    "r",
		Domain:  "d",
		Entity:  "e",
		Content: []ContentSpec{{Field: "nonexistent.path", Pattern: ".*"}},
	})
	msg := simpleMsg("anything")
	matched, _, _ := rs.MatchContent(&Event{Domain: "d", Entity: "e"}, msg)
	assert.False(t, matched)
}

func TestMatchContent_MetaPreconditionFilters(t *testing.T) {
	rs := mustNewSingle(t, RuleSpec{
		Name:    "r",
		Domain:  "cre",
		Entity:  "E",
		Content: []ContentSpec{{Pattern: "secret"}},
	})
	msg := simpleMsg("secret")

	matched, _, _ := rs.MatchContent(&Event{Domain: "other", Entity: "E"}, msg)
	assert.False(t, matched, "wrong domain: content rule must not match")

	matched, _, _ = rs.MatchContent(&Event{Domain: "cre", Entity: "E"}, msg)
	assert.True(t, matched, "correct domain: content rule must match")
}

func TestMatchContent_DryRun(t *testing.T) {
	rs := mustNewSingle(t, RuleSpec{
		Name:    "dry-content",
		Domain:  "d",
		Entity:  "e",
		Content: []ContentSpec{{Pattern: "secret"}},
		DryRun:  true,
	})
	msg := simpleMsg("secret-value")
	matched, rule, enforced := rs.MatchContent(&Event{Domain: "d", Entity: "e"}, msg)
	assert.True(t, matched)
	assert.Equal(t, "dry-content", rule)
	assert.False(t, enforced)
}

func TestLoadFile_Valid(t *testing.T) {
	content := `
rules:
  - name: drop-config
    entity: "BaseMessage"
    domain: "platform"
    dryRun: false
  - name: drop-rpc-keys
    domain: "platform"
    entity: "BaseMessage"
    content:
      - field: "rpc_url"
        pattern: "https?://[^@\\s]*:[^@\\s]*@"
    dryRun: true
`
	path := writeTempYAML(t, content)
	specs, err := LoadFile(path)
	require.NoError(t, err)
	require.Len(t, specs, 2)
	assert.Equal(t, "drop-config", specs[0].Name)
	assert.Equal(t, "drop-rpc-keys", specs[1].Name)
	assert.True(t, specs[1].DryRun)
}

func TestLoadFile_EmptyFile(t *testing.T) {
	path := writeTempYAML(t, "")
	specs, err := LoadFile(path)
	require.NoError(t, err)
	assert.Empty(t, specs)
}

func TestLoadFile_MissingFile(t *testing.T) {
	_, err := LoadFile("/nonexistent/path/event-filter.yaml")
	require.Error(t, err)
}

func TestLoadFile_MalformedYAML(t *testing.T) {
	path := writeTempYAML(t, "rules: [{{{{")
	_, err := LoadFile(path)
	require.Error(t, err)
}

func TestLoadFile_UnknownKeyRejected(t *testing.T) {
	path := writeTempYAML(t, `
rules:
  - name: r
    entity: "Foo"
    unknownField: "bad"
`)
	_, err := LoadFile(path)
	require.Error(t, err)
}

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "event-filter.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

// BenchmarkMatchMetadata measures the metadata-stage cost (no payload decode)
// that runs on every event when filtering is enabled.
//
// Run with:
//
//	go test -run=^$ -bench='BenchmarkMatch' -benchmem ./pkg/eventfilter/
func BenchmarkMatchMetadata(b *testing.B) {
	rs, err := New([]RuleSpec{{Name: "r", Domain: "node-ops", Entity: "NodeConfig"}})
	require.NoError(b, err)
	e := &Event{Domain: "other-domain", Entity: "OtherEntity"}
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _, _ = rs.MatchMetadata(ctx, e)
	}
}

// BenchmarkMatchContent_FieldPath measures the cost of resolving a single
// dot-separated proto field path and running one regex on a small message.
// This is the per-event overhead added by the content stage for candidates.
func BenchmarkMatchContent_FieldPath(b *testing.B) {
	rs, err := New([]RuleSpec{{
		Name:    "r",
		Domain:  "d",
		Entity:  "EventPayload",
		Content: []ContentSpec{{Field: "metadata.string_val", Pattern: "secret"}},
	}})
	require.NoError(b, err)
	e := &Event{Domain: "d", Entity: "EventPayload"}
	msg := nestedMsg("safe-value-with-no-match")

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _, _ = rs.MatchContent(e, msg)
	}
}

// BenchmarkMatchContent_WholeMessage measures the cost of the whole-message
// content mode, which walks every string/bytes leaf in the decoded message.
func BenchmarkMatchContent_WholeMessage(b *testing.B) {
	rs, err := New([]RuleSpec{{
		Name:    "r",
		Domain:  "d",
		Entity:  "EventPayload",
		Content: []ContentSpec{{Pattern: "secret"}},
	}})
	require.NoError(b, err)
	e := &Event{Domain: "d", Entity: "EventPayload"}
	msg := nestedMsg("safe-value-with-no-match")

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _, _ = rs.MatchContent(e, msg)
	}
}
