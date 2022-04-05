package keys

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/smartcontractkit/chainlink-relay/core/server/types"
	"github.com/smartcontractkit/chainlink/core/logger"
)

type Keys struct {
	pubKeys *map[string]string
}

func New(keys *map[string]string) Keys {
	return Keys{
		pubKeys: keys,
	}
}

// ShowKeys returns an object with various keys
func (k Keys) ShowKeys(c *gin.Context) {
	c.JSON(200, k.pubKeys)
}

func (k Keys) SetEIKeys(c *gin.Context) {
	var req types.SetKeyData
	if err := c.BindJSON(&req); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	// validate that both key/secret pairs are present
	if ok := req.Validate(); !ok {
		logger.Error(fmt.Errorf("missing a key: %+v", req))
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	// set environment variables
	kV := map[string]string{}
	kV["IC_ACCESSKEY"] = req.ICKey
	kV["IC_SECRET"] = req.ICSecret
	kV["CI_ACCESSKEY"] = req.CIKey
	kV["CI_SECRET"] = req.CISecret

	for k, v := range kV {
		if err := os.Setenv(k, v); err != nil {
			logger.Error(err)
			c.JSON(http.StatusInternalServerError, nil)
			return
		}
	}

	c.JSON(http.StatusCreated, nil)
}
