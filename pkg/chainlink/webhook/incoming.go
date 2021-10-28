package webhook

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smartcontractkit/chainlink-relay/pkg/server/types"
	"github.com/smartcontractkit/chainlink/core/logger"
)

const (
	webhookAccessKeyHeader = "X-Chainlink-EA-AccessKey"
	webhookSecretHeader    = "X-Chainlink-EA-Secret"
)

// Authenticate checks header for proper authentication
func Authenticate(accessKey, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqAccessKey := c.GetHeader(webhookAccessKeyHeader)
		reqSecret := c.GetHeader(webhookSecretHeader)
		if reqAccessKey == accessKey && reqSecret == secret {
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

// Webhook defines the functions for when the server requests various requests
type Webhook struct {
	store    types.SubscriptionStorer
	services types.ServicesPipeline
}

// New is function called to create new webhook endpoint handler
func New(store types.SubscriptionStorer, serv types.ServicesPipeline) Webhook {
	return Webhook{
		store:    store,
		services: serv,
	}
}

// CreateJob expects a CreateJobReq payload,
// validates the request and subscribes to the job.
func (w *Webhook) CreateJob(c *gin.Context) {
	var req types.CreateJobReq

	if err := c.BindJSON(&req); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	req.Params.JobID = req.JobID

	// store job in DB
	if err := w.store.CreateJob(&req.Params); err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	// start service
	if err := w.services.Start(req.Params); err != nil {
		logger.Error(err)
		// if error starting service, also delete from DB
		if err := w.store.DeleteJob(req.JobID); err != nil {
			logger.Error(err)
		}
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusCreated, types.Resp{ID: req.JobID})
}

// DeleteJob deletes any job with the jobid
// provided as parameter in the request.
func (w *Webhook) DeleteJob(c *gin.Context) {
	jobid := c.Param("jobid")

	// stop service first before deleting from DB
	if err := w.services.Stop(jobid); err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	// delete from DB
	if err := w.store.DeleteJob(jobid); err != nil {
		logger.Error(err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, types.Resp{ID: jobid})
}
