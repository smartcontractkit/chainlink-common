package keyvalue

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
)

func Test_KeyValueStoreClient(t *testing.T) {
	ctx := t.Context()
	// Setup
	client := Client{grpc: &testGrpcClient{store: make(map[string][]byte)}}
	key := "key"
	insertedVal := "aval"

	err := client.Store(ctx, key, []byte(insertedVal))
	assert.NoError(t, err)

	retrievedVal, err := client.Get(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, insertedVal, string(retrievedVal))
}

func Test_KeyValueStoreServer(t *testing.T) {
	ctx := t.Context()
	// Setup
	server := Server{impl: &testKeyValueStore{store: make(map[string][]byte)}}

	_, err := server.StoreKeyValue(ctx, &pb.StoreKeyValueRequest{Key: "key", Value: []byte(`{"A":"a","B":1}`)})
	assert.NoError(t, err)
	resp, err := server.GetValueForKey(ctx, &pb.GetValueForKeyRequest{Key: "key"})
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"A":"a","B":1}`), resp.Value)
}

func Test_KeyValueStoreClient_PruneExpiredEntries(t *testing.T) {
	ctx := context.Background()

	client := Client{grpc: &testGrpcClient{store: make(map[string][]byte)}}

	// Store some test data
	err := client.Store(ctx, "key1", []byte("value1"))
	assert.NoError(t, err)
	err = client.Store(ctx, "key2", []byte("value2"))
	assert.NoError(t, err)

	// Prune entries
	numPruned, err := client.PruneExpiredEntries(ctx, time.Hour)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), numPruned)
}

func Test_KeyValueStoreServer_PruneExpiredEntries(t *testing.T) {
	ctx := context.Background()

	server := Server{impl: &testKeyValueStore{store: make(map[string][]byte)}}

	// Store some test data through the server
	_, err := server.StoreKeyValue(ctx, &pb.StoreKeyValueRequest{Key: "key1", Value: []byte("value1")})
	assert.NoError(t, err)
	_, err = server.StoreKeyValue(ctx, &pb.StoreKeyValueRequest{Key: "key2", Value: []byte("value2")})
	assert.NoError(t, err)

	// Prune entries
	resp, err := server.PruneExpiredEntries(ctx, &pb.PruneExpiredEntriesRequest{
		MaxAge: durationpb.New(time.Hour),
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), resp.NumPruned)
}

type testGrpcClient struct {
	store map[string][]byte
}

func (t *testGrpcClient) StoreKeyValue(ctx context.Context, in *pb.StoreKeyValueRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	t.store[in.Key] = in.Value
	return &emptypb.Empty{}, nil
}

func (t *testGrpcClient) GetValueForKey(ctx context.Context, in *pb.GetValueForKeyRequest, opts ...grpc.CallOption) (*pb.GetValueForKeyResponse, error) {
	return &pb.GetValueForKeyResponse{Value: t.store[in.Key]}, nil
}

func (t *testGrpcClient) PruneExpiredEntries(ctx context.Context, in *pb.PruneExpiredEntriesRequest, opts ...grpc.CallOption) (*pb.PruneExpiredEntriesResponse, error) {
	numPruned := 0
	for k := range t.store {
		delete(t.store, k)
		numPruned++
	}
	return &pb.PruneExpiredEntriesResponse{NumPruned: int64(numPruned)}, nil
}

type testKeyValueStore struct {
	store map[string][]byte
}

func (t *testKeyValueStore) Store(ctx context.Context, key string, val []byte) error {
	t.store[key] = val
	return nil
}

func (t *testKeyValueStore) Get(ctx context.Context, key string) ([]byte, error) {
	return t.store[key], nil
}

func (t *testKeyValueStore) PruneExpiredEntries(ctx context.Context, maxAge time.Duration) (int64, error) {
	numPruned := 0
	for k := range t.store {
		delete(t.store, k)
		numPruned++
	}
	return int64(numPruned), nil
}
