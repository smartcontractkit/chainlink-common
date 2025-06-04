package pkg_test

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/cron"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/cron/cronmock"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/crosschain"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/node/http"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/node/http/httpmock"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/examples/bitgo/workflow/pkg"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/testutils"
)

const (
	publicKeyPath  = "testutils/fixtures/public.pem"
	privateKeyPath = "testutils/fixtures/private.pem"
	signedJSONPath = "testutils/fixtures/signed.json"
)

var testTime = time.Date(2025, 2, 3, 20, 37, 2, 552574000, time.UTC)

const totalReserve = "11.56"

func TestWorkflow_HappyPath(t *testing.T) {
	err := ensureSignedJSON()
	require.NoError(t, err)

	// Load public key
	pubKeyBytes, err := os.ReadFile(publicKeyPath)
	require.NoError(t, err)

	// Load signed.json
	payload, err := os.ReadFile(signedJSONPath)
	require.NoError(t, err)

	config := &pkg.Config{
		EvmTokenAddress: "0x1234567890abcdef1234567890abcdef12345678",
		EvmPorAddress:   "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		PublicKey:       string(pubKeyBytes),
		Schedule:        "0 0 * * *",
		Url:             "https://reserves.gousd.com/por.json",
	}

	// New pattern, hide registry in a map from testingT to registry.
	cronMock, err := cronmock.NewCronCapability(t)
	require.NoError(t, err)
	cronMock.Trigger = func(ctx context.Context, input *cron.Config) (*cron.CronTrigger, error) {
		assert.Equal(t, config.Schedule, input.Schedule)
		triggerTime := testTime.Truncate(24 * time.Hour).Add(time.Hour * 24).Unix()
		return &cron.CronTrigger{ScheduledExecutionTime: triggerTime}, nil
	}

	httpMock, err := httpmock.NewClientCapability(t)
	require.NoError(t, err)
	httpMock.Fetch = func(ctx context.Context, input *http.HttpFetchRequest) (*http.HttpFetchResponse, error) {
		assert.Equal(t, http.HttpMethod_GET, input.Method)
		assert.Equal(t, config.Url, input.Url)
		assert.Empty(t, input.Body)
		return &http.HttpFetchResponse{Body: payload}, nil
	}

	numEvmTokens := new(big.Int).Mul(big.NewInt(103), big.NewInt(1e16))
	totalTokens := numEvmTokens
	evmMock, err := evmmock.NewClientCapability(t)
	require.NoError(t, err)
	evmMock.ReadMethod = func(ctx context.Context, input *evm.ReadMethodRequest) (*crosschain.ByteArray, error) {
		assert.Equal(t, config.EvmTokenAddress, input.Address)
		assert.Equal(t, evm.ConfidenceLevel_FINALITY, input.ConfidenceLevel)
		erc20, err := abi.JSON(strings.NewReader(pkg.Erc20Abi))
		require.NoError(t, err)

		method := erc20.Methods[pkg.TotalSupplyMethod]
		assert.Equal(t, method.ID, input.Calldata)

		response, err := erc20.Methods[pkg.TotalSupplyMethod].Outputs.Pack(numEvmTokens)
		require.NoError(t, err)
		return &crosschain.ByteArray{Value: response}, nil
	}
	evmMock.SubmitTransaction = func(ctx context.Context, input *evm.SubmitTransactionRequest) (*evm.TxID, error) {
		assert.Equal(t, config.EvmPorAddress, input.ToAddress)
		reserveManager, err := abi.JSON(strings.NewReader(pkg.ReserveManagerAbi))
		require.NoError(t, err)
		method := reserveManager.Methods[pkg.UpdateReservesMethod]
		callId := input.Calldata[0:len(method.ID)]
		assert.Equal(t, method.ID, callId)

		argData := input.Calldata[len(method.ID):]
		args := map[string]any{}
		assert.NoError(t, method.Inputs.UnpackIntoMap(args, argData))

		assert.Len(t, args, 2)
		reserve, err := decimal.NewFromString(totalReserve)
		require.NoError(t, err)
		reserve = reserve.Mul(decimal.New(10, 18))
		assert.Equal(t, reserve.BigInt(), args["totalReserve"])
		assert.Equal(t, totalTokens, args["totalMinted"])

		return &evm.TxID{Value: "fake transaction"}, nil
	}

	runner := testutils.NewRunner(t, config)

	runner.Run(pkg.InitWorkflow)

	ok, _, err := runner.Result()
	require.True(t, ok)
	require.NoError(t, err)
}

func ensureSignedJSON() error {
	if fileExists(publicKeyPath) && fileExists(privateKeyPath) && fileExists(signedJSONPath) {
		return nil
	}

	_ = os.Remove(signedJSONPath)

	// Generate RSA key pair
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Save private key
	privOut, err := os.Create(privateKeyPath)
	if err != nil {
		return err
	}
	defer privOut.Close()
	privBytes := x509.MarshalPKCS1PrivateKey(key)
	if err = pem.Encode(privOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}); err != nil {
		return err
	}

	// Save public key
	pubOut, err := os.Create(publicKeyPath)
	if err != nil {
		return err
	}
	defer pubOut.Close()
	pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return err
	}
	if err = pem.Encode(pubOut, &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}); err != nil {
		return err
	}

	reserve, err := decimal.NewFromString(totalReserve)
	if err != nil {
		return err
	}

	dataStruct := pkg.RawReserveInfo{
		LastUpdated:  testTime,
		TotalReserve: reserve,
	}
	dataBytes, err := json.Marshal(dataStruct)
	if err != nil {
		return err
	}

	// Sign data
	hash := sha256.Sum256(dataBytes)
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
	if err != nil {
		return err
	}
	sigEncoded := base64.RawURLEncoding.EncodeToString(sig)

	signed := pkg.PorResponse{
		Data:          string(dataBytes),
		DataSignature: sigEncoded,
		Ripcord:       false,
	}

	signedJSON, err := json.MarshalIndent(signed, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(signedJSONPath, signedJSON, 0644)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}