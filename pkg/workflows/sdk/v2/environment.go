package sdk

import (
	"io"
	"log/slog"

	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

type NodeEnvironment[C any] struct {
	Config    C
	LogWriter io.Writer
	Logger    *slog.Logger
}

type SecretsProvider interface {
	GetSecret(*sdkpb.SecretRequest) Promise[*sdkpb.Secret]
}

type ReportGenerator interface {
	GenerateReport(
		encodedPayload []byte,
		encoderName, signingAlgo, hashingAlgo string,
	) Promise[*sdkpb.ConsensusOutputs]
}

type Environment[C any] struct {
	NodeEnvironment[C]
	SecretsProvider
	ReportGenerator
}
