GENERATOR_PKG = Mtools/generator/v1alpha/cre_metadata.proto=github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/pb
GENERATOR_PKG = Mtools/generator/v1alpha/cre_metadata.proto=github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/pb
VALUES_PKG = Mvalues/v1/values.proto=github.com/smartcontractkit/chainlink-common/pkg/values/pb
SDK_PKG    = Msdk/v1alpha/sdk.proto=github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb

.PHONY: generate
generate:
	protoc -I../../../../../../proto_vendor/chainlink-protos/cre \
	       --go_out=. \
	       --go_opt=paths=source_relative \
	       --go_opt=$(GENERATOR_PKG) \
	       --go_opt=$(VALUES_PKG) \
	       --go_opt=$(SDK_PKG) \
	       tools/generator/v1alpha/cre_metadata.proto
	mv tools/generator/v1alpha/cre_metadata.pb.go cre_metadata.pb.go
	rm -rf tools