package chipingress

import (
	cepb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	ce "github.com/cloudevents/sdk-go/v2"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
)

// IdempotencyKeyAttr is the CloudEvent extension attribute name for a per-event idempotency key.
// Set it via the attributes map of NewEvent.
// When the event is emitted over Kafka using the CloudEvents Kafka binding, extensions become
// Kafka headers named "ce_<name>" (e.g., ce_idempotencykey), enabling downstream deduplication.
const IdempotencyKeyAttr = "idempotencykey"

// ResourceAttributeHeaders is the closed whitelist of producer resource attributes sent as gRPC
// metadata, mapping each attribute key (lowercased — SanitizeMetadataHeaders matches
// case-insensitively) to the fixed chainlink-* metadata header name it travels under. For example
// csa_public_key is sent as chainlink-csa-public-key.
//
// It is the wire contract with chip-ingress, which reads exactly these header names and forwards
// them onto every Kafka record a request produces under resource_<original attribute key> (e.g.
// chainlink-service-name becomes resource_service.name). The set is
// closed on purpose: because no operator-defined key can ever become a header name, no attribute
// can shadow a reserved gRPC metadata key such as the CSA auth token's, and the server needs no
// deny-list to keep resource attributes away from "ce_" or identity headers it derives from the
// verified auth token.
//
// The same mapping exists in chip-ingress (chip-ingress/internal/constants). Duplicating it across
// repositories is deliberate, matching how authHeaderKey is already spelled in both pkg/beholder
// and pkg/chipingress; the two must stay in sync or forwarding silently stops.
var ResourceAttributeHeaders = map[string]string{
	"csa_public_key":   "chainlink-csa-public-key",
	"deployed_by":      "chainlink-deployed-by",
	"donid":            "chainlink-don-id",
	"host.name":        "chainlink-host-name",
	"internal_node_id": "chainlink-internal-node-id",
	"node_id":          "chainlink-node-id",
	"platformenv":      "chainlink-platform-env",
	"service.name":     "chainlink-service-name",
	"service.sha":      "chainlink-service-sha",
	"zone":             "chainlink-zone",
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
