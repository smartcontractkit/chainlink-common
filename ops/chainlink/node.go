package chainlink

import (
	"fmt"
	"net/http"
	"time"

	"github.com/smartcontractkit/chainlink-relay/ops/utils"
	"github.com/smartcontractkit/integrations-framework/client"
)

// Node implements the node parameters
type Node struct {
	Name   string
	Config client.ChainlinkConfig
	Call   client.Chainlink
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
	msg := utils.LogStatus(fmt.Sprintf("Adding %s EA to CL node", name))

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

// func (n *Node) AddEI(name, url string) (map[string]string, error) {
// 	msg := utils.LogStatus(fmt.Sprintf("Adding %s EI to CL node", name))
//
// 	// check if EI exists and delete (old secrets cannot be retrieved)
// 	eis, err := n.Call.ReadEIs()
// 	if err != nil {
// 		return map[string]string{}, err
// 	}
// 	for _, e := range eis.Data {
// 		if e.Attributes.Name == name {
// 			msg.Exists()
// 			fmt.Print(" (recreating)")
// 			if err := n.Call.DeleteEI(name); err != nil {
// 				return map[string]string{}, err
// 			}
// 		}
// 	}
//
// 	ei, err := n.Call.CreateEI(&client.EIAttributes{
// 		Name: name,
// 		URL:  url,
// 	})
// 	params := map[string]string{}
// 	params["ic-key"] = ei.Data.Attributes.IncomingAccessKey
// 	params["ic-secret"] = ei.Data.Attributes.Secret
// 	params["ci-key"] = ei.Data.Attributes.OutgoingToken
// 	params["ci-secret"] = ei.Data.Attributes.OutgoingSecret
// 	return params, msg.Check(err)
// }

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
