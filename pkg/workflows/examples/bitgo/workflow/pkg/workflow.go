package pkg

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/shopspring/decimal"
	"github.com/smartcontractkit/chainlink-common/pkg/chains/evm"

	evmcappb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/http"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/triggers/cron"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/examples/bitgo/workflow/pkg/bindings"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

type Config struct {
	EvmTokenAddress  string
	EvmPorAddress    string
	PublicKey        string
	Schedule         string
	Url              string
	EvmChainSelector uint32
	GasLimit         uint64
}

func InitWorkflow(wcx *sdk.WorkflowContext[*Config]) (sdk.Workflows[*Config], error) {
	config := wcx.Config
	return sdk.Workflows[*Config]{
		sdk.On(
			cron.Cron{}.Trigger(&cron.Config{Schedule: config.Schedule}),
			onCronTrigger,
		),
		sdk.OnValue(
			http.Trigger{}.Request(),
			onHttpTrigger,
		),
	}, nil
}

type httpTrigger struct {
	Reason string `json:"reason"`
}

func onHttpTrigger(wcx *sdk.WorkflowContext[*Config], runtime sdk.Runtime, payload *http.TriggerRequest) (*ReserveInfo, error) {
	trigger := &httpTrigger{}
	if err := json.Unmarshal(payload.Body, trigger); err != nil {
		wcx.Logger.Error("failed to unmarshal http trigger payload", "err", err)
		return nil, err
	}

	wcx.Logger = wcx.Logger.With("trigger", "http").With("reason", trigger.Reason)
	return doPor(wcx, runtime, time.Now())
}

func onCronTrigger(wcx *sdk.WorkflowContext[*Config], runtime sdk.Runtime, trigger *cron.Payload) error {
	wcx.Logger = wcx.Logger.With("trigger", "cron")
	scheduledExecution, err := time.Parse(time.RFC3339Nano, trigger.ScheduledExecutionTime)
	if err != nil {
		wcx.Logger.Error("failed to parse scheduled execution time", "err", err)
		return err
	}

	_, err = doPor(wcx, runtime, scheduledExecution)
	return err
}

func doPor(wcx *sdk.WorkflowContext[*Config], runtime sdk.Runtime, runTime time.Time) (*ReserveInfo, error) {
	logger := wcx.Logger
	config := wcx.Config
	client := &http.Client{}
	reserveInfo, err := http.ConsensusSendRequest(
		wcx,
		runtime,
		client,
		fetchPor,
		sdk.ConsensusAggregationFromTags[*ReserveInfo]()).
		Await()

	if err != nil {
		logger.Warn("error getting por value", "err", err.Error())
		return nil, err
	}

	if time.UnixMilli(reserveInfo.LastUpdated).Before(runTime.Add(-time.Hour * 24)) {
		logger.Warn("reserve time is too old", "time", reserveInfo.LastUpdated)
		return nil, errors.New("reserved time is too old")
	}

	totalSupply := big.NewInt(0)

	evmClient := &evmcappb.Client{ChainSelector: config.EvmChainSelector}

	token := bindings.NewIERC20(bindings.ContractInputs{EVM: evmClient, Address: hexToBytes(config.EvmTokenAddress)})
	reserveManager := bindings.NewIReserveManager(bindings.ContractInputs{EVM: evmClient, Address: hexToBytes(config.EvmPorAddress), Options: &bindings.ContractOptions{
		GasConfig: &evm.GasConfig{
			GasLimit: config.GasLimit,
		},
	}})

	evmTotalSupplyPromise := token.Methods.TotalSupply.Call(runtime, nil)
	evmSupply, err := evmTotalSupplyPromise.Await()
	if err != nil {
		// TODO specify which EVM
		logger.Error("Could not read from evm", "err", err.Error())
		return nil, err
	}

	totalSupply = totalSupply.Add(totalSupply, evmSupply)
	// TODO add other chains

	totalReserveScaled := reserveInfo.TotalReserve.Mul(decimal.NewFromUint64(10e18)).BigInt()

	writeReportReplyPromise := reserveManager.Structs.UpdateReserves.WriteReport(runtime, bindings.UpdateReservesStruct{
		TotalMinted:  totalSupply,
		TotalReserve: totalReserveScaled,
	}, nil)

	writeReportReply, err := writeReportReplyPromise.Await()

	var writeErrors []error
	if err == nil {
		txHash := writeReportReply.TxHash
		logger.Debug("Submitted transaction", "tx hash", txHash)
	} else {
		logger.Error("failed to submit transaction", "err", err)
		writeErrors = append(writeErrors, err)
	}

	return nil, errors.Join(writeErrors...)
}

func fetchPor(wcx *sdk.WorkflowContext[*Config], requester *http.SendRequester) (*ReserveInfo, error) {
	config := wcx.Config

	request := &http.Request{Url: config.Url}
	response, err := requester.SendRequest(request).Await()
	if err != nil {
		return nil, err
	}

	porResponse := &PorResponse{}
	if err = json.Unmarshal(response.Body, porResponse); err != nil {
		return nil, err
	}

	err = verifySignature(porResponse, config.PublicKey)
	if err != nil {
		return nil, err
	}

	if porResponse.Ripcord {
		return nil, errors.New("ripcord is true")
	}

	rawReserve := &RawReserveInfo{}
	if err = json.Unmarshal([]byte(porResponse.Data), rawReserve); err != nil {
		return nil, err
	}

	return &ReserveInfo{
		LastUpdated:  rawReserve.LastUpdated.UnixMilli(),
		TotalReserve: rawReserve.TotalReserve,
	}, nil
}

func verifySignature(porResponse *PorResponse, publicKey string) error {
	// Decode the signature
	rawSig, err := base64.RawURLEncoding.DecodeString(porResponse.DataSignature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// Parse the PEM public key
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil || block.Type != "PUBLIC KEY" {
		return fmt.Errorf("invalid PEM block")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	pubKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("not an RSA public key")
	}

	// Hash the payload
	hasher := crypto.SHA256.New()
	hasher.Write([]byte(porResponse.Data))
	digest := hasher.Sum(nil)

	// Verify
	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, digest, rawSig); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}
	return nil
}

// HexToBytes converts a hex string to a byte array.
// It returns the byte array and any error encountered.
func hexToBytes(hexStr string) []byte {
	bytes, _ := hex.DecodeString(hexStr[2:])
	return bytes
}

type CommonReport struct {
	RawReport     []byte
	ReportContext []byte
	Signatures    [][]byte
	ID            []byte
}
