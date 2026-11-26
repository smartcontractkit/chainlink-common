package telemetry

import "strings"

// AttrLoopID is the beholder resource attribute carrying TelemetryConfig.TelemetryLoopID.
const AttrLoopID = "chainlink.loop.id"

// loopIDSegmentMaxLen caps each segment so free-text parts (e.g. job names) stay label-friendly.
const loopIDSegmentMaxLen = 128

// NewLoopID builds a telemetry-only LOOP identity from the LOOP kind and the parts that make it
// unique on a node, joined by ".". Examples:
//
//	StandardCapabilities.<jobID>.<jobName>
//	Median.<contractID>.<jobID>
//	<pluginName>.<contractID>.<jobID>
//	Relayer.<network>.<chainID>
//
// Every segment is sanitized: runes outside [A-Za-z0-9_-] (including ".") become "_", empty
// segments become "_", and segments longer than 128 runes are truncated. "." therefore only ever
// separates segments. Callers must include a part that is unique per node (a job ID, or network
// and chain ID), so sanitizing and truncation do not make two LOOPs collide.
func NewLoopID(kind string, parts ...string) string {
	segments := make([]string, 0, len(parts)+1)
	segments = append(segments, sanitizeLoopIDSegment(kind))
	for _, p := range parts {
		segments = append(segments, sanitizeLoopIDSegment(p))
	}
	return strings.Join(segments, ".")
}

func sanitizeLoopIDSegment(s string) string {
	var b strings.Builder
	n := 0
	for _, r := range s {
		if n == loopIDSegmentMaxLen {
			break
		}
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
		n++
	}
	if b.Len() == 0 {
		return "_"
	}
	return b.String()
}
