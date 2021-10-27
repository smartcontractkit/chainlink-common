package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Job is the model for the database
type Job struct {
	gorm.Model
	JobID                                  string         `gorm:"unique;not null"`
	IsBootstrapPeer                        bool           `json:"isBootstrapPeer"`
	ContractAddress                        string         `json:"contractAddress"`
	KeyBundleID                            string         `json:"keyBundleID"`
	P2PBootstrapPeers                      pq.StringArray `json:"p2pBootstrapPeers" gorm:"type:text[]"`
	ContractConfigConfirmations            uint16         `json:"contractConfigConfirmations"`
	ContractConfigTrackerSubscribeInterval string         `json:"contractConfigTrackerSubscribeInterval"`
	ObservationTimeout                     string         `json:"observationTimeout"`
	BlockchainTimeout                      string         `json:"blockchainTimeout"`
}

type EncryptedKeyRings struct {
	gorm.Model
	UpdatedAt     time.Time
	EncryptedKeys []byte `gorm:"type:text"`
}

type EthKeyStates gorm.Model

type P2PPeers struct {
	gorm.Model
	Addr   string
	PeerID string `gorm:"not null"`
}
