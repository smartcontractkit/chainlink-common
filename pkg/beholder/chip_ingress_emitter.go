package beholder

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
)

type ChipIngressEmitter struct {
	client chipingress.Client
}

func NewChipIngressEmitter(client chipingress.Client) (Emitter, error) {

	if client == nil {
		return nil, fmt.Errorf("chip ingress client is nil")
	}

	return &ChipIngressEmitter{client: client}, nil
}

func (c *ChipIngressEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {

	sourceDomain, entityType, err := ExtractSourceAndType(attrKVs...)
	if err != nil {
		return err
	}

	event, err := chipingress.NewEvent(sourceDomain, entityType, body, newAttributes(attrKVs...))
	if err != nil {
		return err
	}

	eventPb, err := chipingress.EventToProto(event)
	if err != nil {
		return fmt.Errorf("failed to convert event to proto: %w", err)
	}

	_, err = c.client.Publish(ctx, eventPb)
	if err != nil {
		return err
	}

	return nil
}

// ExtractSourceAndType extracts source domain and entity from the attributes
func ExtractSourceAndType(attrKVs ...any) (string, string, error) {

	attributes := newAttributes(attrKVs...)

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
	for key, value := range attributes {
		attributesMap[key] = value
	}

	return attributesMap
}
