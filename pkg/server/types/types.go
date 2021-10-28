package types

import (
	"github.com/smartcontractkit/chainlink-relay/pkg/store/models"
)

// SubscriptionStorer is the interface for interacting with the database
type SubscriptionStorer interface {
	CreateJob(sub *models.Job) error
	DeleteJob(jobid string) error
}

// ServicesPipeline is the interface for interacting with the services pipeline
type ServicesPipeline interface {
	Start(models.Job) error
	Run(string, string) error
	Stop(string) error
}

// CreateJobReq holds the payload expected for job POSTs
// from the Chainlink node.
type CreateJobReq struct {
	JobID  string     `json:"jobId"`
	Name   string     `json:"type"`
	Params models.Job `json:"params"`
}

// Resp is the struct for returning data
type Resp struct {
	ID string `json:"id"`
}

// JobRunData holds the expected CL job run response
type JobRunData struct {
	JobID  string `json:"jobID"`
	Result string `json:"result"`
}
