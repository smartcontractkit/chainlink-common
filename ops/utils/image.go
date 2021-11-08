package utils

import (
	"github.com/pulumi/pulumi-docker/sdk/v3/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// Image implements the struct for fetching images
type Image struct {
	Name  string
	Tag   string
	Img   *docker.RemoteImage
	Local *docker.Image
}

// Pull retrieves the specified container image
func (i *Image) Pull(ctx *pulumi.Context) error {
	img, err := docker.NewRemoteImage(ctx, i.Name, &docker.RemoteImageArgs{
		Name:        pulumi.String(i.Tag),
		KeepLocally: pulumi.BoolPtr(true),
	})
	i.Img = img
	return err
}

// Build creates the image for the specified relay dockerfile in the YAML config
func (i *Image) Build(ctx *pulumi.Context) error {
	// build local image
	img, err := docker.NewImage(ctx, i.Name, &docker.ImageArgs{
		ImageName: pulumi.String(i.Tag),
		// LocalImageName: pulumi.String(i.Tag),
		SkipPush: pulumi.Bool(true),
		Registry: docker.ImageRegistryArgs{},
		Build: docker.DockerBuildArgs{
			Context:    pulumi.String(config.Require(ctx, "R-CONTEXT")),
			Dockerfile: pulumi.String(config.Require(ctx, "R-DOCKERFILE")),
		},
	})
	i.Local = img
	return err
}
