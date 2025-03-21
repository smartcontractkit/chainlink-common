package custmsg

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/pb"
)

type MessageEmitter interface {
	// Emit sends a message to the labeler's destination.
	Emit(context.Context, any) error

	// WithMapLabels sets the labels for the message to be emitted.  Labels are cumulative.
	WithMapLabels(map[string]string) MessageEmitter

	// With adds multiple key-value pairs to the emission.
	With(keyValues ...string) MessageEmitter

	// Labels returns a view of the current labels.
	Labels() map[string]string
}

type ProtoDetail struct {
	Schema string
	Domain string
	Entity string
}

// ProtoMessage is intended to be a pure function that provides a message and details required
// for publishing to beholder.
type ProtoMessage interface {
	BeholderMessage() (proto.Message, ProtoDetail)
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

func (l Labeler) Emit(ctx context.Context, msg any) error {
	switch typed := msg.(type) {
	case string:
		return sendLogAsStringMessageW(ctx, typed, l.labels)
	default:
		protoMsg, ok := msg.(ProtoMessage)
		if !ok {
			// TODO: can default to JSON encoding instead of returning an error
			return errors.New("must be a proto message")
		}

		custMsg, desc := protoMsg.BeholderMessage()

		return sendLogAsCustomMessageW(ctx, desc, custMsg, l.labels)
	}
}

func (l Labeler) Labels() map[string]string {
	copied := make(map[string]string, len(l.labels))

	maps.Copy(copied, l.labels)

	return copied
}

// SendLogAsCustomMessage emits a BaseMessage With msg and labels as data.
// any key in labels that is not part of orderedLabelKeys will not be transmitted
func (l Labeler) SendLogAsCustomMessage(ctx context.Context, msg string) error {
	return sendLogAsStringMessageW(ctx, msg, l.labels)
}

func sendLogAsStringMessageW(ctx context.Context, msg string, labels map[string]string) error {
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
	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		return fmt.Errorf("sending custom message failed to marshal protobuf: %w", err)
	}

	err = beholder.GetEmitter().Emit(ctx, payloadBytes,
		"beholder_data_schema", "/beholder-base-message/versions/1", // required
		"beholder_domain", "platform", // required
		"beholder_entity", "BaseMessage", // required
	)
	if err != nil {
		return fmt.Errorf("sending custom message failed on emit: %w", err)
	}

	return nil
}

func sendLogAsCustomMessageW(ctx context.Context, desc ProtoDetail, msg proto.Message, labels map[string]string) error {
	payloadBytes, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("sending custom message failed to marshal protobuf: %w", err)
	}

	kvs := []any{
		"beholder_data_schema", desc.Schema, // required
		"beholder_domain", desc.Domain, // required
		"beholder_entity", desc.Entity, // required
	}

	for key, value := range labels {
		kvs = append(kvs, []any{key, value}...)
	}

	err = beholder.GetEmitter().Emit(ctx, payloadBytes, kvs...)
	if err != nil {
		return fmt.Errorf("sending custom message failed on emit: %w", err)
	}

	return nil
}
