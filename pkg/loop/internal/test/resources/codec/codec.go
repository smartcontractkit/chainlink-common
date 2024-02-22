package codec_test

import (
	"context"
	"errors"
)

type StaticCodec struct{}

func (c StaticCodec) GetMaxEncodingSize(ctx context.Context, n int, itemType string) (int, error) {
	return 0, errors.New("not used for these test")
}

func (c StaticCodec) GetMaxDecodingSize(ctx context.Context, n int, itemType string) (int, error) {
	return 0, errors.New("not used for these test")
}

func (c StaticCodec) Encode(ctx context.Context, item any, itemType string) ([]byte, error) {
	return nil, errors.New("not used for these test")
}

func (c StaticCodec) Decode(ctx context.Context, raw []byte, into any, itemType string) error {
	return errors.New("not used for these test")
}
