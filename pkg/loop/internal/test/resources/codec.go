package resources_test

import (
	"context"
	"errors"
)

var CodecImpl = staticCodec{
	staticCodecConfig: staticCodecConfig{
		n:        3,
		itemType: "itemType",
		maxSize:  37,
	},
}

type staticCodecConfig struct {
	n        int
	itemType string
	maxSize  int
}

type staticCodec struct {
	staticCodecConfig
}

func (c staticCodec) GetMaxEncodingSize(ctx context.Context, n int, itemType string) (int, error) {
	return c.maxSize, nil
}

func (c staticCodec) GetMaxDecodingSize(ctx context.Context, n int, itemType string) (int, error) {
	return c.maxSize, nil
}

func (c staticCodec) Encode(ctx context.Context, item any, itemType string) ([]byte, error) {
	return nil, errors.New("staticCoded.Encode not used for these test")
}

func (c staticCodec) Decode(ctx context.Context, raw []byte, into any, itemType string) error {
	return errors.New("staticCodec.Decode not used for these test")
}
