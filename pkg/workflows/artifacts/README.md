## CRE Workflow Artifacts

Common abstraction for managing CRE Workflow Artifacts for different services/clients including CRE-CLI to

* Compile WASM workflows: For Typescript and Go CRE workflows
* Upload artifacts like workflow binaries and configs to storage service

## Testing

```
$ cd /path/to/chainlink-common/pkg/workflows/artifacts/
$ go test -v -count=1 ./
```

### Testdata management

To get the exact CRE-CLI workflow wasm binary data and its hash for testing comparison.

- Add a project.yaml to `./pkg/workflows/artifacts/project.yaml`

```
local-simulation:
  rpcs:
    - chain-name: ethereum-testnet-sepolia
      url: <ETH_SEPOLIA_RPC_URL>
staging:
  cre-cli:
    don-family: "<DON_FAMILY>"
  rpcs:
    - chain-name: ethereum-testnet-sepolia
      url: <ETH_SEPOLIA_RPC_URL>
```

- Add a workflow.yaml to `./pkg/workflows/artifacts/testdata/workflow.yaml`

```
local-simulation:
  user-workflow:
    workflow-name: "wf-test-1"
  workflow-artifacts:
    workflow-path: "./"
    config-path: ""
staging:
  user-workflow:
    workflow-name: "wf-test-1"
  workflow-artifacts:
    workflow-path: "./"
    config-path: ""
```

- Then run `cre workflow deploy` to compile workflow wasm binary from testdata folder to get the exact base64 encoded brotli compressed wasm binary

```
$ cre login
$ cd /path/to/chainlink-common/pkg/workflows/artifacts/
$ cre workflow deploy ./testdata
$ cat testdata/binary.wasm.br.b64 | cast keccak256
```