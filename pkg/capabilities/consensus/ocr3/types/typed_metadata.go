package types

/*
import (
	"google.golang.org/protobuf/proto"

	datafeeds "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/datafeeds"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

type MetadataType int

const (
	MetadataTypeUnknown MetadataType = iota
	MetadataTypeDataFeeds
	MetadataTypeLLO
	MetadataTypeReduce
)

func PopulateTypedMetadata(outcome *AggregationOutcome, mdType MetadataType) {
	if outcome == nil || len(outcome.Metadata) == 0 {
		return
	}

	switch mdType {
	case MetadataTypeDataFeeds:
		var md datafeeds.DataFeedsOutcomeMetadata
		if err := proto.Unmarshal(outcome.Metadata, &md); err == nil {
			outcome.TypedMetadata = &AggregationOutcome_DataFeedsMetadata{
				DataFeedsMetadata: &md,
			}
		}
	case MetadataTypeLLO:
		var md datafeeds.LLOOutcomeMetadata
		if err := proto.Unmarshal(outcome.Metadata, &md); err == nil {
			outcome.TypedMetadata = &AggregationOutcome_LloMetadata{
				LloMetadata: &md,
			}
		}
	case MetadataTypeReduce:
		pb := &values.Map{}
		if err := proto.Unmarshal(outcome.Metadata, pb); err == nil {
			outcome.TypedMetadata = &AggregationOutcome_ReduceMetadata{
				ReduceMetadata: pb,
			}
		}
	}
}

func ExtractTypedMetadata(outcome *AggregationOutcome) (proto.Message, MetadataType) {
	if outcome == nil {
		return nil, MetadataTypeUnknown
	}

	switch tm := outcome.TypedMetadata.(type) {
	case *AggregationOutcome_DataFeedsMetadata:
		if tm.DataFeedsMetadata != nil {
			return tm.DataFeedsMetadata, MetadataTypeDataFeeds
		}
	case *AggregationOutcome_LloMetadata:
		if tm.LloMetadata != nil {
			return tm.LloMetadata, MetadataTypeLLO
		}
	case *AggregationOutcome_ReduceMetadata:
		if tm.ReduceMetadata != nil {
			return tm.ReduceMetadata, MetadataTypeReduce
		}
	}

	return nil, MetadataTypeUnknown
}
*/
