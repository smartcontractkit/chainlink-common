package ops

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/smartcontractkit/chainlink-relay/ops/adapter"
	"github.com/smartcontractkit/chainlink-relay/ops/chainlink"
	"github.com/smartcontractkit/chainlink-relay/ops/database"
	"github.com/smartcontractkit/chainlink-relay/ops/utils"
	"github.com/smartcontractkit/integrations-framework/client"
)

// Deployer interface for deploying contracts
type Deployer interface {
	Load() error                            // upload contracts (may not be necessary)
	DeployLINK() error                      // deploy LINK contract
	DeployOCR() error                       // deploy OCR contract
	TransferLINK() error                    // transfer LINK to OCR contract
	InitOCR(keys []map[string]string) error // initialize OCR contract with provided keys
	OCR2Address() string                    // fetch deployed OCR contract address
}

// ObservationSource creates the observation source for the CL node jobs
type ObservationSource func(priceAdapter string) string

func New(ctx *pulumi.Context, deployer Deployer, obsSource ObservationSource, juelsObsSource ObservationSource) error {
	img := map[string]*utils.Image{}

	// fetch postgres
	img["psql"] = &utils.Image{
		Name: "postgres-image",
		Tag:  "postgres:latest", // always use latest postgres
	}

	buildLocal := config.GetBool(ctx, "CL-BUILD_LOCALLY")
	if !buildLocal {
		// fetch chainlink image
		img["chainlink"] = &utils.Image{
			Name: "chainlink-remote-image",
			Tag:  "public.ecr.aws/chainlink/chainlink:" + config.Require(ctx, "CL-NODE_VERSION"),
		}
	}
	// TODO: build local chainlink image

	// fetch list of EAs
	eas := []string{}
	if err := config.GetObject(ctx, "EA-NAMES", &eas); err != nil {
		return err
	}
	for _, n := range eas {
		img[n] = &utils.Image{
			Name: n + "-adapter-image",
			Tag:  fmt.Sprintf("public.ecr.aws/chainlink/adapters/%s-adapter:develop-latest", n),
		}
	}

	// pull remote images
	for i := range img {
		if err := img[i].Pull(ctx); err != nil {
			return err
		}
	}

	// build local chainlink node
	if buildLocal {
		img["chainlink"] = &utils.Image{
			Name: "chainlink-local-build",
			Tag:  "chainlink:local",
		}
		if err := img["chainlink"].Build(ctx, config.Require(ctx, "CL-BUILD_CONTEXT"), config.Require(ctx, "CL-BUILD_DOCKERFILE")); err != nil {
			return err
		}
	}

	// validate number of relays
	nodeNum := config.GetInt(ctx, "CL-COUNT")
	if nodeNum < 4 {
		return fmt.Errorf("Minimum number of chainlink nodes (4) not met (%d)", nodeNum)
	}

	// start pg + create DBs
	db, err := database.New(ctx, img["psql"].Img)
	if err != nil {
		return err
	}
	if !ctx.DryRun() {
		// wait for readiness check
		if err := db.Ready(); err != nil {
			return err
		}

		// create DB names
		dbNames := []string{"chainlink_bootstrap"}
		for i := 0; i < nodeNum; i++ {
			dbNames = append(dbNames, fmt.Sprintf("chainlink_%d", i))
		}

		// create DBs
		for _, n := range dbNames {
			if err := db.Create(n); err != nil {
				return err
			}
		}
	}

	// start EAs
	adapters := []client.BridgeTypeAttributes{}
	for i, ea := range eas {
		a, err := adapter.New(ctx, img[ea], i)
		if err != nil {
			return err
		}
		adapters = append(adapters, a)
	}

	// start chainlink nodes
	nodes := map[string]*chainlink.Node{}
	for i := 0; i <= nodeNum; i++ {
		// start container
		cl, err := chainlink.New(ctx, img["chainlink"], db.Port, i)
		if err != nil {
			return err
		}
		nodes[cl.Name] = &cl // store in map
	}

	if !ctx.DryRun() {
		for _, cl := range nodes {
			// wait for readiness check
			if err := cl.Ready(); err != nil {
				return err
			}

			// delete all jobs if any exist
			if err := cl.DeleteAllJobs(); err != nil {
				return err
			}

			// add adapters to CL node
			for _, a := range adapters {
				if err := cl.AddBridge(a.Name, a.URL); err != nil {
					return err
				}
			}
		}
	}

	if !ctx.DryRun() {
		// fetch keys from relays
		for k := range nodes {
			if err := nodes[k].GetKeys(); err != nil {
				return err
			}
		}

		// upload contracts
		if err = deployer.Load(); err != nil {
			return err
		}
		// deploy LINK
		if err = deployer.DeployLINK(); err != nil {
			return err
		}

		// deploy OCR2 contract (w/ dummy access controller addresses)
		if err = deployer.DeployOCR(); err != nil {
			return err
		}

		// transfer tokens to OCR2 contract
		if err = deployer.TransferLINK(); err != nil {
			return err
		}

		// set OCR2 config
		var keys []map[string]string
		for k, v := range nodes {
			// skip if bootstrap node
			if k == "chainlink-bootstrap" {
				continue
			}
			keys = append(keys, v.Keys)
		}
		if err = deployer.InitOCR(keys); err != nil {
			return err
		}

		// create job specs
		i := 0
		for k := range nodes {
			// create specs + add to CL node
			ea := eas[i%len(eas)]
			msg := utils.LogStatus(fmt.Sprintf("Adding job spec to '%s' with '%s' EA", k, ea))

			// TODO: swap this for proper ocr2 spec
			spec := &client.OCR2TaskJobSpec{
				Name:        "local testing job",
				ContractID:  deployer.OCR2Address(),
				Relay:       "solana", // TODO: pass custom relay name
				RelayConfig: "{}", // TODO: pass custom relay configs
				P2PPeerID:   nodes[k].Keys["P2PID"],
				P2PBootstrapPeers: []client.P2PData{
					client.P2PData{
						PeerID:   nodes["chainlink-bootstrap"].Keys["P2PID"],
						RemoteIP: nodes["chainlink-bootstrap"].P2P,
					},
				},
				IsBootstrapPeer: k == "chainlink-bootstrap",
				OCRKeyBundleID: nodes[k].Keys["OCRKeyID"],
				TransmitterID:  "", // TODO: needs to be filled in depending on network
				ObservationSource:     obsSource(ea),
				JuelsPerFeeCoinSource: juelsObsSource(ea),
			}
			_, err = nodes[k].Call.CreateJob(spec)
			if msg.Check(err) != nil {
				return err
			}
			i++
		}
	}

	return nil
}
