package chainlink

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/smartcontractkit/chainlink-relay/ops/utils"
	"github.com/smartcontractkit/integrations-framework/client"
)

// Node implements the node parameters
type Node struct {
	Name   string
	P2P    string
	Config client.ChainlinkConfig
	Call   client.Chainlink
	Keys   map[string]string
}

// CredentialsString returns formatted string for node input
func (n *Node) CredentialsString() string {
	return fmt.Sprintf("%s\n%s", n.Config.Email, n.Config.Password)
}

// Health returns if the node is functional or not
func (n *Node) Health() (interface{}, error) {
	return http.Get(n.Config.URL + "/health")
}

// Ready checks when node is ready
func (n *Node) Ready() error {
	msg := utils.LogStatus(fmt.Sprintf("Waiting for health checks on %s", n.Name))
	timeout := 30
	var err error
	time.Sleep(2 * time.Second) // removing this breaks running `up` multiple times...
	for i := 0; i < timeout; i++ {
		_, err = n.Health()
		if err == nil {
			cl, err := client.NewChainlink(&n.Config, http.DefaultClient)
			n.Call = cl
			return msg.Check(err)
		}
		time.Sleep(1 * time.Second)
	}
	return msg.Check(err)
}

// AddBridge adds adapter to CL node
func (n *Node) AddBridge(name, url string) error {
	msg := utils.LogStatus(fmt.Sprintf("Adding %s EA to %s", name, n.Name))

	// check if exists
	_, err := n.Call.ReadBridge(name)
	if err == nil {
		msg.Exists()
		return msg.Check(nil)
	}

	err = n.Call.CreateBridge(&client.BridgeTypeAttributes{
		Name: name,
		URL:  url,
	})
	return msg.Check(err)
}

func (n Node) DeleteAllJobs() error {
	msg := utils.LogStatus("Cleared existing jobs from CL node")

	// get all jobs
	jobs, err := n.Call.ReadJobs()
	for _, j := range jobs.Data {
		// remove job based on ID
		if err := n.Call.DeleteJob(j["id"].(string)); err != nil {
			return msg.Check(err)
		}
	}

	if len(jobs.Data) == 0 {
		fmt.Print(" - No jobs present")
	}
	return msg.Check(err)
}

func (n *Node) GetKeys() error {
	msg := utils.LogStatus(fmt.Sprintf("Retrieved keys from %s", n.Name))

	// TODO: placeholder for fetching keys
	ocrKeys, err := n.Call.ReadOCRKeys()
	if err != nil {
		return msg.Check(err)
	}
	p2pKeys, err := n.Call.ReadP2PKeys()
	if err != nil {
		return msg.Check(err)
	}
	addr, err := n.Call.PrimaryEthAddress()
	if err != nil {
		return msg.Check(err)
	}

	// parse keys into expected format
	n.Keys["OCRKeyID"] = ocrKeys.Data[0].ID
	n.Keys["OCROnchainPublicKey"] = ocrKeys.Data[0].Attributes.OnChainSigningAddress
	n.Keys["OCRTransmitter"] = addr
	n.Keys["OCROffchainPublicKey"] = ocrKeys.Data[0].Attributes.OffChainPublicKey
	n.Keys["OCRConfigPublicKey"] = ocrKeys.Data[0].Attributes.ConfigPublicKey
	n.Keys["P2PID"] = p2pKeys.Data[0].Attributes.PeerID

	// replace value with val without prefix if prefix exists
	for k, val := range n.Keys {
		sArr := strings.Split(val, "_")
		if len(sArr) == 2 {
			n.Keys[k] = sArr[1]
		}
	}

	return msg.Check(err)
}
