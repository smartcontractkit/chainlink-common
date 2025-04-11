package pkg

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/cron"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/node/http"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

//go:embed solc/bin/IERC20.abi
var erc20Abi string

//go:embed solc/bin/IReserveManager.abi
var reserveManagerAbi string

const totalSupplyMethod = "totalSupplyMethod"
const updateReservesMethod = "updateReserves"

type Config struct {
	EvmTokenAddress string
	EvmPorAddress   string
	Schedule        string
	PublicKey       string
}

func Workflow(runner sdk.DonRunner) {
	config := &Config{}
	if err := json.Unmarshal(runner.Config(), config); err != nil {
		// TODO log
		return
	}

	sdk.SubscribeToDonTrigger(
		runner,
		cron.Cron{}.Trigger(&cron.Config{Schedule: config.Schedule}),
		func(runtime sdk.DonRuntime, trigger *cron.CronTrigger) (struct{}, error) {
			return onCronTrigger(runtime, trigger, config)
		})
}

func onCronTrigger(runtime sdk.DonRuntime, trigger *cron.CronTrigger, config *Config) (struct{}, error) {
	reserveInfo, err := sdk.RunInNodeModeWithBuiltInConsensus(runtime, fetchPor, pb.SimpleConsensusType_MEDIAN_OF_FIELDS).
		Await()

	if err != nil {
		return struct{}{}, err
	}

	if reserveInfo.LastUpdated.Before(time.Unix(trigger.ScheduledExecutionTime, 0).Add(-time.Hour)) {
		return struct{}{}, errors.New("reserved time is too old")
	}
	totalSupply := big.NewInt(0)

	erc20, err := abi.JSON(strings.NewReader(erc20Abi))
	if err != nil {
		return struct{}{}, err
	}

	reserveManager, err := abi.JSON(strings.NewReader(reserveManagerAbi))
	if err != nil {
		return struct{}{}, err
	}
	evmClient := &evm.Client{}

	supplyPayload, err := erc20.Pack(totalSupplyMethod)
	if err != nil {
		return struct{}{}, err
	}

	evmPromise := evmClient.ReadMethod(runtime, &evm.ReadMethodRequest{
		Address:         config.EvmTokenAddress,
		Calldata:        supplyPayload,
		ConfidenceLevel: evm.ConfidenceLevel_FINALITY,
	})
	// TODO other blockchains in parallel

	evmRead, err := evmPromise.Await()
	if err != nil {
		// TODO add logging
		return struct{}{}, err
	}

	evmSupplyResponse, err := erc20.Unpack(totalSupplyMethod, evmRead.Value)
	if err != nil {
		return struct{}{}, err
	}

	if len(evmSupplyResponse) != 1 {
		return struct{}{}, errors.New("unexpected number of results")
	}

	evmSupply, ok := evmSupplyResponse[0].(*big.Int)
	if !ok {
		return struct{}{}, errors.New("unexpected type")
	}

	totalSupply = totalSupply.Add(totalSupply, evmSupply)
	// TODO add other chains
	evmSupplyCallData, err := reserveManager.Pack(updateReservesMethod, totalSupply, reserveInfo.TotalReserve)

	evmTx := evmClient.SubmitTransaction(runtime, &evm.SubmitTransactionRequest{
		ToAddress: config.EvmPorAddress,
		Calldata:  evmSupplyCallData,
	})

	var writeErrors []error
	txId, err := evmTx.Await()
	if err == nil {
		// TODO log txId
		_ = txId
	} else {
		writeErrors = append(writeErrors, err)
	}

	if len(writeErrors) > 0 {
		// TODO log errors
		return struct{}{}, errors.Join(writeErrors...)
	}

	return struct{}{}, nil
}

func fetchPor(runtime sdk.NodeRuntime) (*ReserveInfo, error) {
	request := &http.HttpFetchRequest{Url: "https://reserves.gousd.com/por.json"}
	client := &http.Client{}
	response, err := client.Fetch(runtime, request).Await()
	if err != nil {
		return nil, err
	}

	porResponse := &PorResponse{}
	if err = json.Unmarshal(response.Body, porResponse); err != nil {
		return nil, err
	}

	err = verifySignature(porResponse)
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

func verifySignature(porResponse *PorResponse) error {
	rawSig, err := base64.RawURLEncoding.DecodeString(porResponse.DataSignature)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(rawSig)
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

	hasher := crypto.SHA256.New()
	hasher.Write([]byte(porResponse.Data))
	digest := hasher.Sum(nil)

	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, digest, rawSig)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}
	return nil
}
