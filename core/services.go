package core

import (
	"github.com/smartcontractkit/chainlink-relay/core/config"
	"github.com/smartcontractkit/chainlink-relay/core/server/webhook"
	"github.com/smartcontractkit/chainlink-relay/core/services"
	"github.com/smartcontractkit/chainlink-relay/core/services/types"
	"github.com/smartcontractkit/chainlink-relay/core/store/models"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/services/keystore"
	ocrcore "github.com/smartcontractkit/chainlink/core/services/offchainreporting2"
	"gorm.io/gorm"
)

type Service interface {
	Start() // does not return anything because Start() is called via go routine
	Stop() error
	Run([]byte) error
}

type Services struct {
	services    map[string]Service
	webhook     webhook.Trigger
	cfg         config.Config
	db          *gorm.DB
	keys        keystore.Master
	blockchain  types.Blockchain
	Log         *logger.Logger
	peerWrapper *ocrcore.SingletonPeerWrapper
}

// NewServices create services manager
func NewServices(db *gorm.DB, cfg config.Config, keys keystore.Master, blockchainClient types.Blockchain) (Services, error) {
	log := logger.Default.Named("services-pipeline")
	if cfg.Mock() {
		log.Info("Mock service enabled. Disable to use OCR2")
	}

	return Services{
		db:          db,
		services:    map[string]Service{},
		webhook:     webhook.NewTrigger(cfg.ClientNodeURL(), cfg),
		cfg:         cfg,
		keys:        keys,
		blockchain:  blockchainClient,
		Log:         log,
		peerWrapper: ocrcore.NewSingletonPeerWrapper(keys, cfg, db),
	}, nil
}

// Start starts the service with a given jobid
func (s *Services) Start(job models.Job) error {
	var srv Service

	// create mock service
	if s.cfg.Mock() {
		mock, err := services.NewMockService(job, s.webhook, s.blockchain)
		if err != nil {
			return err
		}
		srv = &mock
	}

	// create ocr service
	if !s.cfg.Mock() {
		// start peerWrapper once
		if started := s.peerWrapper.IsStarted(); !started {
			if err := s.peerWrapper.Start(); err != nil {
				return err
			}
		}

		ocr2, err := services.NewOCR2(job, s.db, s.webhook, s.cfg, s.keys, s.blockchain, s.peerWrapper)
		if err != nil {
			return err
		}
		srv = &ocr2
	}

	s.services[job.JobID] = srv
	go s.services[job.JobID].Start()
	return nil
}

// Run is used in the web server to return job run data to the service that triggered a job run
func (s *Services) Run(jobid string, raw []byte) error {
	return s.services[jobid].Run(raw)
}

// Stop starts a specific service with a given jobid
func (s *Services) Stop(jobid string) error {
	// stop service
	if err := s.services[jobid].Stop(); err != nil {
		return err
	}
	delete(s.services, jobid)

	if len(s.services) == 0 {
		s.Log.Info("All jobs stopped")
	}
	return nil
}

// StopAll stops all services using the underlying Stop function
// only called when relay is being completely stopped
func (s *Services) StopAll() {
	// stop services
	for id := range s.services {
		if err := s.Stop(id); err != nil {
			logger.Error(err)
		}
	}

	// stop peerwrapper (if started)
	if started := s.peerWrapper.IsStarted(); started {
		if err := s.peerWrapper.Close(); err != nil {
			logger.Error(err)
		}
	}

}
