package proxy

// The types below mirror libocr's networking.PeerGroupFactory / PeerGroup /
// Stream, but are defined locally so this package does not import
// libocr/networking. That package transitively pulls in go-ethereum (via the
// OCR1 offchainreporting/types), which chainlink-common intentionally avoids.
//
// Callers that own a real libocr peer (e.g. the proxy binary) adapt
// networking.PeerGroupFactory to PeerGroupFactory; consumers that need a
// networking.PeerGroupFactory (e.g. core's Don2DonSharedPeer) adapt the
// PeerGroupFactory returned by the client back to the libocr interface. Both
// adapters are trivial because the method sets are identical.

// PeerGroupFactory creates PeerGroups for DON-to-DON communication.
type PeerGroupFactory interface {
	NewPeerGroup(configDigest [32]byte, peerIDs []string, bootstrappers []BootstrapperInfo) (PeerGroup, error)
}

// PeerGroup is a discovery+messaging group. Streams opened within it are
// automatically closed when the group is closed.
type PeerGroup interface {
	NewStream(remotePeerID string, args StreamArgs) (PeerGroupStream, error)
	Close() error
}

// PeerGroupStream is a single bidirectional stream to one remote peer.
type PeerGroupStream interface {
	SendMessage(data []byte)
	ReceiveMessages() <-chan []byte
	Close() error
}

// BootstrapperInfo identifies a bootstrapper peer and its addresses.
type BootstrapperInfo struct {
	PeerID string
	Addrs  []string
}

// StreamArgs mirrors networking.NewStreamArgs1.
type StreamArgs struct {
	StreamName         string
	OutgoingBufferSize int
	IncomingBufferSize int
	MaxMessageLength   int
	MessagesLimit      RateLimit
	BytesLimit         RateLimit
}

// RateLimit mirrors ragep2p.TokenBucketParams.
type RateLimit struct {
	Rate     float64
	Capacity uint32
}
