package contractreader

import (
	"context"

	"google.golang.org/grpc"

	codecpb "github.com/smartcontractkit/chainlink-common/pkg/internal/codec"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ types.Codec = (*CodecClient)(nil)

// NewCodecTestClient is a test client for [types.Codec]
// internal users should instantiate a client directly and set all private fields.
func NewCodecTestClient(conn *grpc.ClientConn) types.Codec {
	return &CodecClient{grpc: codecpb.NewCodecClient(conn)}
}

type CodecClientOpt func(*CodecClient)

type CodecClient struct {
	*net.BrokerExt
	grpc       codecpb.CodecClient
	encodeWith codecpb.EncodingVersion
}

func NewCodecClient(b *net.BrokerExt, cc grpc.ClientConnInterface, opts ...CodecClientOpt) *CodecClient {
	client := &CodecClient{
		BrokerExt:  b,
		grpc:       codecpb.NewCodecClient(cc),
		encodeWith: codecpb.DefaultEncodingVersion,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func WithCodecClientEncoding(version codecpb.EncodingVersion) CodecClientOpt {
	return func(client *CodecClient) {
		client.encodeWith = version
	}
}

func (c *CodecClient) Encode(ctx context.Context, item any, itemType string) ([]byte, error) {
	versionedParams, err := codecpb.EncodeVersionedBytes(item, c.encodeWith)
	if err != nil {
		return nil, err
	}

	reply, err := c.grpc.GetEncoding(ctx, &codecpb.GetEncodingRequest{
		Params:   versionedParams,
		ItemType: itemType,
	})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return reply.RetVal, nil
}

func (c *CodecClient) Decode(ctx context.Context, raw []byte, into any, itemType string) error {
	request := &codecpb.GetDecodingRequest{
		Encoded:             raw,
		ItemType:            itemType,
		WireEncodingVersion: c.encodeWith.Uint32(),
	}
	resp, err := c.grpc.GetDecoding(ctx, request)
	if err != nil {
		return net.WrapRPCErr(err)
	}

	return codecpb.DecodeVersionedBytes(into, resp.RetVal)
}

func (c *CodecClient) GetMaxEncodingSize(ctx context.Context, n int, itemType string) (int, error) {
	//nolint: gosec // G115
	res, err := c.grpc.GetMaxSize(ctx, &codecpb.GetMaxSizeRequest{N: int32(n), ItemType: itemType, ForEncoding: true})
	if err != nil {
		return 0, net.WrapRPCErr(err)
	}

	return int(res.SizeInBytes), nil
}

func (c *CodecClient) GetMaxDecodingSize(ctx context.Context, n int, itemType string) (int, error) {
	//nolint: gosec // G115
	res, err := c.grpc.GetMaxSize(ctx, &codecpb.GetMaxSizeRequest{N: int32(n), ItemType: itemType, ForEncoding: false})
	if err != nil {
		return 0, net.WrapRPCErr(err)
	}

	return int(res.SizeInBytes), nil
}

var _ codecpb.CodecServer = (*CodecServer)(nil)

type CodecServer struct {
	codecpb.UnimplementedCodecServer
	impl types.Codec
}

func NewCodecServer(impl types.Codec) codecpb.CodecServer {
	return &CodecServer{impl: impl}
}

func (c *CodecServer) GetEncoding(ctx context.Context, req *codecpb.GetEncodingRequest) (*codecpb.GetEncodingResponse, error) {
	encodedType, err := getEncodedType(req.ItemType, c.impl, true)
	if err != nil {
		return nil, err
	}

	if err = codecpb.DecodeVersionedBytes(encodedType, req.Params); err != nil {
		return nil, err
	}

	encoded, err := c.impl.Encode(ctx, encodedType, req.ItemType)
	return &codecpb.GetEncodingResponse{RetVal: encoded}, err
}

func (c *CodecServer) GetDecoding(ctx context.Context, req *codecpb.GetDecodingRequest) (*codecpb.GetDecodingResponse, error) {
	encodedType, err := getEncodedType(req.ItemType, c.impl, false)
	if err != nil {
		return nil, err
	}

	err = c.impl.Decode(ctx, req.Encoded, encodedType, req.ItemType)
	if err != nil {
		return nil, err
	}

	versionBytes, err := codecpb.EncodeVersionedBytes(encodedType, codecpb.EncodingVersion(req.WireEncodingVersion))

	return &codecpb.GetDecodingResponse{RetVal: versionBytes}, err
}

func (c *CodecServer) GetMaxSize(ctx context.Context, req *codecpb.GetMaxSizeRequest) (*codecpb.GetMaxSizeResponse, error) {
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
	//nolint: gosec // G115
	return &codecpb.GetMaxSizeResponse{SizeInBytes: int32(maxSize)}, nil
}

func getEncodedType(itemType string, possibleTypeProvider any, forEncoding bool) (any, error) {
	if tp, ok := possibleTypeProvider.(types.TypeProvider); ok {
		return tp.CreateType(itemType, forEncoding)
	}

	return &map[string]any{}, nil
}
