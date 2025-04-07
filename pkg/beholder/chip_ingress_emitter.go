package beholder

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
)

type ChipIngressEmitter struct {
	client chipingress.ChipIngressClient
}

func NewChipIngressEmitter(client chipingress.ChipIngressClient) (Emitter, error) {

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

	event, err := chipingress.NewEvent(sourceDomain, entityType, body)
	if err != nil {
		return err
	}

	_, err = c.client.Publish(ctx, event)
	if err != nil {
		return err
	}

	return nil
}

// ExtractSourceAndType extracts source domain and entity from the attributes
func ExtractSourceAndType(attrKVs ...any) (string, string, error) {

	var sourceDomain string
	var entityType string

	for i := 0; i < len(attrKVs)-1; i += 2 {

		key, ok := attrKVs[i].(string)
		if !ok {
			continue
		}

		// Retrieve source and type using either ChIP or legacy attribute names, prioritizing source/type
		if key == "source" || (key == "beholder_domain" && sourceDomain == "") {
			if val, ok := attrKVs[i+1].(string); ok {
				sourceDomain = val
			}
		}
		if key == "type" || (key == "beholder_entity" && entityType == "") {
			if val, ok := attrKVs[i+1].(string); ok {
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
