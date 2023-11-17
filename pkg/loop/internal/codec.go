package internal

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ types.Codec = (*codecClient)(nil)

type codecClient struct {
	*brokerExt
	grpc pb.CodecClient
}

func (c *codecClient) Encode(ctx context.Context, item any, itemType string) ([]byte, error) {
	versionedParams, err := encodeVersionedBytes(item, ParamsCurrentEncodingVersion)
	if err != nil {
		return nil, err
	}

	reply, err := c.grpc.GetEncoding(ctx, &pb.GetEncodingRequest{
		Params:   versionedParams,
		ItemType: itemType,
	})

	if err != nil {
		return nil, unwrapClientError(err)
	}

	return reply.RetVal, nil
}

func (c *codecClient) Decode(ctx context.Context, raw []byte, into any, itemType string) error {
	request := &pb.GetDecodingRequest{
		Encoded:  raw,
		ItemType: itemType,
	}
	resp, err := c.grpc.GetDecoding(ctx, request)
	if err != nil {
		return unwrapClientError(err)
	}

	return decodeVersionedBytes(into, resp.RetVal)
}

func (c *codecClient) GetMaxEncodingSize(ctx context.Context, n int, itemType string) (int, error) {
	res, err := c.grpc.GetMaxSize(ctx, &pb.GetMaxSizeRequest{N: int32(n), ItemType: itemType, ForEncoding: true})
	if err != nil {
		return 0, unwrapClientError(err)
	}

	return int(res.SizeInBytes), nil
}

func (c *codecClient) GetMaxDecodingSize(ctx context.Context, n int, itemType string) (int, error) {
	res, err := c.grpc.GetMaxSize(ctx, &pb.GetMaxSizeRequest{N: int32(n), ItemType: itemType, ForEncoding: false})
	if err != nil {
		return 0, unwrapClientError(err)
	}

	return int(res.SizeInBytes), nil
}

var _ pb.CodecServer = (*codecServer)(nil)

type codecServer struct {
	pb.UnimplementedCodecServer
	impl types.Codec
}

func (c *codecServer) GetEncoding(ctx context.Context, req *pb.GetEncodingRequest) (*pb.GetEncodingResponse, error) {
	encodedType, err := getEncodedType(req.ItemType, c.impl, true)
	if err != nil {
		return nil, err
	}

	if err = decodeVersionedBytes(encodedType, req.Params); err != nil {
		return nil, err
	}

	encoded, err := c.impl.Encode(ctx, encodedType, req.ItemType)
	return &pb.GetEncodingResponse{RetVal: encoded}, err
}

func (c *codecServer) GetDecoding(ctx context.Context, req *pb.GetDecodingRequest) (*pb.GetDecodingResponse, error) {
	encodedType, err := getEncodedType(req.ItemType, c.impl, false)
	if err != nil {
		return nil, err
	}

	err = c.impl.Decode(ctx, req.Encoded, encodedType, req.ItemType)
	if err != nil {
		return nil, err
	}

	versionBytes, err := encodeVersionedBytes(encodedType, RetvalCurrentEncodingVersion)
	return &pb.GetDecodingResponse{RetVal: versionBytes}, err
}

func (c *codecServer) GetMaxSize(ctx context.Context, req *pb.GetMaxSizeRequest) (*pb.GetMaxSizeResponse, error) {
	var sizeFn func(context.Context, int, string) (int, error)
	if req.ForEncoding {
		sizeFn = c.impl.GetMaxEncodingSize
	} else {
		sizeFn = c.impl.GetMaxDecodingSize
	}

	maxSize, err := sizeFn(ctx, int(req.N), req.ItemType)
	if err != nil {
		return nil, err
	}
	return &pb.GetMaxSizeResponse{SizeInBytes: int32(maxSize)}, nil
}
