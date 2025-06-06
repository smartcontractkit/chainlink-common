package http

import "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"

type SendRequester struct {
	rt     sdk.NodeRuntime
	client *Client
}

func (f *SendRequester) SendRequest(input *Request) sdk.Promise[*Response] {
	return f.client.SendRequest(f.rt, input)
}

func ConsensusSendRequest[T, C any](
	wcx *sdk.WorkflowContext[C],
	runtime sdk.Runtime,
	client *Client,
	fn func(wcx *sdk.WorkflowContext[C], sendRequester *SendRequester) (T, error),
	ca sdk.ConsensusAggregation[T]) sdk.Promise[T] {
	return sdk.RunInNodeMode[C, T](wcx, runtime, func(wcx *sdk.WorkflowContext[C], nodeRuntime sdk.NodeRuntime) (T, error) {
		return fn(wcx, &SendRequester{rt: nodeRuntime, client: client})
	}, ca)
}
