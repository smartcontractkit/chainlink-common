package beholder

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"google.golang.org/grpc/credentials/insecure"
)

type ChipIngressEmitter struct {
	client chipingress.ChipIngressClient
}

func NewChipIngressEmitter(client chipingress.ChipIngressClient) Emitter {

	if client == nil {
		panic("chip ingress emitter client cannot be nil")
	}

	return &ChipIngressEmitter{client: client}
}

func NewChipIngressClient(cfg Config) (chipingress.ChipIngressClient, error) {

	if cfg.ChipIngressEmitterGRPCEndpoint == "" {
		return nil, fmt.Errorf("missing chip ingress emitter gRPC endpoint")
	}

	// TODO: add support for csa auth signing interceptor
	// We should add csa signed headers, that will be authenticated on the server-side
	client, err := chipingress.NewChipIngressClient(
		cfg.ChipIngressEmitterGRPCEndpoint,
		chipingress.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
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

		if key == "beholder_domain" {
			if val, ok := attrKVs[i+1].(string); ok {
				sourceDomain = val
			}
		}
		if key == "beholder_entity" {
			if val, ok := attrKVs[i+1].(string); ok {
				entityType = val
			}
		}
	}

	if sourceDomain == "" {
		return "", "", fmt.Errorf("beholder_domain not found in provided key/value attributes")
	}

	if entityType == "" {
		return "", "", fmt.Errorf("beholder_entity not found in provided key/value attributes")
	}

	return sourceDomain, entityType, nil
}
