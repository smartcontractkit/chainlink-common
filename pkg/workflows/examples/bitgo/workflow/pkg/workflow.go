package pkg

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/shopspring/decimal"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/cron"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/node/http"
	chaincommonpb "github.com/smartcontractkit/chainlink-common/pkg/loop/chain-common"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"

	evm "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm/chain-service"
	evmpb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm/chain-service"
)

//go:embed solc/bin/IERC20.abi
var Erc20Abi string

//go:embed solc/bin/IReserveManager.abi
var ReserveManagerAbi string

const TotalSupplyMethod = "totalSupply"
const UpdateReservesMethod = "updateReserves"

type Config struct {
	EvmTokenAddress []byte
	EvmPorAddress   []byte
	EvmPorUpdateGasLimit uint64
	PublicKey       string
	Schedule        string
	Url             string
}

func Workflow(runner sdk.DonRunner) {
	logger := slog.Default()
	config := &Config{}
	if err := json.Unmarshal(runner.Config(), config); err != nil {
		logger.Error("error unmarshalling config", "err", err)
		return
	}

	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewEmptyDonHandler(
				cron.Cron{}.Trigger(&cron.Config{Schedule: config.Schedule}),
				onCronTrigger,
			),
		},
	})
}

func onCronTrigger(runtime sdk.DonRuntime, trigger *cron.CronTrigger) error {
	logger := slog.Default()
	config := &Config{}
	if err := json.Unmarshal(runtime.Config(), config); err != nil {
		logger.Error("error unmarshalling config", "err", err)
	}

	reserveInfo, err := sdk.RunInNodeMode(
		runtime,
		fetchPor,
		pb.SimpleConsensusType_MEDIAN_OF_FIELDS).
		Await()

	if err != nil {
		return err
	}

	if reserveInfo.LastUpdated.Before(time.Unix(trigger.ScheduledExecutionTime, 0).Add(-time.Hour * 24)) {
		logger.Warn("reserve time is too old", "time", reserveInfo.LastUpdated)
		return errors.New("reserved time is too old")
	}

	totalSupply := big.NewInt(0)

	erc20, err := abi.JSON(strings.NewReader(Erc20Abi))
	if err != nil {
		return err
	}

	reserveManager, err := abi.JSON(strings.NewReader(ReserveManagerAbi))
	if err != nil {
		return err
	}
	// evmClient := &evm.Client{}

	evmClient := evm.NewEVMClient(nil)

	supplyPayload, err := erc20.Pack(TotalSupplyMethod)
	if err != nil {
		return err
	}

	// TODO other blockchains in parallel
	callReplay, err := evmClient.CallContract(context.Background(), &evmpb.CallContractRequest{
		Call: &evmpb.CallMsg{
			To: &evmpb.Address{
				Address: config.EvmTokenAddress,
			},
			Data: &evmpb.ABIPayload{
				Abi: supplyPayload,
			},
		},
		ConfidenceLevel: *chaincommonpb.Confidence_Finalized.Enum(),
	})

	if err != nil {
		// TODO specify which EVM
		logger.Error("Could not read from evm", "err", err.Error())
		return err
	}

	evmSupplyResponse, err := erc20.Unpack(TotalSupplyMethod, callReplay.Data.Abi)
	if err != nil {
		// TODO specify which EVM
		logger.Error("Could not unpack evm", "err", err.Error())
		return err
	}

	if len(evmSupplyResponse) != 1 {
		err = errors.New("unexpected number of results")
		logger.Error("Could not unpack evm", "err", err)
		return err
	}

	evmSupply, ok := evmSupplyResponse[0].(*big.Int)
	if !ok {
		err = errors.New("unexpected type returned")
		logger.Error("unexpected return type", "type", fmt.Sprintf("%T", evmSupplyResponse[0]))
		return err
	}

	totalSupply = totalSupply.Add(totalSupply, evmSupply)
	// TODO add other chains

	totalReserveScaled := reserveInfo.TotalReserve.Mul(decimal.NewFromUint64(10e18)).BigInt()
	evmSupplyCallData, err := reserveManager.Pack(UpdateReservesMethod, totalSupply, totalReserveScaled)
	if err != nil {
		logger.Error("Could not pack evm reserve call", "err", err.Error())
		return err
	}

	evmClient.CallContract(context.Background(), &evmpb.CallContractRequest{
		Call: &evmpb.CallMsg{
			To: &evmpb.Address{
				Address: config.EvmPorAddress,
			},
			Data: &evmpb.ABIPayload{
				Abi: evmSupplyCallData,
			},
		},
	})
	
	targetChainID := 100
	report := GenerateReport(targetChainID, evmSupplyCallData)
	
	updateSupplyReply, err :=  evmClient.WriteReport(context.Background(), &evmpb.WriteReportRequest{
		Receiver: &evmpb.Address{
			Address: config.EvmPorAddress,
		},
		Report: &evmpb.SignedReport{
			RawReport: report.RawReport,
			ReportContext: report.ReportContext,
			Signatures: report.Signatures,
			Id: report.ID,
		},
		GasConfig: &evmpb.GasConfig{
			GasLimit: config.EvmPorUpdateGasLimit,
		},
	})
	
	var writeErrors []error
	if err == nil {
		logger.Debug("Submitted transaction", "txHash", updateSupplyReply.TxHash)
		if updateSupplyReply.TxStatus == evm.TransactionStatus_TRANSACTION_STATUS_FATAL {
			//TODO missing error message from reply. We need to add an error message.
			writeErrors = append(writeErrors, errors.New("Failed to submit TX on chain"))	
		} else if updateSupplyReply.ReceiverContractExecutionStatus == evm.ReceiverContractExecutionStatus_FAILURE {
			//TODO missing error message from reply. We need to add an error message.
			writeErrors = append(writeErrors, errors.New("Failed executing POR smart contract update. POR contract failed to execute"))	
		} else {
			// TODO Super ugly. We need to improve this.
			time.Sleep(13 * time.Minute) //aprox. time it takes EVM to reach finality.
			for i := 0; i < 5; i++ {
				txStatusReply, err := evmClient.GetTransactionStatus(context.Background(), &evmpb.GetTransactionStatusRequest{
					TxHash: updateSupplyReply.TxHash,		
				})
				if err != nil {
					writeErrors = append(writeErrors, err)
					break
				}
				if txStatusReply.TransactionStatus == evm.TransactionStatus_TRANSACTION_STATUS_FINALIZED {
					return nil;
				}
				time.Sleep(5 * time.Second)
			}
			writeErrors = append(writeErrors, errors.New("Failed to reach finality for tx hash: " + updateSupplyReply.TxStatus.String()))
		}
	} else {
		logger.Error("failed to submit transaction", "err", err)
		writeErrors = append(writeErrors, err)
	}

	return errors.Join(writeErrors...)
}

type CommonReport struct {
	RawReport []byte
	ReportContext []byte
	Signatures [][]byte
	ID []byte
}

//TODO we need to define and implement this function
func GenerateReport(targetChainID int, evmSupplyCallData []byte) CommonReport {
	panic("unimplemented")
}

func fetchPor(runtime sdk.NodeRuntime) (*ReserveInfo, error) {
	config := &Config{}
	if err := json.Unmarshal(runtime.Config(), config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	request := &http.HttpFetchRequest{Url: config.Url}
	client := &http.Client{}
	response, err := client.Fetch(runtime, request).Await()
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

	reserve := &ReserveInfo{}
	if err = json.Unmarshal([]byte(porResponse.Data), reserve); err != nil {
		return nil, err
	}

	return reserve, nil
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
