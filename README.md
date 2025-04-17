If you're reading this you're on my branch to see the Go SDK.

Example workflow is in pkg/workflows/examples/bitgo/workflow, it's a good entry point

Code generation isn't fully plugged into the makefile yet. 

You need to add pkg/capabilities/protoc to your PATH to get the shell scripts below to work. 

pkg/capabilities/stubs/gen.sh generates the testing capabilities. This one should generate with make generate.
It's faster to use this than to run make generate. It also automatically re-compiles the protoc plugin for you.

pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/gen.sh generates the ones for the workflow. 
The folder seems to be excluded from make generate, so you need to run it manually. It also automatically re-compiles the protoc plugin for you.

# chainlink-common [![Go Reference](https://pkg.go.dev/badge/github.com/smartcontractkit/chainlink-common.svg)](https://pkg.go.dev/github.com/smartcontractkit/chainlink-common)

Chain-agnostic SDK for implementing Chainlink Services, like Capabilities, Chain Relayers, Product Plugins, and Workflows.
