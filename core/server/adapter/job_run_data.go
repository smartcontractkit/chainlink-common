package adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smartcontractkit/chainlink-relay/core/server/types"
	"github.com/smartcontractkit/chainlink/core/logger"
)

// Job is the handler for when the CL node returns job data
type Job struct {
	services types.ServicesPipeline
}

// NewJobHandler creates the Job handler for returned data
func NewJobHandler(s types.ServicesPipeline) Job {
	return Job{
		services: s,
	}
}

// Run is the endpoint for returning EA data for a specific job
func (j *Job) Run(c *gin.Context) {
	var req types.JobRunData

	if err := c.BindJSON(&req); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	// return job data to service
	if err := j.services.Run(req.JobID, req.Result); err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusCreated, types.Resp{ID: req.JobID})
}
