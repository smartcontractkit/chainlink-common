package types

import "context"

type MapDecoder interface {
	DecodeSingle(ctx context.Context, raw []byte, itemType string) (map[string]any, error)
	DecodeMany(ctx context.Context, raw []byte, itemType string) ([]map[string]any, error)
	// GetMaxDecodingSize returns the max size in bytes if n elements are supplied for all top level dynamically sized elements.
	// If no elements are dynamically sized, the returned value will be the same for all n.
	// If there are multiple levels of dynamically sized elements, or itemType cannot be found,
	// InvalidTypeError will be returned.
	GetMaxDecodingSize(ctx context.Context, n int, itemType string) (int, error)
}
