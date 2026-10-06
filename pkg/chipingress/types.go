package chipingress

import (
	"errors"
	"fmt"
	"strings"

	cepb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	ce "github.com/cloudevents/sdk-go/v2"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
)

// IdempotencyKeyAttr is the CloudEvent extension attribute name for a per-event idempotency key.
// Set it via the attributes map of NewEvent.
// When the event is emitted over Kafka using the CloudEvents Kafka binding, extensions become
// Kafka headers named "ce_<name>" (e.g., ce_idempotencykey), enabling downstream deduplication.
const IdempotencyKeyAttr = "idempotencykey"

// PartitionKeyAttr is the CloudEvent extension attribute name for an opaque, caller-defined
// partition key (INFOPLAT-2274). chip-ingress does not parse it or use it as the Kafka record
// key; it is forwarded unchanged as the ce_partitionkey header for downstream sinks. By
// convention it may be a colon-delimited list. Set it via the attributes map of NewEvent.
const PartitionKeyAttr = "partitionkey"

// OrderKeyField names a server-resolved field used as (part of) the Kafka ordering key.
// The set is intentionally closed: adding a value is a hot-partition risk and
// requires a PR + INFOPLAT team approval (keep in sync with chip-ingress allowlist).
type OrderKeyField string

// OrderKey is a validated, colon-delimited list of OrderKeyField names (e.g. "csapublickey" or
// "csapublickey:workflowid"). Build it with NewOrderKey. The client never sends a key value;
// chip-ingress resolves each named field and joins the values with ":" to form the Kafka key.
type OrderKey string

const (
	// OrderKeyAttr is the CloudEvent extension attribute name carrying the OrderKey.
	// It is forwarded as the ce_orderkey header with the raw field-name list.
	OrderKeyAttr = "orderkey"

	// OrderKeyFieldCSAPublicKey orders events by the authenticated CSA public key.
	OrderKeyFieldCSAPublicKey OrderKeyField = "csapublickey"

	// orderKeyDelimiter separates field names inside an OrderKey.
	orderKeyDelimiter = ":"

	// maxOrderKeyFields is the maximum number of fields in one OrderKey (matches chip-ingress).
	maxOrderKeyFields = 4
)

var validOrderKeyFields = map[OrderKeyField]struct{}{
	OrderKeyFieldCSAPublicKey: {},
}

// NewOrderKey builds an OrderKey from one or more allowlisted fields, joined with ":".
// It errors when no fields are given, a field is not allowlisted, a field repeats, or more
// than 4 fields are given.
func NewOrderKey(fields ...OrderKeyField) (OrderKey, error) {
	names := make([]string, len(fields))
	for i, f := range fields {
		names[i] = string(f)
	}
	return parseOrderKey(strings.Join(names, orderKeyDelimiter), len(fields))
}

// validate re-checks an OrderKey (split on ":" and check each name), so a hand-converted
// OrderKey("...") cannot bypass NewOrderKey.
func (k OrderKey) validate() error {
	_, err := parseOrderKey(string(k), len(strings.Split(string(k), orderKeyDelimiter)))
	return err
}

func parseOrderKey(joined string, n int) (OrderKey, error) {
	if n == 0 || joined == "" {
		return "", errors.New("order key requires at least one field")
	}
	if n > maxOrderKeyFields {
		return "", fmt.Errorf("order key has %d fields, max is %d", n, maxOrderKeyFields)
	}
	seen := make(map[string]struct{}, n)
	for _, name := range strings.Split(joined, orderKeyDelimiter) {
		if _, ok := validOrderKeyFields[OrderKeyField(name)]; !ok {
			return "", fmt.Errorf("unsupported order key field %q", name)
		}
		if _, dup := seen[name]; dup {
			return "", fmt.Errorf("duplicate order key field %q", name)
		}
		seen[name] = struct{}{}
	}
	return OrderKey(joined), nil
}

// resourceAttributeHeaders is the closed whitelist of producer resource attributes sent as gRPC
// metadata, mapping each attribute key (lowercased — SanitizeMetadataHeaders matches
// case-insensitively) to the fixed chainlink-* metadata header name it travels under. For example
// csa_public_key is sent as chainlink-resource-csa-public-key.
//
// It is the wire contract with chip-ingress, which reads exactly these header names and forwards
// them onto every Kafka record a request produces under resource_<original attribute key> (e.g.
// chainlink-resource-service-name becomes resource_service.name). The set is
// closed on purpose: because no operator-defined key can ever become a header name, no attribute
// can shadow a reserved gRPC metadata key such as the CSA auth token's, and the server needs no
// deny-list to keep resource attributes away from "ce_" or identity headers it derives from the
// verified auth token.
//
// The same mapping exists in chip-ingress (chip-ingress/internal/constants). Duplicating it across
// repositories is deliberate, matching how authHeaderKey is already spelled in both pkg/beholder
// and pkg/chipingress; the two must stay in sync or forwarding silently stops.
var resourceAttributeHeaders = map[string]string{
	"csa_public_key":   "chainlink-resource-csa-public-key",
	"deployed_by":      "chainlink-resource-deployed-by",
	"donid":            "chainlink-resource-don-id",
	"host.name":        "chainlink-resource-host-name",
	"internal_node_id": "chainlink-resource-internal-node-id",
	"node_id":          "chainlink-resource-node-id",
	"platformenv":      "chainlink-resource-platform-env",
	"service.name":     "chainlink-resource-service-name",
	"service.sha":      "chainlink-resource-service-sha",
	"zone":             "chainlink-resource-zone",
}

type (
	// Cloudevents types
	CloudEvent   = ce.Event
	CloudEventPb = cepb.CloudEvent

	// Client
	ChipIngressClient              = pb.ChipIngressClient
	ChipIngress_StreamEventsClient = pb.ChipIngress_StreamEventsClient

	// Message types
	CloudEventBatch      = pb.CloudEventBatch
	EmptyRequest         = pb.EmptyRequest
	PublishErrorCode     = pb.PublishErrorCode
	PingResponse         = pb.PingResponse
	PublishOptions       = pb.PublishOptions
	PublishResponse      = pb.PublishResponse
	PublishResult        = pb.PublishResult
	PublishError         = pb.PublishError
	StreamEventsRequest  = pb.StreamEventsRequest
	StreamEventsResponse = pb.StreamEventsResponse
)
