package store

import (
	"errors"
	"net/url"

	"github.com/smartcontractkit/chainlink-relay/core/store/models"
	"github.com/smartcontractkit/chainlink/core/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Store object
type Store struct {
	DB  *gorm.DB
	log *logger.Logger
}

// Connect to the databse
func (s *Store) Connect(url url.URL) error {
	db, err := gorm.Open(postgres.Open(url.String()), &gorm.Config{})
	if err != nil {
		return err
	}
	s.DB = db
	s.log = logger.Default.Named("database")
	return s.Migrate()
}

func (s *Store) Migrate() error {
	if err := s.DB.AutoMigrate(&models.Job{}); err != nil {
		return err
	}
	if err := s.DB.AutoMigrate(&models.EncryptedKeyRings{}); err != nil {
		return err
	}
	if err := s.DB.AutoMigrate(&models.EthKeyStates{}); err != nil {
		return err
	}
	if err := s.DB.AutoMigrate(&models.P2pPeers{}); err != nil {
		return err
	}
	if err := s.DB.AutoMigrate(&models.Offchainreporting2ContractConfigs{}); err != nil {
		return err
	}
	if err := s.DB.AutoMigrate(&models.Offchainreporting2DiscovererAnnouncements{}); err != nil {
		return err
	}
	if err := s.DB.AutoMigrate(&models.Offchainreporting2PersistentStates{}); err != nil {
		return err
	}
	if err := s.DB.AutoMigrate(&models.Offchainreporting2PendingTransmissions{}); err != nil {
		return err
	}
	return nil
}

// CreateJob saves the job data in the DB
func (s Store) CreateJob(job *models.Job) error {
	if job.JobID == "" {
		return errors.New("JobID cannot be blank")
	}

	s.log.Infof("Job Created: %s", job.JobID)
	return s.DB.Create(job).Error
}

// LoadJobs retrieves all jobs from the db
func (s Store) LoadJobs() ([]models.Job, error) {
	var jobs []models.Job
	err := s.DB.Find(&jobs).Error
	s.log.Infof("Jobs Loaded: %d jobs", len(jobs))
	return jobs, err
}

// loadJob retrieves a specific job from the db
func (s Store) loadJob(jobid string) (*models.Job, error) {
	var job models.Job
	if err := s.DB.Where("job_id = ?", jobid).First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

// DeleteJob removes the job data from the DB
func (s *Store) DeleteJob(jobid string) error {
	if jobid == "" {
		return errors.New("JobID cannot be blank")
	}

	s.log.Infof("Job Deleted %s", jobid)
	job, err := s.loadJob(jobid)
	if err != nil {
		return err
	}
	return s.DB.Delete(job).Error
}
