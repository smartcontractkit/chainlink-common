package dontime

import (
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/limits"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/dontime/pb"
)

const (
	fuzzN         = 4
	fuzzF         = 1
	fuzzBatchSize = 3
	fuzzBaseTS    = int64(1_700_000_000_000)
	fuzzHonest    = fuzzN - 1
	minIncreaseMs = int64(1)
)

func newFuzzPlugin(t *testing.T, sequenced bool) *Plugin {
	t.Helper()
	offchainCfg := &pb.Config{
		MaxBatchSize:         fuzzBatchSize,
		MinTimeIncrease:      int64(time.Millisecond),
		ExecutionRemovalTime: durationpb.New(defaultExecutionRemovalTime),
	}
	p, err := NewPlugin(NewStore(DefaultRequestTimeout), ocr3types.ReportingPluginConfig{N: fuzzN, F: fuzzF}, offchainCfg, logger.Nop())
	require.NoError(t, err)
	if sequenced {
		p.setSequencedTSEnabled(limits.NewRangeLimiter(settings.Range[config.Timestamp]{
			Lower: config.NewTimestamp(time.UnixMilli(0)),
			Upper: config.NewTimestamp(time.UnixMilli(4_000_000_000_000)),
		}))
	}
	return p
}

type fuzzRound struct {
	sequenced bool
	// honest is how many honest oracles contribute: 3, or 2 when one is missing (2F+1 observations in total).
	honest     int
	timestamps [fuzzHonest]int64
	byzantine  []byte
}

func parseRequestIDs(ids string) []string {
	var out []string
	for _, id := range strings.Split(strings.ToValidUTF8(ids, "?"), ",") {
		if id != "" {
			out = append(out, id)
		}
	}
	return out
}

func requestsAt(ids []string, seqs [fuzzHonest]uint8, oracle int) map[string]int64 {
	reqs := map[string]int64{}
	for _, id := range ids {
		reqs[id] = int64(seqs[oracle])
	}
	return reqs
}

func legacyTimestamps(ot *pb.ObservedDonTimes) []int64 {
	return ot.Timestamps //nolint:staticcheck // SA1019: unsequenced mode still populates the deprecated slice
}

func latestTS(ot *pb.ObservedDonTimes) (latest int64, ok bool) {
	all := append(slices.Collect(maps.Values(ot.TimestampsBySequence)), legacyTimestamps(ot)...)
	if len(all) == 0 {
		return 0, false
	}
	return slices.Max(all), true
}

func earliestTS(ot *pb.ObservedDonTimes) (earliest int64, ok bool) {
	if ts := ot.EarliestTS(); ts != nil {
		return *ts, true
	}
	if legacy := legacyTimestamps(ot); len(legacy) > 0 {
		return legacy[0], true
	}
	return 0, false
}

// requireFirstRoundLiveness checks that an execution reported at sequence 0 by more than F honest oracles gets an
// entry, unless batch trimming (which keeps the lexicographically smallest IDs) can legitimately drop it.
func requireFirstRoundLiveness(t *testing.T, outcome *pb.Outcome, honestReqs []map[string]int64) {
	t.Helper()
	votes := map[string]int{}
	known := map[string]struct{}{}
	for _, reqs := range honestReqs {
		for id, seq := range reqs {
			known[id] = struct{}{}
			if seq == 0 {
				votes[id]++
			}
		}
	}
	for rank, id := range slices.Sorted(maps.Keys(known)) {
		if rank < fuzzBatchSize && votes[id] > fuzzF {
			require.Contains(t, outcome.ObservedDonTimes, id, "execution %q reported at seq 0 by %d honest oracles must get a don time", id, votes[id])
		}
	}
}

func runFuzzRound(t *testing.T, prev []byte, r fuzzRound, honestReqs [fuzzHonest]map[string]int64) []byte {
	t.Helper()
	aos := make([]types.AttributedObservation, 0, fuzzN)
	for i := range r.honest {
		raw, err := proto.MarshalOptions{Deterministic: true}.Marshal(&pb.Observation{Timestamp: r.timestamps[i], Requests: honestReqs[i]})
		require.NoError(t, err)
		aos = append(aos, types.AttributedObservation{Observation: raw, Observer: commontypes.OracleID(i)})
	}
	aos = append(aos, types.AttributedObservation{Observation: r.byzantine, Observer: fuzzHonest})

	outctx := ocr3types.OutcomeContext{PreviousOutcome: prev}
	// The honest observations are always valid and prev is empty or a prior successful outcome, so Outcome must succeed.
	first, err := newFuzzPlugin(t, r.sequenced).Outcome(t.Context(), outctx, nil, aos)
	require.NoError(t, err)
	second, err := newFuzzPlugin(t, r.sequenced).Outcome(t.Context(), outctx, nil, aos)
	require.NoError(t, err)
	require.Equal(t, first, second, "same inputs must yield identical outcome")

	prevOutcome, outcome := &pb.Outcome{}, &pb.Outcome{}
	require.NoError(t, proto.Unmarshal(prev, prevOutcome))
	require.NoError(t, proto.Unmarshal(first, outcome))

	honestTimestamps := r.timestamps[:r.honest]
	lo, hi := slices.Min(honestTimestamps), slices.Max(honestTimestamps)
	// With one Byzantine oracle the median stays in the honest range; the only other source of movement is the minimum increase.
	floor := prevOutcome.Timestamp + minIncreaseMs
	require.GreaterOrEqual(t, outcome.Timestamp, max(lo, floor), "don time must advance and not fall below the honest range")
	require.LessOrEqual(t, outcome.Timestamp, max(hi, floor), "one Byzantine timestamp must not move don time above the honest range")

	require.LessOrEqual(t, len(outcome.ObservedDonTimes), fuzzBatchSize, "outcome must respect batch size")

	honestIDs := map[string]bool{}
	for _, reqs := range honestReqs[:r.honest] {
		for id := range reqs {
			honestIDs[id] = true
		}
	}
	removalMs := defaultExecutionRemovalTime.Milliseconds()
	for id, observed := range outcome.ObservedDonTimes {
		if _, existed := prevOutcome.ObservedDonTimes[id]; !existed {
			require.True(t, honestIDs[id], "execution %q was added without an honest observer", id)
		}
		earliest, ok := earliestTS(observed)
		require.True(t, ok, "execution %q must not be kept with no timestamps", id)
		require.Greater(t, earliest+removalMs, outcome.Timestamp, "execution %q must be removed once its first don time is older than the removal time", id)
	}

	if len(prev) == 0 {
		requireFirstRoundLiveness(t, outcome, honestReqs[:r.honest])
		for id, observed := range outcome.ObservedDonTimes {
			latest, _ := latestTS(observed)
			require.Equal(t, outcome.Timestamp, latest, "execution %q created this round must carry the new don time", id)
		}
	}
	return first
}

func FuzzPlugin_OutcomeAcrossRounds(f *testing.F) {
	byz := mustMarshalObservation(f, &pb.Observation{Timestamp: 1, Requests: map[string]int64{"wf-a": 0}})
	byzOnlyHigh := mustMarshalObservation(f, &pb.Observation{Timestamp: 3_900_000_000_000, Requests: map[string]int64{"zzz": 0, "wf-a": 0}})
	const expiryGap = int32(2_000_000) // beyond the 20 minute execution removal time
	f.Add(false, false, false, int32(0), int32(5), int32(-5), int32(1000), "wf-a,wf-b", uint8(0), uint8(0), uint8(1), uint8(1), byz, []byte{})
	f.Add(true, true, false, int32(0), int32(0), int32(0), int32(1000), "wf-a", uint8(0), uint8(0), uint8(2), uint8(2), []byte{}, byz)
	f.Add(true, false, false, int32(0), int32(1), int32(2), int32(1000), "wf-a", uint8(0), uint8(0), uint8(2), uint8(2), byz, byz)
	f.Add(false, true, false, int32(0), int32(0), int32(0), int32(-1), "a,b,c,d,e", uint8(0), uint8(0), uint8(0), uint8(0), []byte{0xff, 0xff}, []byte{0x08})
	f.Add(false, false, false, int32(0), int32(0), int32(0), int32(1000), "wf-a", uint8(0), uint8(0), uint8(0), uint8(0), byzOnlyHigh, byzOnlyHigh)
	f.Add(true, true, false, int32(0), int32(0), int32(0), expiryGap, "wf-a,wf-b", uint8(0), uint8(0), uint8(0), uint8(0), byz, byz)
	f.Add(false, false, false, int32(0), int32(0), int32(0), expiryGap, "wf-a,wf-b", uint8(0), uint8(0), uint8(0), uint8(0), byz, byz)
	f.Add(false, false, true, int32(0), int32(0), int32(0), int32(1000), "wf-a", uint8(0), uint8(0), uint8(1), uint8(1), byzOnlyHigh, byz)

	f.Fuzz(func(t *testing.T, seq1, seq2, dropOneHonest bool, d0, d1, d2, gap int32, ids string, a1, b1, a2, b2 uint8, byz1, byz2 []byte) {
		honest := fuzzHonest
		if dropOneHonest {
			honest--
		}
		reqIDs := parseRequestIDs(ids)
		round1 := fuzzRound{
			sequenced:  seq1,
			honest:     honest,
			timestamps: [fuzzHonest]int64{fuzzBaseTS + int64(d0), fuzzBaseTS + int64(d1), fuzzBaseTS + int64(d2)},
			byzantine:  byz1,
		}
		seqs1 := [fuzzHonest]uint8{a1, a1, b1}
		var reqs1 [fuzzHonest]map[string]int64
		for i := range fuzzHonest {
			reqs1[i] = requestsAt(reqIDs, seqs1, i)
		}
		out1 := runFuzzRound(t, nil, round1, reqs1)

		round2 := fuzzRound{
			sequenced:  seq2,
			honest:     honest,
			timestamps: [fuzzHonest]int64{round1.timestamps[0] + int64(gap), round1.timestamps[1] + int64(gap), round1.timestamps[2] + int64(gap)},
			byzantine:  byz2,
		}
		seqs2 := [fuzzHonest]uint8{a2, a2, b2}
		var reqs2 [fuzzHonest]map[string]int64
		for i := range fuzzHonest {
			reqs2[i] = requestsAt(reqIDs, seqs2, i)
		}
		runFuzzRound(t, out1, round2, reqs2)
	})
}

func mustMarshalObservation(t testing.TB, o *pb.Observation) []byte {
	b, err := proto.Marshal(o)
	require.NoError(t, err)
	return b
}
