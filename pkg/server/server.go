package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Depado/ginprom"
	"github.com/gin-gonic/gin"
	"github.com/smartcontractkit/chainlink-relay/pkg/chainlink/webhook"
	"github.com/smartcontractkit/chainlink-relay/pkg/server/adapter"
	"github.com/smartcontractkit/chainlink-relay/pkg/server/types"
	"github.com/smartcontractkit/chainlink/core/logger"
)

var ginPrometheus *ginprom.Prometheus

func init() {
	// ensure metrics are regsitered once per instance to avoid registering
	// metrics multiple times (panic)
	ginPrometheus = ginprom.New(ginprom.Namespace("service"))
}

// RunWebserver starts a new web server using the access key
// and secret as provided on protected routes.
func RunWebserver(
	accessKey, secret string,
	store types.SubscriptionStorer,
	port int,
	services types.ServicesPipeline,
	pubKeys *map[string]string,
) {
	srv := NewHTTPService(accessKey, secret, store, services, pubKeys)
	addr := fmt.Sprintf(":%v", port)
	err := srv.Router.Run(addr)
	if err != nil {
		logger.Error(err)
	}
}

// HttpService encapsulates router, webhook service
// and access credentials.
type HttpService struct {
	Router    *gin.Engine
	AccessKey string
	Secret    string
	Store     types.SubscriptionStorer
	services  types.ServicesPipeline
	log       *logger.Logger
	pubKeys   *map[string]string
}

// NewHTTPService creates a new HttpService instance
// with the default router.
func NewHTTPService(
	accessKey, secret string,
	store types.SubscriptionStorer,
	services types.ServicesPipeline,
	pubKeys *map[string]string,
) *HttpService {
	srv := HttpService{
		AccessKey: accessKey,
		Secret:    secret,
		Store:     store,
		services:  services,
		log:       logger.Default.Named("server"),
		pubKeys:   pubKeys,
	}
	srv.createRouter()
	return &srv
}

func (srv *HttpService) createRouter() {
	engine := gin.New()

	ginPrometheus.Use(engine)
	engine.Use(
		gin.Recovery(),
		loggerFunc(srv.log),
		ginPrometheus.Instrument(),
	)

	engine.GET("/health", srv.ShowHealth)

	// adapter implementation for job run data
	// TODO: Authentication for this (bridges have auth tokens but different than EI keys)
	job := adapter.NewJobHandler(srv.services)
	engine.POST("/runs", job.Run)

	// TODO: expose the sign and broadcast functionality like an EA
	// tx := adapter.NewTxHandler()
	// engine.POST("/tx", tx.Run)

	// endpoint implementation for key data
	// TODO: add authentication for retrieving keys
	engine.GET("/keys", srv.ShowKeys)

	// webhook implementation
	wh := webhook.New(srv.Store, srv.services)
	authJob := engine.Group("/jobs")
	authJob.Use(webhook.Authenticate(srv.AccessKey, srv.Secret))
	{
		authJob.POST("/", wh.CreateJob)
		authJob.DELETE("/:jobid", wh.DeleteJob)
	}

	srv.Router = engine
}

// ShowHealth returns the following when online:
//  {"chainlink": true}
func (srv *HttpService) ShowHealth(c *gin.Context) {
	c.JSON(200, gin.H{"chainlink": true})
}

// ShowKeys returns an object with various keys
func (srv *HttpService) ShowKeys(c *gin.Context) {
	c.JSON(200, srv.pubKeys)
}

// Inspired by https://github.com/gin-gonic/gin/issues/961
func loggerFunc(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		buf, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Error("Web request log error: ", err.Error())
			// Implicitly relies on limits.RequestSizeLimiter
			// overriding of c.Request.Body to abort gin's Context
			// inside ioutil.ReadAll.
			// Functions as we would like, but horrible from an architecture
			// and design pattern perspective.
			if !c.IsAborted() {
				c.AbortWithStatus(http.StatusBadRequest)
			}
			return
		}
		rdr := bytes.NewBuffer(buf)
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(buf))

		start := time.Now()
		c.Next()
		end := time.Now()

		log.Infow(fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
			"status", c.Writer.Status(),
			"query", c.Request.URL.Query(),
			"body", readBody(rdr),
			"clientIP", c.ClientIP(),
			"errors", c.Errors.String(),
			"servedAt", end.Format("2006-01-02 15:04:05"),
			"latency", fmt.Sprint(end.Sub(start)),
		)
	}
}

func readBody(reader io.Reader) string {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(reader)
	if err != nil {
		logger.Warn("unable to read from body for sanitization: ", err)
		return "*FAILED TO READ BODY*"
	}

	if buf.Len() == 0 {
		return ""
	}

	s, err := readSanitizedJSON(buf)
	if err != nil {
		logger.Warn("unable to sanitize json for logging: ", err)
		return "*FAILED TO READ BODY*"
	}
	return s
}

func readSanitizedJSON(buf *bytes.Buffer) (string, error) {
	var dst map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &dst)
	if err != nil {
		return "", err
	}

	b, err := json.Marshal(dst)
	if err != nil {
		return "", err
	}
	return string(b), err
}
