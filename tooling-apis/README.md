# tooling-protos

This repo holds the protobuf definitions and generated Go SDKs for the `job-distributor` and `orchestrator` services.

Go applications should depend on the necessary SDKs directly

```bash
$ go get github.com/smartcontractkit/chainlink-common/tooling-apis/job-distributor/app
```

Other applications may build their own SDKs directly from the provided protobufs.

## dependencies

Dependencies are managed via [asdf](https://asdf-vm.com/guide/getting-started.html).

```bash
$ asdf install
```

## generating GO SDKs

Generate the GO SDKs to implement gRPC services or clients via [task](https://taskfile.dev/installation/)

```bash
$ task generate
```
