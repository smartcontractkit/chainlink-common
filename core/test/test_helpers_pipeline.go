package test

import (
	"github.com/smartcontractkit/chainlink-relay/core/store/models"
)

// MockPipeline is a mocked services pipeline
type MockPipeline struct {
	Error error
}

func (mp MockPipeline) Start(models.Job) error {
	return mp.Error
}

func (mp MockPipeline) Run(string, string) error {
	return mp.Error
}

func (mp MockPipeline) Stop(string) error {
	return mp.Error
}
