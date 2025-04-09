package custmsg

import (
	"bytes"
	"context"
	"fmt"
	"maps"

	"github.com/golang/protobuf/jsonpb"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/pb"
)

type MessageEmitter interface {
	// Emit sends a message to the labeler's destination.
	Emit(context.Context, string) error

	// WithMapLabels sets the labels for the message to be emitted.  Labels are cumulative.
	WithMapLabels(map[string]string) MessageEmitter

	// With adds multiple key-value pairs to the emission.
	With(keyValues ...string) MessageEmitter

	// Labels returns a view of the current labels.
	Labels() map[string]string
}

type Labeler struct {
	labels map[string]string
}

func NewLabeler() Labeler {
	return Labeler{labels: make(map[string]string)}
}

// WithMapLabels adds multiple key-value pairs to the CustomMessageLabeler for transmission
// With SendLogAsCustomMessage
func (l Labeler) WithMapLabels(labels map[string]string) MessageEmitter {
	newCustomMessageLabeler := NewLabeler()

	// Copy existing labels from the current agent
	maps.Copy(newCustomMessageLabeler.labels, l.labels)

	// Add new key-value pairs
	maps.Copy(newCustomMessageLabeler.labels, labels)

	return newCustomMessageLabeler
}

// With adds multiple key-value pairs to the CustomMessageLabeler for transmission With SendLogAsCustomMessage
func (l Labeler) With(keyValues ...string) MessageEmitter {
	newCustomMessageLabeler := NewLabeler()

	if len(keyValues)%2 != 0 {
		// If an odd number of key-value arguments is passed, return the original CustomMessageLabeler unchanged
		return l
	}

	// Copy existing labels from the current agent
	maps.Copy(newCustomMessageLabeler.labels, l.labels)

	// Add new key-value pairs
	for i := 0; i < len(keyValues); i += 2 {
		key := keyValues[i]
		value := keyValues[i+1]
		newCustomMessageLabeler.labels[key] = value
	}

	return newCustomMessageLabeler
}

func (l Labeler) Emit(ctx context.Context, msg string) error {
	return sendLogAsCustomMessageW(ctx, msg, l.labels)
}

func (l Labeler) Labels() map[string]string {
	copied := make(map[string]string, len(l.labels))

	maps.Copy(copied, l.labels)

	return copied
}

// SendLogAsCustomMessage emits a BaseMessage With msg and labels as data.
// any key in labels that is not part of orderedLabelKeys will not be transmitted
func (l Labeler) SendLogAsCustomMessage(ctx context.Context, msg string) error {
	return sendLogAsCustomMessageW(ctx, msg, l.labels)
}

func sendLogAsCustomMessageW(ctx context.Context, msg string, labels map[string]string) error {
	// TODO un-comment after INFOPLAT-1386
	// cast to map[string]any
	//newLabels := map[string]any{}
	//for k, v := range labels {
	//	newLabels[k] = v
	//}

	//m, err := values.NewMap(newLabels)
	//if err != nil {
	//	return fmt.Errorf("could not wrap labels to map: %w", err)
	//}
	// Define a custom protobuf payload to emit
	payload := &pb.BaseMessage{
		Msg:    msg,
		Labels: labels,
	}

	// We marshal to JSON so we can read these messages in Loki
	marshaler := &jsonpb.Marshaler{}
	var buf bytes.Buffer
	err := marshaler.Marshal(&buf, payload)
	if err != nil {
		return fmt.Errorf("sending custom message failed to marshal to JSON: %w", err)
	}

	err = beholder.GetEmitter().Emit(ctx, buf.Bytes(),
		"beholder_data_schema", "/beholder-base-message/versions/1", // required
		"beholder_domain", "platform", // required
		"beholder_entity", "BaseMessage", // required
	)
	if err != nil {
		return fmt.Errorf("sending custom message failed on emit: %w", err)
	}

	return nil
}
