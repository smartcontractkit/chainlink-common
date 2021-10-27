package webhook

import (
	"fmt"
	"net/http"

	retryHTTP "github.com/hashicorp/go-retryablehttp"
	"github.com/smartcontractkit/chainlink/core/logger"
)

// Trigger is the data object for holding the needed pieces of information
type Trigger struct {
	url, icKey, icSecret string
	client               *retryHTTP.Client
	log                  *logger.Logger
}

// NewTrigger creates a client for triggering jobs in CL node
func NewTrigger(url, icKey, icSecret string) Trigger {
	return Trigger{url, icKey, icSecret, retryHTTP.NewClient(), logger.Default.Named("webhook-outgoing")}
}

// TriggerJob sends a POST request to the CL node to trigger a job run
func (t *Trigger) TriggerJob(jobid string) {
	t.log.Infof("[%s] Sending job trigger", jobid)
	url := fmt.Sprintf("%s/v2/jobs/%s/runs", t.url, jobid)
	request, err := retryHTTP.NewRequest(http.MethodPost, url, []byte{})
	if err != nil {
		logger.Error(err)
		return
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Add(webhookAccessKeyHeader, t.icKey)
	request.Header.Add(webhookSecretHeader, t.icSecret)
	res, err := t.client.Do(request)
	if err != nil {
		logger.Error(err)
		return
	}
	t.log.Infof("[%s] Job trigger status: %s", jobid, res.Status)
}
