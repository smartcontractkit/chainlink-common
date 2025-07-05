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
	DynamicConfig   getOnRampDynamicConfigResponse
	DestChainConfig onRampDestChainConfig
}

type FeeQuoterConfig struct {
	StaticConfig feeQuoterStaticConfig
}

// feeQuoterStaticConfig is used to parse the response from the feeQuoter contract's getStaticConfig method.
// See: https://github.com/smartcontractkit/ccip/blob/a3f61f7458e4499c2c62eb38581c60b4942b1160/contracts/src/v0.8/ccip/FeeQuoter.sol#L946
//
//nolint:lll // It's a URL.
type feeQuoterStaticConfig struct {
	MaxFeeJuelsPerMsg  BigInt `json:"maxFeeJuelsPerMsg"`
	LinkToken          []byte `json:"linkToken"`
	StalenessThreshold uint32 `json:"stalenessThreshold"`
}

type RMNRemoteConfig struct {
	DigestHeader    rmnDigestHeader
	VersionedConfig versionedConfig
}

type OfframpConfig struct {
	CommitLatestOCRConfig OCRConfigResponse
	ExecLatestOCRConfig   OCRConfigResponse
	StaticConfig          offRampStaticChainConfig
	DynamicConfig         offRampDynamicChainConfig
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

type rmnDigestHeader struct {
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
	ConfigDigest                   [32]byte
	F                              uint8
	N                              uint8
	IsSignatureVerificationEnabled bool
}

type RMNCurseResponse struct {
	CursedSubjects [][16]byte
}

// offRampStaticChainConfig is used to parse the response from the offRamp contract's getStaticConfig method.
// See: <chainlink repo>/contracts/src/v0.8/ccip/offRamp/OffRamp.sol:StaticConfig
type offRampStaticChainConfig struct {
	ChainSelector        ChainSelector `json:"chainSelector"`
	GasForCallExactCheck uint16        `json:"gasForCallExactCheck"`
	RmnRemote            []byte        `json:"rmnRemote"`
	TokenAdminRegistry   []byte        `json:"tokenAdminRegistry"`
	NonceManager         []byte        `json:"nonceManager"`
}

// offRampDynamicChainConfig maps to DynamicConfig in OffRamp.sol
type offRampDynamicChainConfig struct {
	FeeQuoter                               []byte `json:"feeQuoter"`
	PermissionLessExecutionThresholdSeconds uint32 `json:"permissionLessExecutionThresholdSeconds"`
	IsRMNVerificationDisabled               bool   `json:"isRMNVerificationDisabled"`
	MessageInterceptor                      []byte `json:"messageInterceptor"`
}

// See DynamicChainConfig in OnRamp.sol
type onRampDynamicConfig struct {
	FeeQuoter              []byte `json:"feeQuoter"`
	ReentrancyGuardEntered bool   `json:"reentrancyGuardEntered"`
	MessageInterceptor     []byte `json:"messageInterceptor"`
	FeeAggregator          []byte `json:"feeAggregator"`
	AllowListAdmin         []byte `json:"allowListAdmin"`
}

// We're wrapping the onRampDynamicConfig this way to map to on-chain return type which is a named struct
// https://github.com/smartcontractkit/chainlink/blob/12af1de88238e0e918177d6b5622070417f48adf/contracts/src/v0.8/ccip/onRamp/OnRamp.sol#L328
//
//nolint:lll
type getOnRampDynamicConfigResponse struct {
	DynamicConfig onRampDynamicConfig `json:"dynamicConfig"`
}

// See DestChainConfig in OnRamp.sol
type onRampDestChainConfig struct {
	SequenceNumber   uint64 `json:"sequenceNumber"`
	AllowListEnabled bool   `json:"allowListEnabled"`
	Router           []byte `json:"router"`
}

// signer is used to parse the response from the RMNRemote contract's getVersionedConfig method.
// See: https://github.com/smartcontractkit/ccip/blob/ccip-develop/contracts/src/v0.8/ccip/rmn/RMNRemote.sol#L42-L45
type signer struct {
	OnchainPublicKey []byte `json:"onchainPublicKey"`
	NodeIndex        uint64 `json:"nodeIndex"`
}

// config is used to parse the response from the RMNRemote contract's getVersionedConfig method.
// See: https://github.com/smartcontractkit/ccip/blob/ccip-develop/contracts/src/v0.8/ccip/rmn/RMNRemote.sol#L49-L53
type config struct {
	RMNHomeContractConfigDigest Bytes32  `json:"rmnHomeContractConfigDigest"`
	Signers                     []signer `json:"signers"`
	FSign                       uint64   `json:"fSign"` // previously: MinSigners
}

// versionedConfig is used to parse the response from the RMNRemote contract's getVersionedConfig method.
// See: https://github.com/smartcontractkit/ccip/blob/ccip-develop/contracts/src/v0.8/ccip/rmn/RMNRemote.sol#L167-L169
type versionedConfig struct {
	Version uint32 `json:"version"`
	Config  config `json:"config"`
}
