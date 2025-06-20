PROTO_IMPORTS := \
	Mtools/generator/v1alpha/cre_metadata.proto=github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/pb \
	Mvalues/v1/values.proto=github.com/smartcontractkit/chainlink-common/pkg/values/pb \
	Msdk/v1alpha/sdk.proto=github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb

define gen
--$1_out=. \
$(foreach m,$(PROTO_IMPORTS),--$1_opt=$(m))
endef

.PHONY: generate
generate:
	protoc -I../../../../../../../proto_vendor/chainlink-protos/cre \
		  -I. \
	       --go_opt=paths=source_relative \
	       $(call gen,go) \
	       --plugin=protoc-gen-cre=../../../protoc-gen-cre \
	       $(call gen,cre) \
	       basic_trigger.proto