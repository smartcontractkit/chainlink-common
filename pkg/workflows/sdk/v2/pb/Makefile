VALUES_PKG = Mvalues/v1/values.proto=github.com/smartcontractkit/chainlink-common/pkg/values/pb
SDK_PKG    = Msdk/v1alpha/sdk.proto=github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb

.PHONY: generate
generate:
	protoc -I../../../../../proto_vendor/chainlink-protos/cre \
	       --go_out=. \
	       --go_opt=paths=source_relative \
	       --go_opt=$(SDK_PKG) \
	       --go_opt=$(VALUES_PKG) \
	       sdk/v1alpha/sdk.proto
	mv sdk/v1alpha/sdk.pb.go sdk.pb.go
	rm -rf sdk