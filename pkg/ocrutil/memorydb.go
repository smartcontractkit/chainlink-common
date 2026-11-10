package ocrutil

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// MemoryDB is an OCR database that holds an oracle's state in memory, for oracles whose state need not outlive their
// process: one that restarts rejoins from the current config, and loses only its in-flight rounds.
type MemoryDB struct {
	// name identifies the oracle in logs and errors.
	name                 string
	lggr                 logger.SugaredLogger
	config               *ocrtypes.ContractConfig
	states               map[ocrtypes.ConfigDigest]*ocrtypes.PersistentState
	pendingTransmissions map[ocrtypes.ReportTimestamp]ocrtypes.PendingTransmission
	protocolStates       map[ocrtypes.ConfigDigest]map[string][]byte

	mu sync.Mutex
}

var (
	_ ocrtypes.Database  = &MemoryDB{}
	_ ocr3types.Database = &MemoryDB{}
)

// NewMemoryDB returns an in-memory database for the oracle named name.
func NewMemoryDB(name string, lggr logger.Logger) *MemoryDB {
	return &MemoryDB{
		name:                 name,
		lggr:                 logger.Sugared(lggr).Named("MemoryDB"),
		states:               make(map[ocrtypes.ConfigDigest]*ocrtypes.PersistentState),
		pendingTransmissions: make(map[ocrtypes.ReportTimestamp]ocrtypes.PendingTransmission),
		protocolStates:       make(map[ocrtypes.ConfigDigest]map[string][]byte),
	}
}

func (db *MemoryDB) ReadState(ctx context.Context, cd ocrtypes.ConfigDigest) (ps *ocrtypes.PersistentState, err error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	ps, ok := db.states[cd]
	if !ok {
		return nil, fmt.Errorf("state not found for oracle %s, config digest %s", db.name, cd)
	}

	return ps, nil
}

func (db *MemoryDB) WriteState(ctx context.Context, cd ocrtypes.ConfigDigest, state ocrtypes.PersistentState) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.states[cd] = &state
	return nil
}

func (db *MemoryDB) ReadConfig(ctx context.Context) (c *ocrtypes.ContractConfig, err error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.config == nil {
		// Returning nil, nil because this is a cache miss
		return nil, nil
	}
	return db.config, nil
}

func (db *MemoryDB) WriteConfig(ctx context.Context, c ocrtypes.ContractConfig) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.config = &c

	cBytes, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("MemoryDB: WriteConfig failed to marshal config: %w", err)
	}

	db.lggr.Debugw("MemoryDB: WriteConfig", "ocrtypes.ContractConfig", string(cBytes))

	return nil
}

func (db *MemoryDB) StorePendingTransmission(ctx context.Context, t ocrtypes.ReportTimestamp, tx ocrtypes.PendingTransmission) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.pendingTransmissions[t] = tx
	return nil
}

func (db *MemoryDB) PendingTransmissionsWithConfigDigest(ctx context.Context, cd ocrtypes.ConfigDigest) (map[ocrtypes.ReportTimestamp]ocrtypes.PendingTransmission, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	m := make(map[ocrtypes.ReportTimestamp]ocrtypes.PendingTransmission)
	for k, v := range db.pendingTransmissions {
		if k.ConfigDigest == cd {
			m[k] = v
		}
	}

	return m, nil
}

func (db *MemoryDB) DeletePendingTransmission(ctx context.Context, t ocrtypes.ReportTimestamp) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	delete(db.pendingTransmissions, t)
	return nil
}

func (db *MemoryDB) DeletePendingTransmissionsOlderThan(ctx context.Context, t time.Time) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for k, v := range db.pendingTransmissions {
		if v.Time.Before(t) {
			delete(db.pendingTransmissions, k)
		}
	}

	return nil
}

func (db *MemoryDB) ReadProtocolState(
	ctx context.Context,
	configDigest ocrtypes.ConfigDigest,
	key string,
) ([]byte, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	value, ok := db.protocolStates[configDigest][key]
	if !ok {
		// Previously implementation returned nil if the state is not found
		return nil, nil
	}
	return value, nil
}

func (db *MemoryDB) WriteProtocolState(
	ctx context.Context,
	configDigest ocrtypes.ConfigDigest,
	key string,
	value []byte,
) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if value == nil {
		delete(db.protocolStates[configDigest], key)
	} else {
		if db.protocolStates[configDigest] == nil {
			db.protocolStates[configDigest] = make(map[string][]byte)
		}
		db.protocolStates[configDigest][key] = value
	}
	return nil
}
