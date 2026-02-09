package keystore

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

func Test_KeyStoreClient(t *testing.T) {
	ctx := t.Context()

	account := "testAccount"
	signedData := []byte("signedData")
	decryptedData := []byte("decryptedData")
	client := Client{grpc: &testGrpcClient{
		account:       account,
		signedData:    signedData,
		decryptedData: decryptedData,
	}}
	accounts, err := client.Accounts(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{account}, accounts)

	sig, err := client.Sign(ctx, account, []byte("123"))
	assert.NoError(t, err)
	assert.Equal(t, signedData, sig)

	dec, err := client.Decrypt(ctx, account, sig)
	assert.NoError(t, err)
	assert.Equal(t, decryptedData, dec)
}

func Test_KeyValueStoreServer(t *testing.T) {
	ctx := t.Context()

	account := "testAccount"
	signedData := []byte("signedData")
	decryptedData := []byte("decryptedData")
	server := Server{impl: &testKeyStore{
		account:       account,
		signedData:    signedData,
		decryptedData: decryptedData,
	}}

	accounts, err := server.Accounts(ctx, &emptypb.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, []string{account}, accounts.Accounts)

	sig, err := server.Sign(ctx, &pb.SignRequest{Account: account, Data: []byte("123")})
	assert.NoError(t, err)
	assert.Equal(t, signedData, sig.SignedData)

	dec, err := server.Decrypt(ctx, &pb.DecryptRequest{Account: account, Data: []byte("456")})
	assert.NoError(t, err)
	assert.Equal(t, decryptedData, dec.DecryptedData)
}

type testGrpcClient struct {
	account       string
	signedData    []byte
	decryptedData []byte
}

func (t *testGrpcClient) Accounts(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.AccountsReply, error) {
	return &pb.AccountsReply{
		Accounts: []string{t.account},
	}, nil
}

func (t *testGrpcClient) Sign(ctx context.Context, req *pb.SignRequest, opts ...grpc.CallOption) (*pb.SignReply, error) {
	return &pb.SignReply{
		SignedData: t.signedData,
	}, nil
}

func (t *testGrpcClient) Decrypt(ctx context.Context, req *pb.DecryptRequest, opts ...grpc.CallOption) (*pb.DecryptReply, error) {
	return &pb.DecryptReply{
		DecryptedData: t.decryptedData,
	}, nil
}

type testKeyStore struct {
	core.UnimplementedKeystore
	account       string
	signedData    []byte
	decryptedData []byte
}

func (t *testKeyStore) Accounts(ctx context.Context) ([]string, error) {
	return []string{t.account}, nil
}

func (t *testKeyStore) Sign(ctx context.Context, account string, data []byte) ([]byte, error) {
	return t.signedData, nil
}

func (t *testKeyStore) Decrypt(ctx context.Context, account string, data []byte) ([]byte, error) {
	return t.decryptedData, nil
}
