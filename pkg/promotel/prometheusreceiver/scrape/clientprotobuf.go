package scrape

import (
	"bytes"
	"encoding/binary"

	"github.com/gogo/protobuf/proto"

	// Intentionally using client model to simulate client in tests.
	dto "github.com/prometheus/client_model/go"
)

// Write a MetricFamily into a protobuf.
// This function is intended for testing scraping by providing protobuf serialized input.
func MetricFamilyToProtobuf(metricFamily *dto.MetricFamily) ([]byte, error) {
	buffer := &bytes.Buffer{}
	err := AddMetricFamilyToProtobuf(buffer, metricFamily)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Append a MetricFamily protobuf representation to a buffer.
// This function is intended for testing scraping by providing protobuf serialized input.
func AddMetricFamilyToProtobuf(buffer *bytes.Buffer, metricFamily *dto.MetricFamily) error {
	protoBuf, err := proto.Marshal(metricFamily)
	if err != nil {
		return err
	}

	varintBuf := make([]byte, binary.MaxVarintLen32)
	varintLength := binary.PutUvarint(varintBuf, uint64(len(protoBuf)))

	_, err = buffer.Write(varintBuf[:varintLength])
	if err != nil {
		return err
	}
	_, err = buffer.Write(protoBuf)
	return err
}
