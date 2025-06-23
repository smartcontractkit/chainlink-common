# chainlink-common [![Go Reference](https://pkg.go.dev/badge/github.com/smartcontractkit/chainlink-common.svg)](https://pkg.go.dev/github.com/smartcontractkit/chainlink-common)

Chain-agnostic SDK for implementing Chainlink Services, like Capabilities, Chain Relayers, Product Plugins, and Workflows.

## Updating chainlink-protos for the CRE proto generation

Use the command 
`make update-clprotos COMMIT=<commit>` to update the `chainlink-protos` dependency in the CRE. Replace `<commit>` with the commit hash or tag you want to update to.