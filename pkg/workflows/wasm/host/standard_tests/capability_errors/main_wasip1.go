package main

import (
	"fmt"

	caperrors "github.com/smartcontractkit/chainlink-common/pkg/capabilities/errors"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func main() {
	for _, code := range caperrors.AllErrorCodes {
		err := rawsdk.DoRequestErr(
			"basic-test-action@1.0.0",
			"PerformAction",
			sdk.Mode_MODE_DON,
			&basicaction.Inputs{InputThing: true},
		)
		if err == nil {
			rawsdk.SendError(fmt.Errorf("expected capability error for code %s", code))
		}

		capErr := caperrors.DeserializeErrorFromString(err.Error())
		if capErr.Code() != code {
			rawsdk.SendError(fmt.Errorf("expected error code %s, got %s", code, capErr.Code()))
		}
		if capErr.Origin() != caperrors.OriginSystem {
			rawsdk.SendError(fmt.Errorf("expected system origin for code %s, got %s", code, capErr.Origin()))
		}
		if capErr.Visibility() != caperrors.VisibilityPublic {
			rawsdk.SendError(fmt.Errorf("expected public visibility for code %s, got %s", code, capErr.Visibility()))
		}
	}

	output := &basicaction.Outputs{}
	rawsdk.DoRequest(
		"basic-test-action@1.0.0",
		"PerformAction",
		sdk.Mode_MODE_DON,
		&basicaction.Inputs{InputThing: true},
		output,
	)
	if output.AdaptedThing != "Done" {
		rawsdk.SendError(fmt.Errorf("expected Done response, got %s", output.AdaptedThing))
	}

	rawsdk.SendResponse("Done")
}
