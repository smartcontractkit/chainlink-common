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
	expKeccak256Hash, err := hex.DecodeString("65831f453dd5524a67befcddd23b14d3b2ec0dc63a9993ba34a5f2ebab4cd937")
	s.NoError(err, "failed to decode expected keccak256 hash")
	keccak256FromSha3Lib := sha3.NewLegacyKeccak256()
	keccak256FromSha3Lib.Write(b64EncodedBinaryData)
	actualKeccak256Hash := keccak256FromSha3Lib.Sum(nil)
	s.Equal(actualKeccak256Hash, expKeccak256Hash)

	err = artifacts.Prepare()
	s.NoError(err, "failed to prepare artifacts")

	base64EncodedBinaryData := artifacts.GetBinaryData()
	// Compare if the compiled WASM binary is the same as the CRE-CLI output
	s.Equal(len(base64EncodedBinaryData), 564092, "binary data size should be same as CRE-CLI output")
	s.Equal(string(base64EncodedBinaryData[0:80]),
		"m1AJGVtQvQYAAACA3cF5AACorT8P1wjYOAAAUP0frgDWZABEMD0AAFTfeQBUVVVVVVVVVVV1S0JAOGbb")
	s.Equal(string(base64EncodedBinaryData[len(base64EncodedBinaryData)-80:]),
		"iGHhBKX0vhPpt2UOzcgaNJjVFKYV5wxNy8li2JYlmssCk9P/bZPwE1mZoekjGceWJYaMoszwy63SqE0=")

	s.Equal(string(artifacts.GetConfigData()), "")

	s.Equal(artifacts.GetWorkflowID(), "00350e368a4b6732b5d819953984a8d85a2af6d132c1407c400130bbde877906")
}

func TestArtifactsTestSuite(t *testing.T) {
	suite.Run(t, new(ArtifactsTestSuite))
}
