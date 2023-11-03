package types

import "context"

type MapDecoder interface {
	DecodeSingle(ctx context.Context, raw []byte, itemType string) (map[string]any, error)
	DecodeMany(ctx context.Context, raw []byte, itemType string) ([]map[string]any, error)
}
