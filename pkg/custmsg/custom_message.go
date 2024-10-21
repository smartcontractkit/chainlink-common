package custmsg

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

type Labeler interface {
	WithMapLabels(labels map[string]string) Labeler
	With(keyValues ...string) Labeler
	Emit(msg string) error
	Labels() map[string]string
}

type labeler struct {
	labels map[string]string
}

func NewLabeler() labeler {
	return labeler{labels: make(map[string]string)}
}

// WithMapLabels adds multiple key-value pairs to the CustomMessageLabeler for transmission
// With SendLogAsCustomMessage
func (c labeler) WithMapLabels(labels map[string]string) labeler {
	newCustomMessageLabeler := NewLabeler()

	// Copy existing labels from the current agent
	for k, v := range c.labels {
		newCustomMessageLabeler.labels[k] = v
	}

	// Add new key-value pairs
	for k, v := range labels {
		newCustomMessageLabeler.labels[k] = v
	}

	return newCustomMessageLabeler
}

// With adds multiple key-value pairs to the CustomMessageLabeler for transmission With SendLogAsCustomMessage
func (c labeler) With(keyValues ...string) labeler {
	newCustomMessageLabeler := NewLabeler()

	if len(keyValues)%2 != 0 {
		// If an odd number of key-value arguments is passed, return the original CustomMessageLabeler unchanged
		return c
	}

	// Copy existing labels from the current agent
	for k, v := range c.labels {
		newCustomMessageLabeler.labels[k] = v
	}

	// Add new key-value pairs
	for i := 0; i < len(keyValues); i += 2 {
		key := keyValues[i]
		value := keyValues[i+1]
		newCustomMessageLabeler.labels[key] = value
	}

	return newCustomMessageLabeler
}

func (c labeler) Emit(msg string) error {
	return sendLogAsCustomMessageW(msg, c.labels)
}

func (c labeler) Labels() map[string]string {
	copied := make(map[string]string, len(c.labels))
	for k, v := range c.labels {
		copied[k] = v
	}
	return copied
}

// SendLogAsCustomMessage emits a BaseMessage With msg and labels as data.
// any key in labels that is not part of orderedLabelKeys will not be transmitted
func (c labeler) SendLogAsCustomMessage(msg string) error {
	return sendLogAsCustomMessageW(msg, c.labels)
}

func sendLogAsCustomMessageW(msg string, labels map[string]string) error {
	// cast to map[string]any
	newLabels := map[string]any{}
	for k, v := range labels {
		newLabels[k] = v
	}

	m, err := values.NewMap(newLabels)
	if err != nil {
		return fmt.Errorf("could not wrap labels to map: %w", err)
	}

	// Define a custom protobuf payload to emit
	payload := &pb.BaseMessage{
		Msg:    msg,
		Labels: values.ProtoMap(m),
	}
	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		return fmt.Errorf("sending custom message failed to marshal protobuf: %w", err)
	}

	err = beholder.GetEmitter().Emit(context.Background(), payloadBytes,
		"beholder_data_schema", "/beholder-base-message/versions/1", // required
		"beholder_data_type", "custom_message",
	)
	if err != nil {
		return fmt.Errorf("sending custom message failed on emit: %w", err)
	}

	return nil
}
