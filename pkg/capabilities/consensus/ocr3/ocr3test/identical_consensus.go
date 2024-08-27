package ocr3test

import (
	"encoding/base64"
	"errors"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/testutils"
)

// IdenticalConsensus registers a new capability mock with the runner
// If a step is explicitly mocked by another mock, that will take priority over this one for that step.
func IdenticalConsensus[T any](runner *testutils.Runner) *IdenticalConsensusMock[T] {
	consensus := &IdenticalConsensusMock[T]{
		Mock: testutils.MockCapability[ConsensusInput[T], ocr3.SignedReport]("offchain_reporting@1.0.0", identicalConsensus[T]),
	}
	runner.MockCapability(consensus.ID(), nil, consensus)
	return consensus
}

// IdenticalConsensusForStep registers a new capability mock with the runner, but only for a given step.
// if another mock was registered for the same capability without a step, this mock will take priority for that step.
func IdenticalConsensusForStep[T any](runner *testutils.Runner, step string) *IdenticalConsensusMock[T] {
	consensus := &IdenticalConsensusMock[T]{
		Mock: testutils.MockCapability[ConsensusInput[T], ocr3.SignedReport]("offchain_reporting@1.0.0", identicalConsensus[T]),
	}
	runner.MockCapability(consensus.ID(), &step, consensus)
	return consensus
}

func identicalConsensus[T any](inputs ConsensusInput[T]) (ocr3.SignedReport, error) {
	if len(inputs.Observations) == 0 {
		return ocr3.SignedReport{}, errors.New("no observations were made")
	} else if len(inputs.Observations) > 1 {
		return ocr3.SignedReport{}, errors.New("more than one observation was made, but this mock isn't set up to support that")
	}

	wrapped, err := values.Wrap(inputs.Observations[0])
	if err != nil {
		return ocr3.SignedReport{}, err
	}

	bytes, err := proto.Marshal(values.Proto(wrapped))
	if err != nil {
		return ocr3.SignedReport{}, err
	}

	return ocr3.SignedReport{
		Context:    "this is a test",
		ID:         "12",
		Report:     base64.StdEncoding.EncodeToString(bytes),
		Signatures: []string{"sig1", "sig2", "sig3", "sig4"},
	}, nil
}

// IdenticalConsensusMock is a mock of the identical consensus capability
// Note that the mock ignores the encoding and it's config, only validating that they conform to the schema
// The mock will encode the single value using values.Value and signatures will be random bytes
type IdenticalConsensusMock[T any] struct {
	*testutils.Mock[ConsensusInput[T], ocr3.SignedReport]
}

var _ testutils.ConsensusMock = &IdenticalConsensusMock[struct{}]{}

func (c *IdenticalConsensusMock[T]) SingleToManyObservations(input values.Value) (values.Value, error) {
	tmp := singleConsensusInput[T]{}
	if err := input.UnwrapTo(&tmp); err != nil {
		return nil, err
	}

	return values.Wrap(ConsensusInput[T]{Observations: []T{tmp.Observation}})
}

func (c *IdenticalConsensusMock[T]) GetStepDecoded(ref string) testutils.StepResults[ConsensusInput[T], T] {
	step := c.GetStep(ref)
	var t T
	if step.WasRun && step.Error == nil {
		bytes, _ := base64.StdEncoding.DecodeString(step.Output.Report)
		wrapped := &pb.Value{}

		// safe because we marshalled it in the mock step
		_ = proto.Unmarshal(bytes, wrapped)
		mv, _ := values.FromProto(wrapped)
		_ = mv.UnwrapTo(&t)
	}

	return testutils.StepResults[ConsensusInput[T], T]{WasRun: step.WasRun, Input: step.Input, Output: t, Error: step.Error}
}
