package beholder

import (
	"context"
	"fmt"
	"maps"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"google.golang.org/protobuf/proto"
)

type ChipIngressEmitter struct {
	client chipingress.Client
}

var _ Emitter = (*ChipIngressEmitter)(nil)

func NewChipIngressEmitter(client chipingress.Client) (Emitter, error) {

	if client == nil {
		return nil, fmt.Errorf("chip ingress client is nil")
	}

	return &ChipIngressEmitter{client: client}, nil
}

func (c *ChipIngressEmitter) Close() error {
	return c.client.Close()
}

func (c *ChipIngressEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	_, err := c.BatchEmit(ctx, []Message{
		NewMessage(body, attrKVs...),
	})
	return err
}

func (c *ChipIngressEmitter) BatchEmit(ctx context.Context, messages []Message, options ...BatchEmitOption) ([]*chipingress.PublishResult, error) {
	emitOpts := DefaultBatchEmitOptions
	for _, opt := range options {
		opt(&emitOpts)
	}

	events := make([]chipingress.CloudEvent, len(messages))
	for i, msg := range messages {
		sourceDomain, entityType, err := ExtractSourceAndType(msg.Attrs)
		if err != nil {
			return nil, err
		}

		event, err := chipingress.NewEvent(sourceDomain, entityType, msg.Body, msg.Attrs)
		if err != nil {
			return nil, err
		}

		events[i] = event
	}

	eventPb, err := chipingress.EventsToBatch(events)
	if err != nil {
		return nil, fmt.Errorf("failed to convert event to proto: %w", err)
	}

	eventPb.Options = &chipingress.PublishOptions{
		AllOrNothing: proto.Bool(emitOpts.AllOrNothing),
	}

	response, err := c.client.PublishBatch(ctx, eventPb)
	if err != nil {
		return nil, err
	}

	if response == nil {
		return nil, nil
	}

	return response.Results, nil
}

// ExtractSourceAndType extracts source domain and entity from the attributes
func ExtractSourceAndType(attributes Attributes) (string, string, error) {
	var sourceDomain string
	var entityType string

	for key, value := range attributes {

		// Retrieve source and type using either ChIP or legacy attribute names, prioritizing source/type
		if key == "source" || (key == AttrKeyDomain && sourceDomain == "") {
			if val, ok := value.(string); ok {
				sourceDomain = val
			}
		}
		if key == "type" || (key == AttrKeyEntity && entityType == "") {
			if val, ok := value.(string); ok {
				entityType = val
			}
		}
	}

	if sourceDomain == "" {
		return "", "", fmt.Errorf("source/beholder_domain not found in provided key/value attributes")
	}

	if entityType == "" {
		return "", "", fmt.Errorf("type/beholder_entity not found in provided key/value attributes")
	}

	return sourceDomain, entityType, nil
}

func ExtractAttributes(attrKVs ...any) map[string]any {
	attributes := newAttributes(attrKVs...)

	attributesMap := make(map[string]any)
	maps.Copy(attributesMap, attributes)

	return attributesMap
}
