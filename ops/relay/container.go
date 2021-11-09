package relay

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pulumi/pulumi-docker/sdk/v3/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/smartcontractkit/chainlink-relay/ops/utils"
)

// New creates the container for a new relay
func New(ctx *pulumi.Context, image *docker.Image, dbPort int, index int, secrets map[string]string) (Relay, error) {
	// treat index 0 as bootstrap
	indexStr := ""
	if index == 0 {
		indexStr = "bootstrap"
	} else {
		indexStr = strconv.Itoa(index - 1)
	}

	// get env vars from YAML file (staddard across relays)
	envs, err := utils.GetEnvVars(ctx, "R")
	if err != nil {
		return Relay{}, err
	}

	// add additional configs (collected or calculated from environment configs)
	envs = append(envs,
		fmt.Sprintf("DATABASE_URL=postgresql://postgres@localhost:%d/relay_%s?sslmode=disable", dbPort, indexStr),
		fmt.Sprintf("CLIENT_NODE_URL=%s", "http://localhost:6688"),
		fmt.Sprintf("IC_ACCESSKEY=%s", secrets["ic-key"]),
		fmt.Sprintf("IC_SECRET=%s", secrets["ic-secret"]),
		fmt.Sprintf("CI_ACCESSKEY=%s", secrets["ci-key"]),
		fmt.Sprintf("CI_SECRET=%s", secrets["ci-secret"]),
		fmt.Sprintf("CHAINLINK_PORT=%d", config.RequireInt(ctx, "R-PORT-START")+index),
		fmt.Sprintf("OCR2_P2PV2_LISTEN_ADDRESSES=127.0.0.1:%d", config.RequireInt(ctx, "R-P2P_LISTEN_PORT-START")+index),
		fmt.Sprintf("OCR2_P2PV2_ANNOUNCE_ADDRESSES=127.0.0.1:%d", config.RequireInt(ctx, "R-P2P_LISTEN_PORT-START")+index),
	)

	// fetch additional env vars (specific to each relay)
	envListR, err := utils.GetEnvList(ctx, "R_X")
	envsR := utils.GetVars(ctx, "R_"+strings.ToUpper(indexStr), envListR)
	envs = append(envs, envsR...)

	_, err = docker.NewContainer(ctx, "relay-"+indexStr, &docker.ContainerArgs{
		Image:         image.BaseImageName,
		Logs:          pulumi.BoolPtr(true),
		NetworkMode:   pulumi.String("host"),
		Envs:          pulumi.StringArrayInput(pulumi.ToStringArray(envs)),
		Restart:       pulumi.String("on-failure"),
		MaxRetryCount: pulumi.Int(1),
	})
	return Relay{
		Name: fmt.Sprintf("relay_%s", indexStr),
		URL:  fmt.Sprintf("localhost:%d", config.RequireInt(ctx, "R-PORT-START")+index),
		P2P:  fmt.Sprintf("localhost:%d", config.RequireInt(ctx, "R-P2P_LISTEN_PORT-START")+index),
	}, err
}
