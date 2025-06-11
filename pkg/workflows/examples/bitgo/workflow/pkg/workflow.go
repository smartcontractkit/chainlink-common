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

type EvmConfig struct {
	TokenAddress  string
	PorAddress    string
	ChainSelector uint32
	GasLimit      uint64
}

type Config struct {
	PublicKey string
	Schedule  string
	Url       string
	Evms      []EvmConfig
}

func InitWorkflow(wcx *sdk.WorkflowContext[*Config]) (sdk.Workflows[*Config], error) {
	config := wcx.Config
	workflows := sdk.Workflows[*Config]{
		sdk.On(
			cron.Cron{}.Trigger(&cron.Config{Schedule: config.Schedule}),
			onCronTrigger,
		),
		sdk.On(
			http.Trigger{}.Request(),
			onHttpTrigger,
		),
	}

	for _, evmConfig := range config.Evms {
		address, err := hex.DecodeString(evmConfig.TokenAddress[2:])
		if err != nil {
			return nil, fmt.Errorf("failed to decode token address %s: %w", evmConfig.TokenAddress, err)
		}
		evmClient := &evmcappb.Client{ChainSelector: evmConfig.ChainSelector}
		workflow := sdk.On(
			bindings.NewIReserveManager(bindings.ContractInputs{EVM: evmClient, Address: address, Options: &bindings.ContractOptions{}}).RequestReserveUpdateTrigger(evmcappb.ConfidenceLevel_FINALIZED),
			onEvmTrigger,
		)
		workflows = append(workflows, workflow)
	}

	return workflows, nil
}

type httpTrigger struct {
	Reason string `json:"reason"`
}

func onEvmTrigger(wcx *sdk.WorkflowContext[*Config], runtime sdk.Runtime, log *bindings.RequestReserveUpdateLog) (*ReserveInfo, error) {
	wcx.Logger = wcx.Logger.With("trigger", "evm").With("selector", log.ChainSelector)

	wcx.Logger = wcx.Logger.With("request id", log.RequestId.String())
	return doPor(wcx, runtime, time.Now())
}

func onHttpTrigger(wcx *sdk.WorkflowContext[*Config], runtime sdk.Runtime, request *http.TriggerRequest) (*ReserveInfo, error) {
	trigger := &httpTrigger{}
	if err := json.Unmarshal(request.Body, trigger); err != nil {
		wcx.Logger.Error("failed to unmarshal http trigger payload", "err", err)
		return nil, err
	}

	wcx.Logger = wcx.Logger.With("trigger", "http").With("reason", trigger.Reason)
	return doPor(wcx, runtime, time.Now())
}

func onCronTrigger(wcx *sdk.WorkflowContext[*Config], runtime sdk.Runtime, trigger *cron.Payload) (*ReserveInfo, error) {
	wcx.Logger = wcx.Logger.With("trigger", "cron")
	return doPor(wcx, runtime, trigger.ScheduledExecutionTime.AsTime())
}

func doPor(wcx *sdk.WorkflowContext[*Config], runtime sdk.Runtime, runTime time.Time) (*ReserveInfo, error) {
	logger := wcx.Logger
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

	if reserveInfo.LastUpdated.Before(runTime.Add(-time.Hour * 24)) {
		logger.Warn("reserve time is too old", "time", reserveInfo.LastUpdated)
		return nil, errors.New("reserved time is too old")
	}

	totalSupply, err := getTotalSupply(wcx, runtime)

	totalReserveScaled := reserveInfo.TotalReserve.Mul(decimal.NewFromUint64(10e18)).BigInt()

	if err = updateReserve(wcx, runtime, totalSupply, totalReserveScaled); err != nil {
		return nil, err
	}

	return reserveInfo, nil
}

func getTotalSupply(wcx *sdk.WorkflowContext[*Config], runtime sdk.Runtime) (*big.Int, error) {
	// Fetch supply from all EVMs in parallel
	evms := wcx.Config.Evms
	logger := wcx.Logger
	supplyPromises := make([]sdk.Promise[*big.Int], len(evms))
	for i, evmConfig := range evms {
		evmClient := &evmcappb.Client{ChainSelector: evmConfig.ChainSelector}

		address, err := hexToBytes(evmConfig.TokenAddress)
		if err != nil {
			logger.Error("failed to decode token address", "address", evmConfig.TokenAddress, "err", err)
			return nil, fmt.Errorf("failed to decode token address %s: %w", evmConfig.TokenAddress, err)
		}
		token := bindings.NewIERC20(bindings.ContractInputs{EVM: evmClient, Address: address})
		evmTotalSupplyPromise := token.Methods.TotalSupply.Call(runtime, nil)
		supplyPromises[i] = evmTotalSupplyPromise
	}

	// We can add sdk.AwaitAll that takes []sdk.Promise[T] and returns ([]T, error)
	totalSupply := big.NewInt(0)
	for i, promise := range supplyPromises {
		supply, err := promise.Await()
		if err != nil {
			selector := evms[i].ChainSelector
			logger.Error("Could not read from contract", "contract_chain", selector, "err", err.Error())
			return nil, err
		}

		totalSupply = totalSupply.Add(totalSupply, supply)
	}

	return totalSupply, nil
}

func updateReserve(wcx *sdk.WorkflowContext[*Config], runtime sdk.Runtime, totalSupply, totalReserveScaled *big.Int) error {
	evms := wcx.Config.Evms

	reportWrites := make([]sdk.Promise[*evm.WriteReportReply], len(evms))
	for i, evmConfig := range evms {
		evmClient := &evmcappb.Client{ChainSelector: evmConfig.ChainSelector}

		// Address must be parsable or the workflow would fail to initialize the trigger.
		address, _ := hexToBytes(evmConfig.PorAddress)
		reserveManager := bindings.NewIReserveManager(bindings.ContractInputs{EVM: evmClient, Address: address, Options: &bindings.ContractOptions{
			GasConfig: &evm.GasConfig{
				GasLimit: evmConfig.GasLimit,
			},
		}})
		reportWrites[i] = reserveManager.WriteReportUpdateReserves(runtime, bindings.UpdateReservesStruct{
			TotalMinted:  totalSupply,
			TotalReserve: totalReserveScaled,
		}, nil)
	}

	var errs []error
	for i, promise := range reportWrites {
		writeReportReply, err := promise.Await()
		if err == nil {
			wcx.Logger.Debug("update reserve write report reply", "chain_selector", evms[i].ChainSelector, "tx hash", writeReportReply.TxHash)
		} else {
			selector := evms[i].ChainSelector
			wcx.Logger.Error("Could not write to contract", "contract_chain", selector, "err", err.Error())
			errs = append(errs, fmt.Errorf("failed to write report for chain %d: %w", selector, err))
			continue
		}
	}

	return errors.Join(errs...)
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

	reserveInfo := &ReserveInfo{}
	if err = json.Unmarshal([]byte(porResponse.Data), reserveInfo); err != nil {
		return nil, err
	}

	return reserveInfo, nil
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

func hexToBytes(hexStr string) ([]byte, error) {
	return hex.DecodeString(hexStr[2:])
}

type CommonReport struct {
	RawReport     []byte
	ReportContext []byte
	Signatures    [][]byte
	ID            []byte
}
