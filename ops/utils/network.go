package utils

import (
	"github.com/pulumi/pulumi-docker/sdk/v3/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Create network
func CreateNetwork(ctx *pulumi.Context, nwName string) (*docker.Network, error) {
	network, err := docker.GetNetwork(ctx, nwName, nil, nil, nil)
	if err != nil {
		network, err := docker.NewNetwork(ctx, nwName, &docker.NetworkArgs{Name: pulumi.String(nwName)}, nil)
		return network, err
	}
	return network, err
}
