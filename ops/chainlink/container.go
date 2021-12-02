package chainlink

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pulumi/pulumi-docker/sdk/v3/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/smartcontractkit/chainlink-relay/ops/utils"
	"github.com/smartcontractkit/integrations-framework/client"
)

// New spins up image for a chainlink node
func New(ctx *pulumi.Context, image *utils.Image, dbPort int, index int) (Node, error) {
	// treat index 0 as bootstrap
	indexStr := ""
	if index == 0 {
		indexStr = "bootstrap"
	} else {
		indexStr = strconv.Itoa(index - 1)
	}

	// TODO: Default ports?

	portStr := fmt.Sprintf("%d", config.RequireInt(ctx, "CL-PORT-START")+index)
	p2pPort := fmt.Sprintf("%d", config.RequireInt(ctx, "CL-P2P_PORT-START")+index)

	node := Node{
		Name: "chainlink-" + indexStr,
		P2P: "http://localhost:"+p2pPort,
		Config: client.ChainlinkConfig{
			URL:      "http://localhost:" + portStr,
			Email:    "admin@chain.link",
			Password: "twoChains",
			RemoteIP: "localhost",
		},
		Keys: map[string]string{},
	}

	// get env vars from YAML file
	envs, err := utils.GetEnvVars(ctx, "CL")
	if err != nil {
		return Node{}, err
	}

	// add additional configs (collected or calculated from environment configs)
	envs = append(envs,
		fmt.Sprintf("DATABASE_URL=postgresql://postgres@localhost:%d/chainlink_%s?sslmode=disable", dbPort, indexStr),
		fmt.Sprintf("CHAINLINK_PORT=%s", portStr),
		fmt.Sprintf("OCR2_P2PV2_LISTEN_ADDRESSES=127.0.0.1:%s", p2pPort),
		fmt.Sprintf("OCR2_P2PV2_ANNOUNCE_ADDRESSES=127.0.0.1:%s", p2pPort),
	)

	// fetch additional env vars (specific to each chainlink node)
	envListR, err := utils.GetEnvList(ctx, "CL_X")
	envsR := utils.GetVars(ctx, "CL_"+strings.ToUpper(indexStr), envListR)
	envs = append(envs, envsR...)

	entrypoints := pulumi.ToStringArray([]string{"chainlink", "node", "start", "-d", "-p", "/run/secrets/node_password", "-a", "/run/secrets/apicredentials"})
	uploads := docker.ContainerUploadArray{docker.ContainerUploadArgs{File: pulumi.String("/run/secrets/node_password"), Content: pulumi.String("abcd1234ABCD!@#$")}, docker.ContainerUploadArgs{File: pulumi.String("/run/secrets/apicredentials"), Content: pulumi.String(node.CredentialsString())}}

	var imageName pulumi.StringInput
	if config.GetBool(ctx, "CL-BUILD_LOCALLY") {
		imageName = image.Local.BaseImageName
	} else {
		imageName = image.Img.Name
	}

	_, err = docker.NewContainer(ctx, node.Name, &docker.ContainerArgs{
		Image:         imageName,
		Logs:          pulumi.BoolPtr(true),
		NetworkMode:   pulumi.String("host"),
		Envs:          pulumi.StringArrayInput(pulumi.ToStringArray(envs)),
		Uploads:       uploads.ToContainerUploadArrayOutput(),
		Entrypoints:   entrypoints.ToStringArrayOutput(),
		Restart:       pulumi.String("on-failure"),
		MaxRetryCount: pulumi.Int(3),
		// Attach:        pulumi.BoolPtr(true),
	})
	return node, err
}
