package artifacts

import (
	"encoding/hex"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/sha3"
)

type ArtifactsTestSuite struct {
	suite.Suite

	lggr *slog.Logger
}

func (s *ArtifactsTestSuite) SetupSuite() {
	s.lggr = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *ArtifactsTestSuite) TearDownSuite() {
	// Cleanup code
}

func (s *ArtifactsTestSuite) TestArtifacts() {
	artifacts := NewWorkflowArtifacts(&Input{
		WorkflowOwner: "0x97f8a56d48290f35A23A074e7c73615E93f21885",
		WorkflowName:  "wf-test-1",
		WorkflowPath:  "./testdata/main.go",
		ConfigPath:    "",
	}, s.lggr)
	err := artifacts.Compile()
	s.NoError(err, "failed to compile workflow")

	err = artifacts.Prepare()
	s.NoError(err, "failed to prepare artifacts")

	b64EncodedBinaryData := artifacts.GetBinaryData()
	s.NotEmpty(b64EncodedBinaryData, "binary data should not be empty")

	s.lggr.Info("WorkflowCompiledBinary Size", "size", len(b64EncodedBinaryData))

	// Compare the keccak256 hash of the binary data with the keccak256 hash of the
	// base64 encoded binary from CRE-CLI
	expKeccak256Hash, err := hex.DecodeString("7385ae61173b2886a12250b508e2d361af0f6a40b6d0dda153fd4c20cb7e0c10")
	s.NoError(err, "failed to decode expected keccak256 hash")
	keccak256FromSha3Lib := sha3.NewLegacyKeccak256()
	keccak256FromSha3Lib.Write(b64EncodedBinaryData)
	actualKeccak256Hash := keccak256FromSha3Lib.Sum(nil)
	s.Equal(actualKeccak256Hash, expKeccak256Hash)

	err = artifacts.Prepare()
	s.NoError(err, "failed to prepare artifacts")

	base64EncodedBinaryData := artifacts.GetBinaryData()
	// Compare if the compiled WASM binary is the same as the CRE-CLI output
	s.Equal(563960, len(base64EncodedBinaryData), "binary data size should be same as CRE-CLI output")
	s.Equal("m6kIGWtQvQYAALADgIPzAADQt3oerhGwcQAA8PV/uALY1AHwSGF6AACo9vMAqKqqqqqqqqqqqmsSAsIx",
		string(base64EncodedBinaryData[0:80]))
	s.Equal("iBgWTlBKb+iQflnikISk/ALWkBmXnTIkLiWJYdvkGOrbpWD+9ybhB5ISI+JHMO5tcviErMSoqk5p1CY=",
		string(base64EncodedBinaryData[len(base64EncodedBinaryData)-80:]))

	s.Equal("", string(artifacts.GetConfigData()))

	s.Equal("009380535a54d6e2d4d7125fdf3befbf24c1637e836587fc07f45d5d27451ad0", artifacts.GetWorkflowID())
}

func TestArtifactsTestSuite(t *testing.T) {
	suite.Run(t, new(ArtifactsTestSuite))
}
