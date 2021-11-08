package relay

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/smartcontractkit/chainlink-relay/ops/utils"
)

// Relay is an object that contains the methods for queurying the relay
type Relay struct {
	Name string
	URL  string
	P2P  string
	Keys map[string]string
}

func (r *Relay) GetKeys() error {
	msg := utils.LogStatus(fmt.Sprintf("Retrieved keys from %s", r.Name))
	resp, err := retryablehttp.Get("http://" + r.URL + "/keys")
	if err != nil {
		return msg.Check(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return msg.Check(err)
	}

	return msg.Check(json.Unmarshal(body, &r.Keys))
}
