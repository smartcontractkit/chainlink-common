package pluginprovider_test

import (
	"bytes"
	"context"
	"fmt"

	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"github.com/stretchr/testify/assert"
)

type ContractTransmitterTestConfig struct {
	ReportContext libocr.ReportContext
	Report        libocr.Report
	Sigs          []libocr.AttributedOnchainSignature

	ConfigDigest libocr.ConfigDigest
	Account      libocr.Account
	Epoch        uint32
}

type StaticContractTransmitter struct {
	ContractTransmitterTestConfig
}

func (s StaticContractTransmitter) Transmit(ctx context.Context, rc libocr.ReportContext, r libocr.Report, ss []libocr.AttributedOnchainSignature) error {
	if !assert.ObjectsAreEqual(s.ReportContext, rc) {
		return fmt.Errorf("expected report context %v but got %v", s.ReportContext, rc)
	}
	if !bytes.Equal(s.Report, r) {
		return fmt.Errorf("expected report %x but got %x", s.Report, r)
	}
	if !assert.ObjectsAreEqual(s.Sigs, ss) {
		return fmt.Errorf("expected signatures %v but got %v", s.Sigs, ss)
	}
	return nil
}

func (s StaticContractTransmitter) LatestConfigDigestAndEpoch(ctx context.Context) (libocr.ConfigDigest, uint32, error) {
	return s.ConfigDigest, s.Epoch, nil
}

func (s StaticContractTransmitter) FromAccount() (libocr.Account, error) {
	return s.Account, nil
}
