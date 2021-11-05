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

type P2pPeers struct {
	gorm.Model
	Addr   string
	PeerID string `gorm:"not null"`
}

type Offchainreporting2ContractConfigs struct {
	gorm.Model
	Offchainreporting2OracleSpecID uint32         `gorm:"primaryKey;unique"`
	ConfigDigest                   [32]byte       `gorm:"type:bytea"`
	Signers                        pq.ByteaArray  `gorm:"type:bytea[]"`
	Transmitters                   pq.StringArray `gorm:"type:text[]"`
	OnchainConfig                  []byte         `gorm:"type:bytea"`
	OffchainConfig                 []byte         `gorm:"type:bytea"`
	ConfigCount                    uint64
	F                              uint8
	OffchainConfigVersion          uint64
}

type Offchainreporting2DiscovererAnnouncements struct {
	LocalPeerID  string `gorm:"primaryKey"`
	RemotePeerID string `gorm:"primaryKey"`
	Ann          []byte `gorm:"type:bytea"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Offchainreporting2PersistentStates struct {
	Offchainreporting2OracleSpecID uint32 `gorm:"primaryKey"`
	ConfigDigest                   []byte `gorm:"primaryKey;type:bytea"`
	Epoch                          uint32
	HighestSentEpoch               uint32
	HighestReceivedEpoch           []uint32 `gorm:"type:bigint[]"`
	CreatedAt                      time.Time
	UpdatedAt                      time.Time
}

type Offchainreporting2PendingTransmissions struct {
	Offchainreporting2OracleSpecID uint32        `gorm:"primaryKey"`
	ConfigDigest                   []byte        `gorm:"primaryKey;type:bytea"`
	Epoch                          uint32        `gorm:"primaryKey"`
	Round                          uint8         `gorm:"primaryKey"`
	ExtraHash                      []byte        `gorm:"type:bytea"`
	Report                         []byte        `gorm:"type:bytea"`
	AttributedSignatures           pq.ByteaArray `gorm:"type:bytea[]"`
	Time                           time.Time
	CreatedAt                      time.Time
	UpdatedAt                      time.Time
}
