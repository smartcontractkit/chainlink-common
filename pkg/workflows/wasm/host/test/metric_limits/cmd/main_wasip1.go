package main

import (
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
	wfpb "github.com/smartcontractkit/chainlink-protos/workflows/go/v2"
)

func main() {
	rawsdk.SwitchModes(int32(sdkpb.Mode_MODE_DON))
	request := rawsdk.GetRequest()

	emitMetric("valid_counter", 1, wfpb.UserMetricType_USER_METRIC_TYPE_COUNTER, map[string]string{"k": "v"})

	emitMetric("this_name_is_way_too_long", 2, wfpb.UserMetricType_USER_METRIC_TYPE_COUNTER, nil)

	emitMetric("valid_gauge", 42, wfpb.UserMetricType_USER_METRIC_TYPE_GAUGE, nil)

	emitMetric("third_one", 3, wfpb.UserMetricType_USER_METRIC_TYPE_COUNTER, nil)

	emitMetric("fourth_one", 4, wfpb.UserMetricType_USER_METRIC_TYPE_COUNTER, nil)

	rawsdk.SendResponse(request.Config)
}

func emitMetric(name string, value float64, metricType wfpb.UserMetricType, labels map[string]string) {
	m := &wfpb.WorkflowUserMetric{
		Name:   name,
		Value:  value,
		Type:   metricType,
		Labels: labels,
	}
	b := rawsdk.Must(proto.Marshal(m))
	rawsdk.EmitMetric(rawsdk.BufferToPointerLen(b))
}
