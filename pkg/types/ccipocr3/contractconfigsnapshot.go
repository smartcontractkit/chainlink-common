package ccipocr3

// ChainConfigSnapshot is a legacy type used in chain accessor's GetAllConfigLegacySnapshot() function. This will
// eventually be replaced by a more explicit approach using an interface returned by a future GetAllConfig() function.
type ChainConfigSnapshot struct {
	Offramp   OfframpConfig
	RMNProxy  RMNProxyConfig
	RMNRemote RMNRemoteConfig
	FeeQuoter FeeQuoterConfig
	OnRamp    OnRampConfig
	Router    RouterConfig
	CurseInfo CurseInfo
}

type OnRampConfig struct {
	DynamicConfig   GetOnRampDynamicConfigResponse
	DestChainConfig OnRampDestChainConfig
}

type FeeQuoterConfig struct {
	StaticConfig FeeQuoterStaticConfig
}

// FeeQuoterStaticConfig is used to parse the response from the feeQuoter contract's getStaticConfig method.
// See: https://github.com/smartcontractkit/ccip/blob/a3f61f7458e4499c2c62eb38581c60b4942b1160/contracts/src/v0.8/ccip/FeeQuoter.sol#L946
//
//nolint:lll // It's a URL.
type FeeQuoterStaticConfig struct {
	MaxFeeJuelsPerMsg  BigInt `json:"maxFeeJuelsPerMsg"`
	LinkToken          []byte `json:"linkToken"`
	StalenessThreshold uint32 `json:"stalenessThreshold"`
}

type RMNRemoteConfig struct {
	DigestHeader    RMNDigestHeader
	VersionedConfig VersionedConfig
}

type OfframpConfig struct {
	CommitLatestOCRConfig OCRConfigResponse
	ExecLatestOCRConfig   OCRConfigResponse
	StaticConfig          OffRampStaticChainConfig
	DynamicConfig         OffRampDynamicChainConfig
}

// sourceChainConfig is used to parse the response from the offRamp contract's getSourceChainConfig method.
// See: https://github.com/smartcontractkit/ccip/blob/a3f61f7458e4499c2c62eb38581c60b4942b1160/contracts/src/v0.8/ccip/offRamp/OffRamp.sol#L94
//
//nolint:lll // It's a URL.
type SourceChainConfig struct {
	Router                    []byte // local router
	IsEnabled                 bool
	IsRMNVerificationDisabled bool
	MinSeqNr                  uint64
	OnRamp                    UnknownAddress
}

type RouterConfig struct {
	WrappedNativeAddress Bytes
}

type RMNProxyConfig struct {
	RemoteAddress []byte
}

type RMNDigestHeader struct {
	DigestHeader Bytes32
}

type OCRConfigResponse struct {
	OCRConfig OCRConfig
}

type OCRConfig struct {
	ConfigInfo   ConfigInfo
	Signers      [][]byte
	Transmitters [][]byte
}

type ConfigInfo struct {
	ConfigDigest                   Bytes32
	F                              uint8
	N                              uint8
	IsSignatureVerificationEnabled bool
}

type RMNCurseResponse struct {
	CursedSubjects [][16]byte
}

// OffRampStaticChainConfig is used to parse the response from the offRamp contract's getStaticConfig method.
// See: <chainlink repo>/contracts/src/v0.8/ccip/offRamp/OffRamp.sol:StaticConfig
type OffRampStaticChainConfig struct {
	ChainSelector        ChainSelector `json:"chainSelector"`
	GasForCallExactCheck uint16        `json:"gasForCallExactCheck"`
	RmnRemote            []byte        `json:"rmnRemote"`
	TokenAdminRegistry   []byte        `json:"tokenAdminRegistry"`
	NonceManager         []byte        `json:"nonceManager"`
}

// OffRampDynamicChainConfig maps to DynamicConfig in OffRamp.sol
type OffRampDynamicChainConfig struct {
	FeeQuoter                               []byte `json:"feeQuoter"`
	PermissionLessExecutionThresholdSeconds uint32 `json:"permissionLessExecutionThresholdSeconds"`
	IsRMNVerificationDisabled               bool   `json:"isRMNVerificationDisabled"`
	MessageInterceptor                      []byte `json:"messageInterceptor"`
}

// OnRampDynamicConfig - See DynamicChainConfig in OnRamp.sol
type OnRampDynamicConfig struct {
	FeeQuoter              []byte `json:"feeQuoter"`
	ReentrancyGuardEntered bool   `json:"reentrancyGuardEntered"`
	MessageInterceptor     []byte `json:"messageInterceptor"`
	FeeAggregator          []byte `json:"feeAggregator"`
	AllowListAdmin         []byte `json:"allowListAdmin"`
}

// GetOnRampDynamicConfigResponse wraps the OnRampDynamicConfig this way to map to on-chain return type which is a named struct
// https://github.com/smartcontractkit/chainlink/blob/12af1de88238e0e918177d6b5622070417f48adf/contracts/src/v0.8/ccip/onRamp/OnRamp.sol#L328
//
//nolint:lll
type GetOnRampDynamicConfigResponse struct {
	DynamicConfig OnRampDynamicConfig `json:"dynamicConfig"`
}

// OnRampDestChainConfig - See DestChainConfig in OnRamp.sol
type OnRampDestChainConfig struct {
	SequenceNumber   uint64 `json:"sequenceNumber"`
	AllowListEnabled bool   `json:"allowListEnabled"`
	Router           []byte `json:"router"`
}

// Signer is used to parse the response from the RMNRemote contract's getVersionedConfig method.
// See: https://github.com/smartcontractkit/ccip/blob/ccip-develop/contracts/src/v0.8/ccip/rmn/RMNRemote.sol#L42-L45
type Signer struct {
	OnchainPublicKey []byte `json:"onchainPublicKey"`
	NodeIndex        uint64 `json:"nodeIndex"`
}

// Config is used to parse the response from the RMNRemote contract's getVersionedConfig method.
// See: https://github.com/smartcontractkit/ccip/blob/ccip-develop/contracts/src/v0.8/ccip/rmn/RMNRemote.sol#L49-L53
type Config struct {
	RMNHomeContractConfigDigest Bytes32  `json:"rmnHomeContractConfigDigest"`
	Signers                     []Signer `json:"signers"`
	FSign                       uint64   `json:"fSign"` // previously: MinSigners
}

// VersionedConfig is used to parse the response from the RMNRemote contract's getVersionedConfig method.
// See: https://github.com/smartcontractkit/ccip/blob/ccip-develop/contracts/src/v0.8/ccip/rmn/RMNRemote.sol#L167-L169
type VersionedConfig struct {
	Version uint32 `json:"version"`
	Config  Config `json:"config"`
}
