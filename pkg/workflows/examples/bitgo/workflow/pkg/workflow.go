package pkg

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
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

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/cron"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/node/http"
	evmcappb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/examples/bitgo/workflow/pkg/bindings"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

//go:embed solc/bin/IERC20.abi
var Erc20Abi string

//go:embed solc/bin/IReserveManager.abi
var ReserveManagerAbi string

const TotalSupplyMethod = "totalSupply"
const UpdateReservesMethod = "updateReserves"

type Config struct {
	EvmTokenAddress  string
	EvmPorAddress    string
	PublicKey        string
	Schedule         string
	Url              string
	EvmChainSelector uint
	GasLimit         uint64
}

func InitWorkflow(wcx *sdk.WorkflowContext[*Config]) (sdk.Workflows[*Config], error) {
	config := wcx.Config
	return sdk.Workflows[*Config]{
		sdk.On(
			cron.Cron{}.Trigger(&cron.Config{Schedule: config.Schedule}),
			onCronTrigger,
		),
	}, nil
}

func onCronTrigger(wcx *sdk.WorkflowContext[*Config], runtime sdk.Runtime, trigger *cron.CronTrigger) error {
	logger := wcx.Logger
	config := wcx.Config
	client := &http.Client{}
	reserveInfo, err := http.ConsensusFetch(
		wcx,
		runtime,
		client,
		fetchPor,
		sdk.ConsensusAggregationFromTags[*ReserveInfo]()).
		Await()

	if err != nil {
		return err
	}

	if time.UnixMilli(reserveInfo.LastUpdated).Before(time.Unix(trigger.ScheduledExecutionTime, 0).Add(-time.Hour * 24)) {
		logger.Warn("reserve time is too old", "time", reserveInfo.LastUpdated)
		return errors.New("reserved time is too old")
	}

	totalSupply := big.NewInt(0)

	evmClient := evmcappb.Client{}

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
		return err
	}

	totalSupply = totalSupply.Add(totalSupply, evmSupply)
	// TODO add other chains

	totalReserveScaled := reserveInfo.TotalReserve.Mul(decimal.NewFromUint64(10e18)).BigInt()

	writeReportReplyPromise := reserveManager.Structs.UpdateReserves.WriteReport(runtime, bindings.UpdateReservesStruct{
		TotalMinted:  *totalSupply,
		TotalReserve: *totalReserveScaled,
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

	return errors.Join(writeErrors...)
}

func fetchPor(wcx *sdk.WorkflowContext[*Config], fetcher *http.Fetcher) (*ReserveInfo, error) {
	config := wcx.Config

	request := &http.HttpFetchRequest{Url: config.Url}
	response, err := fetcher.Fetch(request).Await()
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
	bytes, _ := hex.DecodeString(hexStr)
	return bytes
}

type CommonReport struct {
	RawReport     []byte
	ReportContext []byte
	Signatures    [][]byte
	ID            []byte
}

// TODO we need to define and implement this function
func GenerateReport(targetChainID uint, evmSupplyCallData []byte) CommonReport {
	panic("unimplemented")
}
