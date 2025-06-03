package http

import "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"

type Fetcher struct {
	rt     sdk.NodeRuntime
	client *Client
}

func (f *Fetcher) Fetch(input *HttpFetchRequest) sdk.Promise[*HttpFetchResponse] {
	return f.client.Fetch(f.rt, input)
}

func ConsensusFetch[T, C any](
	wcx *sdk.WorkflowContext[C],
	runtime sdk.Runtime,
	client *Client,
	fn func(wcx *sdk.WorkflowContext[C], f *Fetcher) (T, error),
	ca sdk.ConsensusAggregation[T]) sdk.Promise[T] {
	return sdk.RunInNodeMode[C, T](wcx, runtime, func(wcx *sdk.WorkflowContext[C], nodeRuntime sdk.NodeRuntime) (T, error) {
		fetcher := &Fetcher{rt: nodeRuntime, client: client}
		return fn(wcx, fetcher)
	}, ca)
}
