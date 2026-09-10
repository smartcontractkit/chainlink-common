package remote

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/smartcontractkit/libocr/ragep2p/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

// The converters below decode the registry's wire records (DON, node) into the
// domain types. They are written against minimal getter interfaces because the
// same records are carried by two protocols - the go-plugin registry protocol
// in pkg/loop/internal/pb and the plain-gRPC one in
// pkg/capabilities/registry/pb - whose generated types are distinct packages
// but share this shape. Decoding against the shape lets both transports use
// one conversion instead of drifting copies.

// DONProto is the wire shape of a DON as generated for both registry protocols.
type DONProto interface {
	GetId() uint32
	GetName() string
	GetMembers() [][]byte
	GetF() uint32
	GetConfigVersion() uint32
	GetFamilies() []string
	GetConfig() []byte
}

// DONWorkflowMetadataProto is the optional extension carrying a DON's workflow
// metadata. Only the plain-gRPC protocol declares it; the go-plugin protocol
// predates it, so its records decode with the fields at their zero values.
type DONWorkflowMetadataProto interface {
	GetIsPublic() bool
	GetAcceptsWorkflows() bool
}

// DONFromProto converts a wire DON to the Go type.
func DONFromProto[T DONProto](d T) capabilities.DON {
	if isNil(d) {
		return capabilities.DON{}
	}
	var members []types.PeerID
	for _, m := range d.GetMembers() {
		var peerID types.PeerID
		copy(peerID[:], m)
		members = append(members, peerID)
	}
	don := capabilities.DON{
		ID:            d.GetId(),
		Name:          d.GetName(),
		Members:       members,
		F:             uint8(d.GetF()),
		ConfigVersion: d.GetConfigVersion(),
		Families:      d.GetFamilies(),
		Config:        d.GetConfig(),
	}
	if wm, ok := any(d).(DONWorkflowMetadataProto); ok {
		don.IsPublic = wm.GetIsPublic()
		don.AcceptsWorkflows = wm.GetAcceptsWorkflows()
	}
	return don
}

// isNil reports whether d is a nil pointer (or other nil-able kind). The
// generated wire types are always pointers, and getters on a typed nil
// receiver answer zero values, which would otherwise read as a present but
// empty record.
func isNil[T any](d T) bool {
	if any(d) == nil {
		return true
	}
	v := reflect.ValueOf(d)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	}
	return false
}

// NodeProto is the wire shape of a node reply as generated for both registry
// protocols. The DONs a reply carries are decoded separately and passed in,
// since the generated types of the two protocols cannot be named by one
// interface. The peer ID and node operator ID are likewise excluded: the
// go-plugin protocol names its fields peerID and nodeOperatorID while the
// plain-gRPC one names them peer_id and node_operator_id, so their getters
// differ and they are read below with a type switch.
type NodeProto interface {
	GetSigner() []byte
	GetEncryptionPublicKey() []byte
}

// NodeFromProto converts a wire node reply to the Go type. workflowDON and
// capabilityDONs are the reply's DONs, still in wire form; they decode with
// [DONFromProto].
func NodeFromProto[T DONProto](n NodeProto, workflowDON T, capabilityDONs []T) (capabilities.Node, error) {
	if n == nil {
		return capabilities.Node{}, errors.New("nil node reply")
	}

	var peerIDBytes []byte
	switch p := any(n).(type) {
	case interface{ GetPeerID() []byte }:
		peerIDBytes = p.GetPeerID()
	case interface{ GetPeerId() []byte }:
		peerIDBytes = p.GetPeerId()
	}

	var peerID types.PeerID
	if len(peerIDBytes) != 0 {
		if len(peerIDBytes) != len(peerID) {
			return capabilities.Node{}, fmt.Errorf("invalid peer ID length %d", len(peerIDBytes))
		}
		copy(peerID[:], peerIDBytes)
	}

	var nodeOperatorID uint32
	switch op := any(n).(type) {
	case interface{ GetNodeOperatorID() uint32 }:
		nodeOperatorID = op.GetNodeOperatorID()
	case interface{ GetNodeOperatorId() uint32 }:
		nodeOperatorID = op.GetNodeOperatorId()
	}

	var signer, encryptionPublicKey [32]byte
	copy(signer[:], n.GetSigner())
	copy(encryptionPublicKey[:], n.GetEncryptionPublicKey())

	dons := make([]capabilities.DON, 0, len(capabilityDONs))
	for _, d := range capabilityDONs {
		dons = append(dons, DONFromProto(d))
	}

	var pid *types.PeerID
	if len(peerIDBytes) != 0 {
		pid = &peerID
	}

	return capabilities.Node{
		PeerID:              pid,
		NodeOperatorID:      nodeOperatorID,
		Signer:              signer,
		EncryptionPublicKey: encryptionPublicKey,
		WorkflowDON:         DONFromProto(workflowDON),
		CapabilityDONs:      dons,
	}, nil
}
