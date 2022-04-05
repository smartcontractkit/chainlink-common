package test

import (
	"os"

	"github.com/smartcontractkit/chainlink-relay/core/store/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// MockStore is a mocked version of the required DB interfaces
type MockStore struct {
	err error
}

func (ms MockStore) CreateJob(*models.Job) error {
	return ms.err
}

func (ms MockStore) DeleteJob(string) error {
	return ms.err
}

// NewGormDB returns a connection to a docker container postgres instance
func NewGormDB() (*gorm.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	return gorm.Open(postgres.Open(dbURL), &gorm.Config{})
}

// CreateTable creates drops the table if it exists and creates
func CreateTable(gDB *gorm.DB, m interface{}) error {
	migrator := gDB.Migrator()

	// check if table exists (remove if necessary)
	if migrator.HasTable(m) {
		if err := migrator.DropTable(m); err != nil {
			return err
		}
	}

	// create DB
	return migrator.AutoMigrate(m)
}
