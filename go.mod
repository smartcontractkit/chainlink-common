module github.com/smartcontractkit/chainlink-common

go 1.23.3

require (
	github.com/XSAM/otelsql v0.29.0
	github.com/andybalholm/brotli v1.1.1
	github.com/atombender/go-jsonschema v0.16.1-0.20240916205339-a74cd4e2851c
	github.com/bytecodealliance/wasmtime-go/v28 v28.0.0
	github.com/confluentinc/confluent-kafka-go/v2 v2.3.0
	github.com/dominikbraun/graph v0.23.0
	github.com/fxamacker/cbor/v2 v2.7.0
	github.com/go-json-experiment/json v0.0.0-20231102232822-2e55bd4e08b0
	github.com/go-playground/validator/v10 v10.4.1
	github.com/go-viper/mapstructure/v2 v2.2.1
	github.com/google/go-cmp v0.6.0
	github.com/google/uuid v1.6.0
	github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus v1.0.1
	github.com/hashicorp/consul/sdk v0.16.1
	github.com/hashicorp/go-hclog v1.6.3
	github.com/hashicorp/go-plugin v1.6.2
	github.com/iancoleman/strcase v0.3.0
	github.com/invopop/jsonschema v0.12.0
	github.com/jackc/pgx/v4 v4.18.3
	github.com/jmoiron/sqlx v1.4.0
	github.com/jonboulle/clockwork v0.4.0
	github.com/jpillora/backoff v1.0.0
	github.com/lib/pq v1.10.9
	github.com/linkedin/goavro/v2 v2.12.0
	github.com/marcboeker/go-duckdb v1.8.3
	github.com/pelletier/go-toml/v2 v2.2.0
	github.com/pkcll/opentelemetry-collector-contrib/receiver/prometheusreceiver v0.115.7
	github.com/prometheus/client_golang v1.20.5
	github.com/prometheus/common v0.62.0
	github.com/riferrei/srclient v0.5.4
	github.com/santhosh-tekuri/jsonschema/v5 v5.2.0
	github.com/scylladb/go-reflectx v1.0.1
	github.com/shopspring/decimal v1.4.0
	github.com/smartcontractkit/grpc-proxy v0.0.0-20240830132753-a7e17fec5ab7
	github.com/smartcontractkit/libocr v0.0.0-20241007185508-adbe57025f12
	github.com/stretchr/testify v1.10.0
	github.com/zeebo/assert v1.3.0
	go.opentelemetry.io/collector/component v0.119.0
	go.opentelemetry.io/collector/component/componenttest v0.119.0
	go.opentelemetry.io/collector/config/configauth v0.119.0
	go.opentelemetry.io/collector/config/configcompression v1.25.0
	go.opentelemetry.io/collector/config/configgrpc v0.118.0
	go.opentelemetry.io/collector/config/configopaque v1.25.0
	go.opentelemetry.io/collector/config/configretry v1.24.0
	go.opentelemetry.io/collector/config/configtelemetry v0.119.0
	go.opentelemetry.io/collector/config/configtls v1.25.0
	go.opentelemetry.io/collector/confmap v1.25.0
	go.opentelemetry.io/collector/consumer v1.25.0
	go.opentelemetry.io/collector/consumer/consumertest v0.119.0
	go.opentelemetry.io/collector/exporter v0.118.0
	go.opentelemetry.io/collector/exporter/exportertest v0.118.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.118.0
	go.opentelemetry.io/collector/pdata v1.25.0
	go.opentelemetry.io/collector/receiver v0.119.0
	go.opentelemetry.io/collector/receiver/receivertest v0.119.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.58.0
	go.opentelemetry.io/otel v1.34.0
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.10.0
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.7.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.32.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.32.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.33.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.33.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.7.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.32.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.31.0
	go.opentelemetry.io/otel/log v0.10.0
	go.opentelemetry.io/otel/metric v1.34.0
	go.opentelemetry.io/otel/sdk v1.34.0
	go.opentelemetry.io/otel/sdk/log v0.10.0
	go.opentelemetry.io/otel/sdk/metric v1.34.0
	go.opentelemetry.io/otel/trace v1.34.0
	go.uber.org/goleak v1.3.0
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.32.0
	golang.org/x/exp v0.0.0-20250128182459-e0ece0dbea4c
	golang.org/x/tools v0.29.0
	gonum.org/v1/gonum v0.15.1
	google.golang.org/grpc v1.70.0
	google.golang.org/protobuf v1.36.4
	gopkg.in/yaml.v2 v2.4.0
	sigs.k8s.io/yaml v1.4.0
)

require (
	cloud.google.com/go/auth v0.14.1 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.7 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.17.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.8.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.10.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5 v5.7.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4 v4.3.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.3.3 // indirect
	github.com/Code-Hex/go-generics-cache v1.5.1 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/alecthomas/units v0.0.0-20240927000941-0f3dac36c52b // indirect
	github.com/apache/arrow-go/v18 v18.0.0 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/aws/aws-sdk-go v1.55.6 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cncf/xds/go v0.0.0-20250121191232-2f005788dc42 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dennwc/varint v1.0.0 // indirect
	github.com/digitalocean/godo v1.136.0 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/docker v27.5.1+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emicklei/go-restful/v3 v3.12.1 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.32.4 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-resty/resty/v2 v2.16.5 // indirect
	github.com/go-zookeeper/zk v1.0.4 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/goccy/go-yaml v1.12.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/google/flatbuffers v24.3.25+incompatible // indirect
	github.com/google/gnostic-models v0.6.9 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.4 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/gophercloud/gophercloud v1.14.1 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/grafana/regexp v0.0.0-20240518133315-a468a5bfb3bc // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.1.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.25.1 // indirect
	github.com/hashicorp/consul/api v1.31.0 // indirect
	github.com/hashicorp/cronexpr v1.1.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-metrics v0.5.4 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/nomad/api v0.0.0-20250205130000-9fef959daf79 // indirect
	github.com/hashicorp/serf v0.10.2 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/hetznercloud/hcloud-go/v2 v2.19.1 // indirect
	github.com/ionos-cloud/sdk-go/v6 v6.3.2 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgtype v1.14.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/providers/confmap v0.1.0 // indirect
	github.com/knadh/koanf/v2 v2.1.2 // indirect
	github.com/kolo/xmlrpc v0.0.0-20220921171641-a4b6fa1dd06b // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/linode/linodego v1.47.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/miekg/dns v1.1.63 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.1-0.20231216201459-8508981c8b6c // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/user v0.3.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatautil v0.119.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus v0.119.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/ovh/go-ovh v1.6.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pkcll/prometheus v0.54.3-0.20250205152829-6e977873d1b5 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common/sigv4 v0.1.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/prometheus/prometheus v0.54.1 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/sanity-io/litter v1.5.5 // indirect
	github.com/scaleway/scaleway-sdk-go v1.0.0-beta.32 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/vultr/govultr/v2 v2.17.2 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/collector/client v1.25.0 // indirect
	go.opentelemetry.io/collector/config/confighttp v0.119.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.24.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.119.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror/xconsumererror v0.118.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.119.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper/xexporterhelper v0.118.0 // indirect
	go.opentelemetry.io/collector/exporter/xexporter v0.118.0 // indirect
	go.opentelemetry.io/collector/extension v0.119.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.119.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.118.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.25.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.119.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.119.0 // indirect
	go.opentelemetry.io/collector/pipeline/xpipeline v0.118.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.119.0 // indirect
	go.opentelemetry.io/collector/semconv v0.119.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.59.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.33.0 // indirect
	go.opentelemetry.io/proto/otlp v1.5.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/mod v0.23.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/oauth2 v0.26.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/term v0.29.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	golang.org/x/time v0.10.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/api v0.220.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250204164813-702378808489 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250204164813-702378808489 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.32.1 // indirect
	k8s.io/apimachinery v0.32.1 // indirect
	k8s.io/client-go v0.32.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20241212222426-2c72e554b1e7 // indirect
	k8s.io/utils v0.0.0-20241210054802-24370beab758 // indirect
	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.5.0 // indirect
)
