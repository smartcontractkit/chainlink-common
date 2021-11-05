package services

import (
	"fmt"
	"math/big"
	"time"

	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink-relay/core/config"
	"github.com/smartcontractkit/chainlink-relay/core/server/webhook"
	"github.com/smartcontractkit/chainlink-relay/core/services/types"
	"github.com/smartcontractkit/chainlink-relay/core/store"
	"github.com/smartcontractkit/chainlink-relay/core/store/models"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/services/keystore"
	"github.com/smartcontractkit/chainlink/core/services/ocrcommon"
	ocrcore "github.com/smartcontractkit/chainlink/core/services/offchainreporting2"
	"github.com/smartcontractkit/libocr/commontypes"
	ocr "github.com/smartcontractkit/libocr/offchainreporting2"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2/types"
	"gorm.io/gorm"
)

type ocrService interface {
	Start() error
	Close() error
}

// OCR2 is the service for offchain reporting 2 bootstrap and normal nodes
type OCR2 struct {
	oracle   ocrService
	job      models.Job
	webhook  webhook.Trigger
	runData  chan *big.Int
	contract types.ContractTracker
	log      *logger.Logger
}

// NewOCR2 creates a new OCR2 service
func NewOCR2(job models.Job, gormdb *gorm.DB, trigger webhook.Trigger, cfg config.Config, keys keystore.Master, blockchain types.Blockchain, peerWrapper *ocrcore.SingletonPeerWrapper) (OCR2, error) {

	// create local config
	bTimeout, err := time.ParseDuration(job.BlockchainTimeout)
	if err != nil {
		return OCR2{}, err
	}
	localConfig := ocrtypes.LocalConfig{
		BlockchainTimeout:                  bTimeout,
		ContractConfigConfirmations:        job.ContractConfigConfirmations,
		SkipContractConfigConfirmations:    false,
		ContractConfigTrackerPollInterval:  cfg.OCR2ContractPollInterval(),
		ContractTransmitterTransmitTimeout: cfg.OCR2ContractTransmitterTransmitTimeout(),
		DatabaseTimeout:                    cfg.OCR2DatabaseTimeout(),
		DevelopmentMode:                    "",
	}

	// connect to DB
	sqldb, errdb := gormdb.DB()
	if errdb != nil {
		return OCR2{}, errors.Wrap(errdb, "unable to open sql db")
	}
	ocrdb := ocrcore.NewDB(sqldb, 1)

	// contract tracker implements all terra configs necessary
	contractTracker, err := blockchain.NewContractTracker(job.ContractAddress, job.JobID)
	if err != nil {
		return OCR2{}, err
	}

	// parse bootstrap peers
	var v2BootstrapPeers []commontypes.BootstrapperLocator
	for _, peer := range job.P2PBootstrapPeers {
		var bootstrapPeer commontypes.BootstrapperLocator
		if err := bootstrapPeer.UnmarshalText([]byte(peer)); err != nil {
			return OCR2{}, err
		}
		v2BootstrapPeers = append(v2BootstrapPeers, bootstrapPeer)
	}

	// create logger
	defaultLogger := logger.Default.Named("OCR2")
	ocrLogger := ocrcommon.NewLogger(defaultLogger, true, func(string) {})

	// create DataSource
	runData := make(chan *big.Int)
	ds := DataSource{id: job.JobID, webhook: &trigger, runData: &runData, log: defaultLogger}

	// create median plug in
	numericalMedianFactory := median.NumericalMedianFactory{
		ContractTransmitter:   contractTracker,
		DataSource:            &ds,
		JuelsPerEthDataSource: &JuelsToEthDataSource{},
		Logger:                ocrLogger,
		ReportCodec:           contractTracker,
	}

	// fetch key
	ocrkey, err := keys.OCR().Get(job.KeyBundleID)
	if err != nil {
		return OCR2{}, err
	}
	ocr2keyring := store.NewOCR2KeyWrapper(ocrkey) // see keystore.go for details

	// create bootstrap node or reporting node
	var oracleNode ocrService
	if job.IsBootstrapPeer {
		bootstrapArgs := ocr.BootstrapperArgs{
			BootstrapperFactory:    peerWrapper.Peer,
			V2Bootstrappers:        v2BootstrapPeers,
			ContractConfigTracker:  contractTracker,
			Database:               ocrdb,
			LocalConfig:            localConfig,
			Logger:                 ocrLogger,
			MonitoringEndpoint:     Monitoring(),
			OffchainConfigDigester: contractTracker,
		}

		oracleNode, err = ocr.NewBootstrapper(bootstrapArgs)
		if err != nil {
			return OCR2{}, err
		}
	} else {
		ocrArgs := ocr.OracleArgs{
			BinaryNetworkEndpointFactory: peerWrapper.Peer,
			V2Bootstrappers:              v2BootstrapPeers,
			ContractTransmitter:          contractTracker,
			ContractConfigTracker:        contractTracker,
			Database:                     ocrdb,
			LocalConfig:                  localConfig,
			Logger:                       ocrLogger,
			MonitoringEndpoint:           Monitoring(),
			OffchainConfigDigester:       contractTracker,
			OffchainKeyring:              &ocr2keyring,
			OnchainKeyring:               blockchain.OCR(),
			ReportingPluginFactory:       numericalMedianFactory,
		}

		oracleNode, err = ocr.NewOracle(ocrArgs)
		if err != nil {
			return OCR2{}, err
		}
	}

	// return OCR2 object
	return OCR2{
		job:      job,
		oracle:   oracleNode,
		webhook:  trigger,
		runData:  runData,
		contract: contractTracker,
		log:      defaultLogger,
	}, nil
}

// Start provides the standard service interface
func (o *OCR2) Start() {
	// start contract subscription
	go o.contract.Start()

	o.log.Infof("[%s] Start OCR2 service", o.job.JobID)
	if err := o.oracle.Start(); err != nil {
		o.log.Errorf("[%s] Failed to start OCR2 service", o.job.JobID)
	}
}

// Stop provides the standard service interface (wraps other stop functionality together)
func (o *OCR2) Stop() error {
	o.log.Infof("[%s] Stopping OCR2 service", o.job.JobID)
	close(o.runData)
	if err := o.contract.Close(); err != nil {
		return err
	}
	return o.oracle.Close()
}

// Run provides the standard service interface (wraps number parsing together)
func (o *OCR2) Run(data string) error {
	var val big.Int
	_, ok := val.SetString(data, 10)
	if !ok {
		return fmt.Errorf("Failed to parse data string (%s) to *big.Int", data)
	}

	o.runData <- &val // pass data back to job run
	return nil
}
