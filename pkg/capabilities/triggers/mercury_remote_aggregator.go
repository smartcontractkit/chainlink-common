package triggers

import (
	"errors"
	"sort"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/datastreams"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

type mercuryRemoteAggregator struct {
	codec                 datastreams.ReportCodec
	allowedSigners        [][]byte
	minRequiredSignatures int
	previousLatestReports map[datastreams.FeedID]datastreams.FeedReport
	lggr                  logger.Logger
}

// This aggregator is used by TriggerSubscriber to aggregate trigger events from multiple remote nodes.
// NOTE: Once Mercury supports parallel composition (and thus guarantee identical sets of reports),
// this will be replaced by the default MODE aggregator.
func NewMercuryRemoteAggregator(codec datastreams.ReportCodec, allowedSigners [][]byte, minRequiredSignatures int, lggr logger.Logger) *mercuryRemoteAggregator {
	if allowedSigners == nil {
		allowedSigners = [][]byte{}
	}
	return &mercuryRemoteAggregator{
		codec:                 codec,
		allowedSigners:        allowedSigners,
		minRequiredSignatures: minRequiredSignatures,
		previousLatestReports: make(map[datastreams.FeedID]datastreams.FeedReport),
		lggr:                  lggr,
	}
}

func (a *mercuryRemoteAggregator) Aggregate(triggerEventID string, responses [][]byte) (capabilities.TriggerResponse, error) {
	latestReports := make(map[datastreams.FeedID]datastreams.FeedReport)
	latestGlobalTs := int64(0) // to be used as the timestamp of the combined trigger event
	for _, response := range responses {
		unmarshaled, err := pb.UnmarshalTriggerResponse(response)
		if err != nil {
			a.lggr.Errorw("could not unmarshal one of capability responses (faulty sender?)", "error", err)
			continue
		}
		feedReports, err := a.codec.Unwrap(unmarshaled.Event.Outputs)
		if err != nil {
			a.lggr.Errorw("could not unwrap one of capability responses", "error", err)
			continue
		}
		// save latest valid report for each feed ID
		for _, report := range feedReports {
			latestReport, ok := latestReports[datastreams.FeedID(report.FeedID)]
			if !ok {
				// on first occurrence of a feed ID, check if we saw it in any of the past events
				latestReport, ok = a.previousLatestReports[datastreams.FeedID(report.FeedID)]
				if ok {
					latestReports[datastreams.FeedID(report.FeedID)] = latestReport
					if latestReport.ObservationTimestamp > latestGlobalTs {
						latestGlobalTs = report.ObservationTimestamp
					}
				}
			}
			if !ok || report.ObservationTimestamp > latestReport.ObservationTimestamp {
				// lazy signature validation
				if err := a.codec.Validate(report, a.allowedSigners, a.minRequiredSignatures); err != nil {
					a.lggr.Errorw("invalid report", "error", err)
				} else {
					latestReports[datastreams.FeedID(report.FeedID)] = report
					a.previousLatestReports[datastreams.FeedID(report.FeedID)] = report
					if report.ObservationTimestamp > latestGlobalTs {
						latestGlobalTs = report.ObservationTimestamp
					}
				}
			}
		}
	}
	if len(latestReports) == 0 {
		return capabilities.TriggerResponse{}, errors.New("no valid reports found")
	}
	reportList := []datastreams.FeedReport{}
	allIDs := []string{}
	for _, report := range latestReports {
		allIDs = append(allIDs, report.FeedID)
	}
	sort.Strings(allIDs)
	for _, feedID := range allIDs {
		reportList = append(reportList, latestReports[datastreams.FeedID(feedID)])
	}
	meta := datastreams.SignersMetadata{
		Signers:               a.allowedSigners,
		MinRequiredSignatures: a.minRequiredSignatures,
	}
	return wrapReports(reportList, triggerEventID, latestGlobalTs, meta)
}
