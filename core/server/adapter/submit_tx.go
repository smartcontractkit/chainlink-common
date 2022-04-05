package adapter

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smartcontractkit/chainlink-relay/core/server/types"
	"github.com/smartcontractkit/chainlink/core/logger"
)

// TODO: implement (not high priority yet)

// Tx is the handler for when the CL node returns job data
type Tx struct{}

// NewTxHandler creates the Job handler for returned data
func NewTxHandler() Job {
	return Job{}
}

// Run is the endpoint for returning EA data for a specific job
func (t *Tx) Run(c *gin.Context) {
	var req types.JobRunData

	if err := c.BindJSON(&req); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	fmt.Println(req)
	c.JSON(http.StatusCreated, types.Resp{ID: req.JobID})
}
