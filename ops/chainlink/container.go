package chainlink

import (
	"fmt"

	"github.com/pulumi/pulumi-docker/sdk/v3/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/smartcontractkit/chainlink-relay/ops/utils"
	"github.com/smartcontractkit/integrations-framework/client"
)

// New spins up image for a chainlink node
func New(ctx *pulumi.Context, image *docker.RemoteImage, dbPort int) (Node, error) {
	node := Node{
		Name: "chainlink-node",
		Config: client.ChainlinkConfig{
			URL:      "http://localhost:6688",
			Email:    "admin@chain.link",
			Password: "twoChains",
			RemoteIP: "localhost",
		},
	}

	// get env vars from YAML file
	envs, err := utils.GetEnvVars(ctx, "CL")
	if err != nil {
		return Node{}, err
	}

	// add DB url
	envs = append(envs, fmt.Sprintf("DATABASE_URL=postgresql://postgres@localhost:%d/chainlink?sslmode=disable", dbPort))

	entrypoints := pulumi.ToStringArray([]string{"chainlink", "node", "start", "-d", "-p", "/run/secrets/node_password", "-a", "/run/secrets/apicredentials"})
	uploads := docker.ContainerUploadArray{docker.ContainerUploadArgs{File: pulumi.String("/run/secrets/node_password"), Content: pulumi.String("abcd1234ABCD!@#$")}, docker.ContainerUploadArgs{File: pulumi.String("/run/secrets/apicredentials"), Content: pulumi.String(node.CredentialsString())}}

	_, err = docker.NewContainer(ctx, node.Name, &docker.ContainerArgs{
		Image: image.Name,
		// Attach:           pulumi.BoolPtr(true),
		Logs:          pulumi.BoolPtr(true),
		NetworkMode:   pulumi.String("host"),
		Envs:          pulumi.StringArrayInput(pulumi.ToStringArray(envs)),
		Uploads:       uploads.ToContainerUploadArrayOutput(),
		Entrypoints:   entrypoints.ToStringArrayOutput(),
		Restart:       pulumi.String("on-failure"),
		MaxRetryCount: pulumi.Int(3),
	})
	return node, err
}
