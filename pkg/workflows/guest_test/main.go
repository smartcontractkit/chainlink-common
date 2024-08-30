package main

import (
	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/yaml"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/ocr3cap"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/targets/chainwriter"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/streams"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

type Config struct {
	Workflow    sdk.NewWorkflowParams
	Streams     *streams.TriggerConfig
	Ocr         *ocr3cap.DataFeedsConsensusConfig
	ChainWriter *chainwriter.TargetConfig
	TargetChain string
}

func main() {
	conf, err := UnmarshalYaml[Config]([]byte{})
	if err != nil {
		panic(err)
	}

	workflow := sdk.NewWorkflowSpecFactory(conf.Workflow)
	streamsTrigger := conf.Streams.New(workflow)
	consensus := conf.Ocr.New(workflow, "ccip_feeds", ocr3cap.DataFeedsConsensusInput{
		Observations: sdk.ListOf[streams.Feed](streamsTrigger)},
	)

	conf.ChainWriter.New(workflow, conf.TargetChain, chainwriter.TargetInput{SignedReport: consensus})
	w, _ := values.Wrap(workflow)
	proto.Marshal(values.Proto(w))
}

func UnmarshalYaml[T any](raw []byte) (*T, error) {
	var v T
	err := yaml.Unmarshal(raw, &v)
	return &v, err
}
