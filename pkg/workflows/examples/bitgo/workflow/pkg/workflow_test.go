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
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	evmmock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm/capabilitymock"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/http"
	httpmock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/http/actionmock"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/triggers/cron"
	cronmock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/triggers/cron/cron_triggermock"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/examples/bitgo/workflow/pkg"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/examples/bitgo/workflow/pkg/bindings"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/examples/bitgo/workflow/pkg/bindings/bindingsmock"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	publicKeyPath  = "testutils/fixtures/public.pem"
	privateKeyPath = "testutils/fixtures/private.pem"
	signedJSONPath = "testutils/fixtures/signed.json"
)

var testTime = time.Date(2025, 2, 3, 20, 37, 2, 552574000, time.UTC)

const totalReserve = "11.56"

const anyEvmChainSelector = uint32(123)

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
		EvmTokenAddress:  "0x1234567890abcdef1234567890abcdef12345678",
		EvmPorAddress:    "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		PublicKey:        string(pubKeyBytes),
		Schedule:         "0 0 * * *",
		Url:              "https://reserves.gousd.com/por.json",
		EvmChainSelector: anyEvmChainSelector,
	}

	// New pattern, hide registry in a map from testingT to registry.
	cronMock, err := cronmock.NewCronCapability(t)
	require.NoError(t, err)
	cronMock.Trigger = func(ctx context.Context, input *cron.Config) (*cron.Payload, error) {
		assert.Equal(t, config.Schedule, input.Schedule)
		triggerTime := testTime.Truncate(24 * time.Hour).Add(time.Hour * 24)
		return &cron.Payload{ScheduledExecutionTime: triggerTime.Format(time.RFC3339Nano)}, nil
	}

	httpMock, err := httpmock.NewClientCapability(t)
	require.NoError(t, err)
	httpMock.SendRequest = func(ctx context.Context, input *http.Request) (*http.Response, error) {
		assert.Equal(t, http.Method_GET, input.Method)
		assert.Equal(t, config.Url, input.Url)
		assert.Empty(t, input.Body)
		return &http.Response{Body: payload}, nil
	}

	numEvmTokens := new(big.Int).Mul(big.NewInt(103), big.NewInt(1e16))
	totalTokens := numEvmTokens
	evmMock, err := evmmock.NewClientCapability(t, config.EvmChainSelector)
	require.NoError(t, err)

	erc20Mock := bindingsmock.NewIERC20Mock(common.HexToAddress(config.EvmTokenAddress), evmMock)
	erc20Mock.TotalSupply = func() (*big.Int, error) { return numEvmTokens, nil }

	reserveManager := bindingsmock.NewIReserverManagerMock(common.HexToAddress(config.EvmPorAddress), evmMock)
	reserveManager.UpdateReserves = func(reserves *bindings.UpdateReservesStruct) error {
		assert.Equal(t, totalTokens, reserves.TotalMinted)
		reserve, err := decimal.NewFromString(totalReserve)
		require.NoError(t, err)
		reserve = reserve.Mul(decimal.New(10, 18))
		assert.Equal(t, reserve.BigInt(), reserves.TotalReserve)
		assert.Equal(t, totalTokens, reserves.TotalMinted)
		return nil
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
