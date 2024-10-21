package keystore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestKeystore(t *testing.T) {
	ctx := tests.Context(t)
	stopCh := make(chan struct{})
	log := logger.Test(t)

	pluginName := "keystore-test"
	client, server := plugin.TestPluginGRPCConn(
		t,
		true,
		map[string]plugin.Plugin{
			pluginName: &testKeystorePlugin{
				log:  log,
				impl: &testKeystore{},
				brokerExt: &net.BrokerExt{
					BrokerConfig: net.BrokerConfig{
						StopCh: stopCh,
						Logger: log,
					},
				},
			},
		},
	)

	defer client.Close()
	defer server.Stop()

	keystoreClient, err := client.Dispense(pluginName)
	require.NoError(t, err)

	ks, ok := keystoreClient.(*Client)
	require.True(t, ok)

	r, err := ks.Sign(ctx, keyID, data)
	require.NoError(t, err)
	require.Equal(t, r, sign)

	r2, err := ks.SignBatch(ctx, keyID, dataList)
	require.NoError(t, err)
	require.Equal(t, r2, signBatch)

	r3, err := ks.Verify(ctx, keyID, data)
	require.NoError(t, err)
	require.Equal(t, r3, verify)

	r4, err := ks.VerifyBatch(ctx, keyID, dataList)
	require.NoError(t, err)
	require.Equal(t, r4, verifyBatch)

	r5, err := ks.ListKeys(ctx, tags)
	require.NoError(t, err)
	require.Equal(t, r5, list)

	r6, err := ks.RunUDF(ctx, udfName, keyID, data)
	require.NoError(t, err)
	require.Equal(t, r6, runUDF)

	r7, err := ks.ImportKey(ctx, keyType, data, tags)
	require.NoError(t, err)
	require.Equal(t, r7, importResponse)

	r8, err := ks.ExportKey(ctx, keyID)
	require.NoError(t, err)
	require.Equal(t, r8, export)

	r9, err := ks.CreateKey(ctx, keyType, tags)
	require.NoError(t, err)
	require.Equal(t, r9, create)

	err = ks.DeleteKey(ctx, keyID)
	require.ErrorContains(t, err, errDelete.Error())

	err = ks.AddTag(ctx, keyID, tag)
	require.ErrorContains(t, err, errAddTag.Error())

	err = ks.RemoveTag(ctx, keyID, tag)
	require.ErrorContains(t, err, errRemoveTag.Error())

	r10, err := ks.ListTags(ctx, keyID)
	require.NoError(t, err)
	require.Equal(t, r10, listTag)
}

var (
	//Inputs
	keyID    = []byte("this-is-a-keyID")
	data     = []byte("some-data")
	dataList = [][]byte{[]byte("some-data-in-a-list"), []byte("some-more-data-in-a-list")}
	tags     = []string{"tag1", "tag2"}
	tag      = "just-one-tag"
	udfName  = "i-am-a-udf-method-name"
	keyType  = "some-keyType"

	//Outputs
	sign           = []byte("signed")
	signBatch      = [][]byte{[]byte("signed1"), []byte("signed2")}
	verify         = true
	verifyBatch    = []bool{true, false}
	list           = [][]byte{[]byte("item1"), []byte("item2")}
	runUDF         = []byte("udf-response")
	importResponse = []byte("imported")
	export         = []byte("exported")
	create         = []byte("created")
	listTag        = []string{"tag1", "tag2"}
	errDelete      = errors.New("delete-err")
	errAddTag      = errors.New("add-tag-err")
	errRemoveTag   = errors.New("remove-tag-err")
)

type testKeystorePlugin struct {
	log logger.Logger
	plugin.NetRPCUnsupportedPlugin
	brokerExt *net.BrokerExt
	impl      GRPCService
}

func (r *testKeystorePlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, client *grpc.ClientConn) (any, error) {
	r.brokerExt.Broker = broker

	return NewKeystoreClient(r.brokerExt.Broker, r.brokerExt.BrokerConfig, client), nil
}

func (r *testKeystorePlugin) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	r.brokerExt.Broker = broker

	err := RegisterKeystoreServer(server, r.brokerExt.Broker, r.brokerExt.BrokerConfig, r.impl)
	if err != nil {
		return err
	}
	return nil
}

type testKeystore struct {
	services.Service
}

func checkKeyID(target []byte) error {
	if !bytes.Equal(target, keyID) {
		return fmt.Errorf("checkKeyID: expected %v but got %v", keyID, target)
	}
	return nil
}

func checkData(target []byte) error {
	if !bytes.Equal(target, data) {
		return fmt.Errorf("checkData: expected %v but got %v", data, target)
	}
	return nil
}

func checkDataList(target [][]byte) error {
	if !reflect.DeepEqual(target, dataList) {
		return fmt.Errorf("checkDataList: nexpected %v but got %v", data, target)
	}
	return nil
}

func checkTags(target []string) error {
	if !reflect.DeepEqual(target, tags) {
		return fmt.Errorf("checkTags: expected %v but got %v", tags, target)
	}
	return nil
}

func checkUdfName(target string) error {
	if target != udfName {
		return fmt.Errorf("checkUdfName: expected %v but got %v", udfName, target)
	}
	return nil
}

func checkKeyType(target string) error {
	if target != keyType {
		return fmt.Errorf("checkKeyType: expected %q but got %q", keyType, target)
	}
	return nil
}

func checkTag(target string) error {
	if target != tag {
		return fmt.Errorf("checkTag: expected %q but got %q", tag, target)
	}
	return nil
}

func (t testKeystore) Sign(ctx context.Context, _keyID []byte, _data []byte) ([]byte, error) {
	return sign, errors.Join(checkKeyID(_keyID), checkData(_data))
}

func (t testKeystore) SignBatch(ctx context.Context, _keyID []byte, _dataList [][]byte) ([][]byte, error) {
	return signBatch, errors.Join(checkKeyID(_keyID), checkDataList(_dataList))
}

func (t testKeystore) Verify(ctx context.Context, _keyID []byte, _data []byte) (bool, error) {
	return verify, errors.Join(checkKeyID(_keyID), checkData(_data))
}

func (t testKeystore) VerifyBatch(ctx context.Context, _keyID []byte, _dataList [][]byte) ([]bool, error) {
	return verifyBatch, errors.Join(checkKeyID(_keyID), checkDataList(_dataList))
}

func (t testKeystore) ListKeys(ctx context.Context, _tags []string) ([][]byte, error) {
	return list, checkTags(_tags)
}

func (t testKeystore) RunUDF(ctx context.Context, _udfName string, _keyID []byte, _data []byte) ([]byte, error) {
	return runUDF, errors.Join(checkUdfName(_udfName), checkKeyID(_keyID), checkData(_data))
}

func (t testKeystore) ImportKey(ctx context.Context, _keyType string, _data []byte, _tags []string) ([]byte, error) {
	return importResponse, errors.Join(checkKeyType(_keyType), checkData(_data), checkTags(_tags))
}

func (t testKeystore) ExportKey(ctx context.Context, _keyID []byte) ([]byte, error) {
	return export, checkKeyID(_keyID)
}

func (t testKeystore) CreateKey(ctx context.Context, _keyType string, _tags []string) ([]byte, error) {
	return create, errors.Join(checkKeyType(_keyType), checkTags(_tags))
}

func (t testKeystore) DeleteKey(ctx context.Context, _keyID []byte) error {
	return errors.Join(errDelete, checkKeyID(_keyID))
}

func (t testKeystore) AddTag(ctx context.Context, _keyID []byte, _tag string) error {
	return errors.Join(errAddTag, checkKeyID(_keyID), checkTag(_tag))
}

func (t testKeystore) RemoveTag(ctx context.Context, _keyID []byte, _tag string) error {
	return errors.Join(errRemoveTag, checkKeyID(_keyID), checkTag(_tag))
}

func (t testKeystore) ListTags(ctx context.Context, _keyID []byte) ([]string, error) {
	return listTag, checkKeyID(_keyID)
}
