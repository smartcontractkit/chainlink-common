package ccipocr3

import (
	"fmt"
	"math/big"

	ccipocr3pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccipocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"
)

// Helper function to convert protobuf BigInt to big.Int
func pbBigIntToInt(b *ccipocr3pb.BigInt) *big.Int {
	if b == nil || len(b.Value) == 0 {
		return big.NewInt(0)
	}
	return new(big.Int).SetBytes(b.Value)
}

// Helper function to convert protobuf BigInt to ccipocr3.BigInt, preserving nil
func pbToBigInt(b *ccipocr3pb.BigInt) ccipocr3.BigInt {
	if b == nil || len(b.Value) == 0 {
		return ccipocr3.BigInt{Int: nil}
	}
	return ccipocr3.NewBigInt(new(big.Int).SetBytes(b.Value))
}

// Helper function to convert big.Int to protobuf BigInt
func intToPbBigInt(i *big.Int) *ccipocr3pb.BigInt {
	if i == nil {
		return &ccipocr3pb.BigInt{Value: []byte{}}
	}
	return &ccipocr3pb.BigInt{Value: i.Bytes()}
}

// Helper function to convert ConfidenceLevel to protobuf uint32
func confidenceLevelToPb(confidence primitives.ConfidenceLevel) uint32 {
	switch confidence {
	case primitives.Unconfirmed:
		return 0
	case primitives.Finalized:
		return 1
	default:
		return 0
	}
}

// Convert protobuf uint32 to ConfidenceLevel
func pbToConfidenceLevel(pb uint32) primitives.ConfidenceLevel {
	switch pb {
	case 0:
		return primitives.Unconfirmed
	case 1:
		return primitives.Finalized
	default:
		return primitives.Unconfirmed
	}
}

// Convert protobuf SourceChainConfig to ccipocr3.SourceChainConfig
func pbToSourceChainConfig(pb *ccipocr3pb.SourceChainConfig) ccipocr3.SourceChainConfig {
	if pb == nil {
		return ccipocr3.SourceChainConfig{}
	}
	return ccipocr3.SourceChainConfig{
		Router:                    pb.Router,
		IsEnabled:                 pb.IsEnabled,
		IsRMNVerificationDisabled: pb.IsRmnVerificationDisabled,
		MinSeqNr:                  pb.MinSeqNr,
		OnRamp:                    ccipocr3.UnknownAddress(pb.OnRamp),
	}
}

// Convert ccipocr3.SourceChainConfig to protobuf SourceChainConfig
func sourceChainConfigToPb(config ccipocr3.SourceChainConfig) *ccipocr3pb.SourceChainConfig {
	return &ccipocr3pb.SourceChainConfig{
		Router:                    config.Router,
		IsEnabled:                 config.IsEnabled,
		IsRmnVerificationDisabled: config.IsRMNVerificationDisabled,
		MinSeqNr:                  config.MinSeqNr,
		OnRamp:                    []byte(config.OnRamp),
	}
}

// Convert protobuf ChainConfigSnapshot to ccipocr3.ChainConfigSnapshot
func pbToChainConfigSnapshotDetailed(pb *ccipocr3pb.ChainConfigSnapshot) ccipocr3.ChainConfigSnapshot {
	if pb == nil {
		return ccipocr3.ChainConfigSnapshot{}
	}

	return ccipocr3.ChainConfigSnapshot{
		Offramp:   pbToOfframpConfig(pb.Offramp),
		RMNProxy:  pbToRMNProxyConfig(pb.RmnProxy),
		RMNRemote: pbToRMNRemoteConfig(pb.RmnRemote),
		FeeQuoter: pbToFeeQuoterConfig(pb.FeeQuoter),
		OnRamp:    pbToOnRampConfig(pb.OnRamp),
		Router:    pbToRouterConfig(pb.Router),
		CurseInfo: pbToCurseInfo(pb.CurseInfo),
	}
}

// Convert ccipocr3.ChainConfigSnapshot to protobuf ChainConfigSnapshot
func chainConfigSnapshotToPbDetailed(snapshot ccipocr3.ChainConfigSnapshot) *ccipocr3pb.ChainConfigSnapshot {
	return &ccipocr3pb.ChainConfigSnapshot{
		Offramp:   offrampConfigToPb(snapshot.Offramp),
		RmnProxy:  rmnProxyConfigToPb(snapshot.RMNProxy),
		RmnRemote: rmnRemoteConfigToPb(snapshot.RMNRemote),
		FeeQuoter: feeQuoterConfigToPb(snapshot.FeeQuoter),
		OnRamp:    onRampConfigToPb(snapshot.OnRamp),
		Router:    routerConfigToPb(snapshot.Router),
		CurseInfo: curseInfoToPb(snapshot.CurseInfo),
	}
}

// Convert protobuf FeeQuoterDestChainConfig to ccipocr3.FeeQuoterDestChainConfig
func pbToFeeQuoterDestChainConfigDetailed(pb *ccipocr3pb.FeeQuoterDestChainConfig) ccipocr3.FeeQuoterDestChainConfig {
	if pb == nil {
		return ccipocr3.FeeQuoterDestChainConfig{}
	}

	var chainFamilySelector [4]byte
	copy(chainFamilySelector[:], pb.ChainFamilySelector)

	return ccipocr3.FeeQuoterDestChainConfig{
		IsEnabled:                         pb.IsEnabled,
		MaxNumberOfTokensPerMsg:           uint16(pb.MaxNumberOfTokensPerMsg), // proto uint32 to Go uint16
		MaxDataBytes:                      pb.MaxDataBytes,
		MaxPerMsgGasLimit:                 pb.MaxPerMsgGasLimit,
		DestGasOverhead:                   pb.DestGasOverhead,
		DestGasPerPayloadByteBase:         pb.DestGasPerPayloadByte,
		DestGasPerPayloadByteHigh:         pb.DestGasPerPayloadByteHigh,
		DestGasPerPayloadByteThreshold:    pb.DestGasPerPayloadByteThreshold,
		DestDataAvailabilityOverheadGas:   pb.DestDataAvailabilityOverheadGas,
		DestGasPerDataAvailabilityByte:    uint16(pb.DestGasPerDataAvailabilityByte),    // proto uint32 to Go uint16
		DestDataAvailabilityMultiplierBps: uint16(pb.DestDataAvailabilityMultiplierBps), // proto uint32 to Go uint16
		DefaultTokenFeeUSDCents:           uint16(pb.DefaultTokenFeeUsdcCents),          // proto uint32 to Go uint16
		DefaultTokenDestGasOverhead:       pb.DefaultTokenDestGasOverhead,
		DefaultTxGasLimit:                 pb.DefaultTxGasLimit,
		GasMultiplierWeiPerEth:            pb.GasMultiplierWad,
		NetworkFeeUSDCents:                pb.NetworkFeeUsdcCents,
		GasPriceStalenessThreshold:        pb.GasPriceStalenessThreshold,
		EnforceOutOfOrder:                 pb.EnforceOutOfOrder,
		ChainFamilySelector:               chainFamilySelector,
	}
}

// Convert ccipocr3.FeeQuoterDestChainConfig to protobuf FeeQuoterDestChainConfig
func feeQuoterDestChainConfigToPb(config ccipocr3.FeeQuoterDestChainConfig) *ccipocr3pb.FeeQuoterDestChainConfig {
	return &ccipocr3pb.FeeQuoterDestChainConfig{
		IsEnabled:                         config.IsEnabled,
		MaxNumberOfTokensPerMsg:           uint32(config.MaxNumberOfTokensPerMsg), // Go uint16 to proto uint32 (safe: 0-65535)
		MaxDataBytes:                      config.MaxDataBytes,
		MaxPerMsgGasLimit:                 config.MaxPerMsgGasLimit,
		DestGasOverhead:                   config.DestGasOverhead,
		DestGasPerPayloadByte:             config.DestGasPerPayloadByteBase,
		DestGasPerPayloadByteHigh:         config.DestGasPerPayloadByteHigh,
		DestGasPerPayloadByteThreshold:    config.DestGasPerPayloadByteThreshold,
		DestDataAvailabilityOverheadGas:   config.DestDataAvailabilityOverheadGas,
		DestGasPerDataAvailabilityByte:    uint32(config.DestGasPerDataAvailabilityByte),    // Go uint16 to proto uint32 (safe: 0-65535)
		DestDataAvailabilityMultiplierBps: uint32(config.DestDataAvailabilityMultiplierBps), // Go uint16 to proto uint32 (safe: 0-65535)
		DefaultTokenFeeUsdcCents:          uint32(config.DefaultTokenFeeUSDCents),           // Go uint16 to proto uint32 (safe: 0-65535)
		DefaultTokenDestGasOverhead:       config.DefaultTokenDestGasOverhead,
		DefaultTxGasLimit:                 config.DefaultTxGasLimit,
		GasMultiplierWad:                  config.GasMultiplierWeiPerEth,
		NetworkFeeUsdcCents:               config.NetworkFeeUSDCents,
		GasPriceStalenessThreshold:        config.GasPriceStalenessThreshold,
		EnforceOutOfOrder:                 config.EnforceOutOfOrder,
		ChainFamilySelector:               config.ChainFamilySelector[:],
	}
}

// Convert ccipocr3.CurseInfo to protobuf CurseInfo
func curseInfoToPb(curseInfo ccipocr3.CurseInfo) *ccipocr3pb.CurseInfo {
	pb := &ccipocr3pb.CurseInfo{
		CursedSourceChains: make(map[uint64]bool),
		CursedDestination:  curseInfo.CursedDestination,
		GlobalCurse:        curseInfo.GlobalCurse,
	}

	for chainSel, cursed := range curseInfo.CursedSourceChains {
		pb.CursedSourceChains[uint64(chainSel)] = cursed
	}

	return pb
}

// Convert protobuf OfframpConfig to ccipocr3.OfframpConfig
func pbToOfframpConfig(pb *ccipocr3pb.OfframpConfig) ccipocr3.OfframpConfig {
	if pb == nil {
		return ccipocr3.OfframpConfig{}
	}
	return ccipocr3.OfframpConfig{
		CommitLatestOCRConfig: pbToOCRConfigResponse(pb.CommitLatestOcrConfig),
		ExecLatestOCRConfig:   pbToOCRConfigResponse(pb.ExecLatestOcrConfig),
		StaticConfig:          pbToOffRampStaticChainConfig(pb.StaticConfig),
		DynamicConfig:         pbToOffRampDynamicChainConfig(pb.DynamicConfig),
	}
}

func offrampConfigToPb(config ccipocr3.OfframpConfig) *ccipocr3pb.OfframpConfig {
	return &ccipocr3pb.OfframpConfig{
		CommitLatestOcrConfig: ocrConfigResponseToPb(config.CommitLatestOCRConfig),
		ExecLatestOcrConfig:   ocrConfigResponseToPb(config.ExecLatestOCRConfig),
		StaticConfig:          offRampStaticChainConfigToPb(config.StaticConfig),
		DynamicConfig:         offRampDynamicChainConfigToPb(config.DynamicConfig),
	}
}

func pbToOCRConfigResponse(pb *ccipocr3pb.OCRConfigResponse) ccipocr3.OCRConfigResponse {
	if pb == nil {
		return ccipocr3.OCRConfigResponse{}
	}
	return ccipocr3.OCRConfigResponse{
		OCRConfig: pbToOCRConfig(pb.OcrConfig),
	}
}

func ocrConfigResponseToPb(resp ccipocr3.OCRConfigResponse) *ccipocr3pb.OCRConfigResponse {
	return &ccipocr3pb.OCRConfigResponse{
		OcrConfig: ocrConfigToPb(resp.OCRConfig),
	}
}

func pbToOCRConfig(pb *ccipocr3pb.OCRConfig) ccipocr3.OCRConfig {
	if pb == nil {
		return ccipocr3.OCRConfig{}
	}
	return ccipocr3.OCRConfig{
		ConfigInfo:   pbToConfigInfo(pb.ConfigInfo),
		Signers:      pb.Signers,
		Transmitters: pb.Transmitters,
	}
}

func ocrConfigToPb(config ccipocr3.OCRConfig) *ccipocr3pb.OCRConfig {
	return &ccipocr3pb.OCRConfig{
		ConfigInfo:   configInfoToPb(config.ConfigInfo),
		Signers:      config.Signers,
		Transmitters: config.Transmitters,
	}
}

func pbToConfigInfo(pb *ccipocr3pb.ConfigInfo) ccipocr3.ConfigInfo {
	if pb == nil {
		return ccipocr3.ConfigInfo{}
	}
	var configDigest ccipocr3.Bytes32
	copy(configDigest[:], pb.ConfigDigest)
	return ccipocr3.ConfigInfo{
		ConfigDigest:                   configDigest,
		F:                              uint8(pb.F),
		N:                              uint8(pb.N),
		IsSignatureVerificationEnabled: pb.IsSignatureVerificationEnabled,
	}
}

func configInfoToPb(info ccipocr3.ConfigInfo) *ccipocr3pb.ConfigInfo {
	return &ccipocr3pb.ConfigInfo{
		ConfigDigest:                   info.ConfigDigest[:],
		F:                              uint32(info.F),
		N:                              uint32(info.N),
		IsSignatureVerificationEnabled: info.IsSignatureVerificationEnabled,
	}
}

func pbToOffRampStaticChainConfig(pb *ccipocr3pb.OffRampStaticChainConfig) ccipocr3.OffRampStaticChainConfig {
	if pb == nil {
		return ccipocr3.OffRampStaticChainConfig{}
	}
	return ccipocr3.OffRampStaticChainConfig{
		ChainSelector:        ccipocr3.ChainSelector(pb.ChainSelector),
		GasForCallExactCheck: uint16(pb.GasForCallExactCheck),
		RmnRemote:            pb.RmnRemote,
		TokenAdminRegistry:   pb.TokenAdminRegistry,
		NonceManager:         pb.NonceManager,
	}
}

func offRampStaticChainConfigToPb(config ccipocr3.OffRampStaticChainConfig) *ccipocr3pb.OffRampStaticChainConfig {
	return &ccipocr3pb.OffRampStaticChainConfig{
		ChainSelector:        uint64(config.ChainSelector),
		GasForCallExactCheck: uint32(config.GasForCallExactCheck),
		RmnRemote:            config.RmnRemote,
		TokenAdminRegistry:   config.TokenAdminRegistry,
		NonceManager:         config.NonceManager,
	}
}

func pbToOffRampDynamicChainConfig(pb *ccipocr3pb.OffRampDynamicChainConfig) ccipocr3.OffRampDynamicChainConfig {
	if pb == nil {
		return ccipocr3.OffRampDynamicChainConfig{}
	}
	return ccipocr3.OffRampDynamicChainConfig{
		FeeQuoter:                               pb.FeeQuoter,
		PermissionLessExecutionThresholdSeconds: pb.PermissionLessExecutionThresholdSeconds,
		IsRMNVerificationDisabled:               pb.IsRmnVerificationDisabled,
		MessageInterceptor:                      pb.MessageInterceptor,
	}
}

func offRampDynamicChainConfigToPb(config ccipocr3.OffRampDynamicChainConfig) *ccipocr3pb.OffRampDynamicChainConfig {
	return &ccipocr3pb.OffRampDynamicChainConfig{
		FeeQuoter:                               config.FeeQuoter,
		PermissionLessExecutionThresholdSeconds: config.PermissionLessExecutionThresholdSeconds,
		IsRmnVerificationDisabled:               config.IsRMNVerificationDisabled,
		MessageInterceptor:                      config.MessageInterceptor,
	}
}

// Convert protobuf RMNProxyConfig to ccipocr3.RMNProxyConfig
func pbToRMNProxyConfig(pb *ccipocr3pb.RMNProxyConfig) ccipocr3.RMNProxyConfig {
	if pb == nil {
		return ccipocr3.RMNProxyConfig{}
	}
	return ccipocr3.RMNProxyConfig{
		RemoteAddress: pb.RemoteAddress,
	}
}

func rmnProxyConfigToPb(config ccipocr3.RMNProxyConfig) *ccipocr3pb.RMNProxyConfig {
	return &ccipocr3pb.RMNProxyConfig{
		RemoteAddress: config.RemoteAddress,
	}
}

// Convert protobuf RMNRemoteConfigStruct to ccipocr3.RMNRemoteConfig
func pbToRMNRemoteConfig(pb *ccipocr3pb.RMNRemoteConfigStruct) ccipocr3.RMNRemoteConfig {
	if pb == nil {
		return ccipocr3.RMNRemoteConfig{}
	}
	return ccipocr3.RMNRemoteConfig{
		DigestHeader:    pbToRMNDigestHeader(pb.DigestHeader),
		VersionedConfig: pbToVersionedConfig(pb.VersionedConfig),
	}
}

func rmnRemoteConfigToPb(config ccipocr3.RMNRemoteConfig) *ccipocr3pb.RMNRemoteConfigStruct {
	return &ccipocr3pb.RMNRemoteConfigStruct{
		DigestHeader:    rmnDigestHeaderToPb(config.DigestHeader),
		VersionedConfig: versionedConfigToPb(config.VersionedConfig),
	}
}

func pbToRMNDigestHeader(pb *ccipocr3pb.RMNDigestHeader) ccipocr3.RMNDigestHeader {
	if pb == nil {
		return ccipocr3.RMNDigestHeader{}
	}
	var digestHeader ccipocr3.Bytes32
	copy(digestHeader[:], pb.DigestHeader)
	return ccipocr3.RMNDigestHeader{
		DigestHeader: digestHeader,
	}
}

func rmnDigestHeaderToPb(header ccipocr3.RMNDigestHeader) *ccipocr3pb.RMNDigestHeader {
	return &ccipocr3pb.RMNDigestHeader{
		DigestHeader: header.DigestHeader[:],
	}
}

func pbToVersionedConfig(pb *ccipocr3pb.VersionedConfig) ccipocr3.VersionedConfig {
	if pb == nil {
		return ccipocr3.VersionedConfig{}
	}
	return ccipocr3.VersionedConfig{
		Version: pb.Version,
		Config:  pbToRMNConfig(pb.Config),
	}
}

func versionedConfigToPb(config ccipocr3.VersionedConfig) *ccipocr3pb.VersionedConfig {
	return &ccipocr3pb.VersionedConfig{
		Version: config.Version,
		Config:  rmnConfigToPb(config.Config),
	}
}

func pbToRMNConfig(pb *ccipocr3pb.RMNConfig) ccipocr3.Config {
	if pb == nil {
		return ccipocr3.Config{}
	}
	var rmnDigest ccipocr3.Bytes32
	copy(rmnDigest[:], pb.RmnHomeContractConfigDigest)

	signers := make([]ccipocr3.Signer, len(pb.Signers))
	for i, pbSigner := range pb.Signers {
		signers[i] = pbToSigner(pbSigner)
	}

	return ccipocr3.Config{
		RMNHomeContractConfigDigest: rmnDigest,
		Signers:                     signers,
		FSign:                       pb.FSign,
	}
}

func rmnConfigToPb(config ccipocr3.Config) *ccipocr3pb.RMNConfig {
	signers := make([]*ccipocr3pb.SignerInfo, len(config.Signers))
	for i, signer := range config.Signers {
		signers[i] = signerToPb(signer)
	}

	return &ccipocr3pb.RMNConfig{
		RmnHomeContractConfigDigest: config.RMNHomeContractConfigDigest[:],
		Signers:                     signers,
		FSign:                       config.FSign,
	}
}

func pbToSigner(pb *ccipocr3pb.SignerInfo) ccipocr3.Signer {
	if pb == nil {
		return ccipocr3.Signer{}
	}
	return ccipocr3.Signer{
		OnchainPublicKey: pb.OnchainPublicKey,
		NodeIndex:        pb.NodeIndex,
	}
}

func signerToPb(signer ccipocr3.Signer) *ccipocr3pb.SignerInfo {
	return &ccipocr3pb.SignerInfo{
		OnchainPublicKey: signer.OnchainPublicKey,
		NodeIndex:        signer.NodeIndex,
	}
}

// Convert protobuf FeeQuoterConfigStruct to ccipocr3.FeeQuoterConfig
func pbToFeeQuoterConfig(pb *ccipocr3pb.FeeQuoterConfigStruct) ccipocr3.FeeQuoterConfig {
	if pb == nil {
		return ccipocr3.FeeQuoterConfig{}
	}
	return ccipocr3.FeeQuoterConfig{
		StaticConfig: pbToFeeQuoterStaticConfig(pb.StaticConfig),
	}
}

func feeQuoterConfigToPb(config ccipocr3.FeeQuoterConfig) *ccipocr3pb.FeeQuoterConfigStruct {
	return &ccipocr3pb.FeeQuoterConfigStruct{
		StaticConfig: feeQuoterStaticConfigToPb(config.StaticConfig),
	}
}

func pbToFeeQuoterStaticConfig(pb *ccipocr3pb.FeeQuoterStaticConfigStruct) ccipocr3.FeeQuoterStaticConfig {
	if pb == nil {
		return ccipocr3.FeeQuoterStaticConfig{}
	}
	return ccipocr3.FeeQuoterStaticConfig{
		MaxFeeJuelsPerMsg:  pbToBigInt(pb.MaxFeeJuelsPerMsg),
		LinkToken:          pb.LinkToken,
		StalenessThreshold: pb.StalenessThreshold,
	}
}

func feeQuoterStaticConfigToPb(config ccipocr3.FeeQuoterStaticConfig) *ccipocr3pb.FeeQuoterStaticConfigStruct {
	return &ccipocr3pb.FeeQuoterStaticConfigStruct{
		MaxFeeJuelsPerMsg:  intToPbBigInt(config.MaxFeeJuelsPerMsg.Int),
		LinkToken:          config.LinkToken,
		StalenessThreshold: config.StalenessThreshold,
	}
}

// Convert protobuf OnRampConfigStruct to ccipocr3.OnRampConfig
func pbToOnRampConfig(pb *ccipocr3pb.OnRampConfigStruct) ccipocr3.OnRampConfig {
	if pb == nil {
		return ccipocr3.OnRampConfig{}
	}
	return ccipocr3.OnRampConfig{
		DynamicConfig:   pbToGetOnRampDynamicConfigResponse(pb.DynamicConfig),
		DestChainConfig: pbToOnRampDestChainConfig(pb.DestChainConfig),
	}
}

func onRampConfigToPb(config ccipocr3.OnRampConfig) *ccipocr3pb.OnRampConfigStruct {
	return &ccipocr3pb.OnRampConfigStruct{
		DynamicConfig:   getOnRampDynamicConfigResponseToPb(config.DynamicConfig),
		DestChainConfig: onRampDestChainConfigToPb(config.DestChainConfig),
	}
}

func pbToGetOnRampDynamicConfigResponse(pb *ccipocr3pb.GetOnRampDynamicConfigResponse) ccipocr3.GetOnRampDynamicConfigResponse {
	if pb == nil {
		return ccipocr3.GetOnRampDynamicConfigResponse{}
	}
	return ccipocr3.GetOnRampDynamicConfigResponse{
		DynamicConfig: pbToOnRampDynamicConfig(pb.DynamicConfig),
	}
}

func getOnRampDynamicConfigResponseToPb(resp ccipocr3.GetOnRampDynamicConfigResponse) *ccipocr3pb.GetOnRampDynamicConfigResponse {
	return &ccipocr3pb.GetOnRampDynamicConfigResponse{
		DynamicConfig: onRampDynamicConfigToPb(resp.DynamicConfig),
	}
}

func pbToOnRampDynamicConfig(pb *ccipocr3pb.OnRampDynamicConfig) ccipocr3.OnRampDynamicConfig {
	if pb == nil {
		return ccipocr3.OnRampDynamicConfig{}
	}
	return ccipocr3.OnRampDynamicConfig{
		FeeQuoter:              pb.FeeQuoter,
		ReentrancyGuardEntered: pb.ReentrancyGuardEntered,
		MessageInterceptor:     pb.MessageInterceptor,
		FeeAggregator:          pb.FeeAggregator,
		AllowListAdmin:         pb.AllowListAdmin,
	}
}

func onRampDynamicConfigToPb(config ccipocr3.OnRampDynamicConfig) *ccipocr3pb.OnRampDynamicConfig {
	return &ccipocr3pb.OnRampDynamicConfig{
		FeeQuoter:              config.FeeQuoter,
		ReentrancyGuardEntered: config.ReentrancyGuardEntered,
		MessageInterceptor:     config.MessageInterceptor,
		FeeAggregator:          config.FeeAggregator,
		AllowListAdmin:         config.AllowListAdmin,
	}
}

func pbToOnRampDestChainConfig(pb *ccipocr3pb.OnRampDestChainConfig) ccipocr3.OnRampDestChainConfig {
	if pb == nil {
		return ccipocr3.OnRampDestChainConfig{}
	}
	return ccipocr3.OnRampDestChainConfig{
		SequenceNumber:   pb.SequenceNumber,
		AllowListEnabled: pb.AllowListEnabled,
		Router:           pb.Router,
	}
}

func onRampDestChainConfigToPb(config ccipocr3.OnRampDestChainConfig) *ccipocr3pb.OnRampDestChainConfig {
	return &ccipocr3pb.OnRampDestChainConfig{
		SequenceNumber:   config.SequenceNumber,
		AllowListEnabled: config.AllowListEnabled,
		Router:           config.Router,
	}
}

// Convert protobuf RouterConfigStruct to ccipocr3.RouterConfig
func pbToRouterConfig(pb *ccipocr3pb.RouterConfigStruct) ccipocr3.RouterConfig {
	if pb == nil {
		return ccipocr3.RouterConfig{}
	}
	return ccipocr3.RouterConfig{
		WrappedNativeAddress: pb.WrappedNativeAddress,
	}
}

func routerConfigToPb(config ccipocr3.RouterConfig) *ccipocr3pb.RouterConfigStruct {
	return &ccipocr3pb.RouterConfigStruct{
		WrappedNativeAddress: config.WrappedNativeAddress,
	}
}

func rmnReportToPb(report ccipocr3.RMNReport) *ccipocr3pb.RMNReport {
	laneUpdates := make([]*ccipocr3pb.RMNLaneUpdate, len(report.LaneUpdates))
	for i, update := range report.LaneUpdates {
		laneUpdates[i] = rmnLaneUpdateToPb(update)
	}

	return &ccipocr3pb.RMNReport{
		ReportVersionDigest:         report.ReportVersionDigest[:],
		DestChainId:                 intToPbBigInt(report.DestChainID.Int),
		DestChainSelector:           uint64(report.DestChainSelector),
		RmnRemoteContractAddress:    report.RmnRemoteContractAddress,
		OfframpAddress:              report.OfframpAddress,
		RmnHomeContractConfigDigest: report.RmnHomeContractConfigDigest[:],
		LaneUpdates:                 laneUpdates,
	}
}

func pbToRMNReport(pb *ccipocr3pb.RMNReport) ccipocr3.RMNReport {
	if pb == nil {
		return ccipocr3.RMNReport{}
	}

	var reportVersionDigest, rmnHomeDigest ccipocr3.Bytes32
	copy(reportVersionDigest[:], pb.ReportVersionDigest)
	copy(rmnHomeDigest[:], pb.RmnHomeContractConfigDigest)

	laneUpdates := make([]ccipocr3.RMNLaneUpdate, len(pb.LaneUpdates))
	for i, pbUpdate := range pb.LaneUpdates {
		laneUpdates[i] = pbToRMNLaneUpdate(pbUpdate)
	}

	return ccipocr3.RMNReport{
		ReportVersionDigest:         reportVersionDigest,
		DestChainID:                 pbToBigInt(pb.DestChainId),
		DestChainSelector:           ccipocr3.ChainSelector(pb.DestChainSelector),
		RmnRemoteContractAddress:    pb.RmnRemoteContractAddress,
		OfframpAddress:              pb.OfframpAddress,
		RmnHomeContractConfigDigest: rmnHomeDigest,
		LaneUpdates:                 laneUpdates,
	}
}

func rmnLaneUpdateToPb(update ccipocr3.RMNLaneUpdate) *ccipocr3pb.RMNLaneUpdate {
	return &ccipocr3pb.RMNLaneUpdate{
		SourceChainSelector: uint64(update.SourceChainSelector),
		OnRampAddress:       update.OnRampAddress,
		MinSeqNr:            uint64(update.MinSeqNr),
		MaxSeqNr:            uint64(update.MaxSeqNr),
		MerkleRoot:          update.MerkleRoot[:],
	}
}

func pbToRMNLaneUpdate(pb *ccipocr3pb.RMNLaneUpdate) ccipocr3.RMNLaneUpdate {
	if pb == nil {
		return ccipocr3.RMNLaneUpdate{}
	}

	var merkleRoot ccipocr3.Bytes32
	copy(merkleRoot[:], pb.MerkleRoot)

	return ccipocr3.RMNLaneUpdate{
		SourceChainSelector: ccipocr3.ChainSelector(pb.SourceChainSelector),
		OnRampAddress:       pb.OnRampAddress,
		MinSeqNr:            ccipocr3.SeqNum(pb.MinSeqNr),
		MaxSeqNr:            ccipocr3.SeqNum(pb.MaxSeqNr),
		MerkleRoot:          merkleRoot,
	}
}

// goMapToPbMap converts a Go map[string]any to protobuf Map using the values package
func goMapToPbMap(goMap map[string]any) (*pb.Map, error) {
	if goMap == nil {
		return nil, nil
	}

	valuesMap, err := values.NewMap(goMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert Go map to values.Map: %w", err)
	}

	return values.ProtoMap(valuesMap), nil
}

// pbMapToGoMap converts a protobuf Map to Go map[string]any using the values package
func pbMapToGoMap(pbMap *pb.Map) (map[string]any, error) {
	if pbMap == nil {
		return nil, nil
	}

	valuesMap, err := values.FromMapValueProto(pbMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert protobuf Map to values.Map: %w", err)
	}

	var goMap map[string]any
	err = valuesMap.UnwrapTo(&goMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap values.Map to Go map: %w", err)
	}

	return goMap, nil
}

func messageToPb(msg ccipocr3.Message) *ccipocr3pb.Message {
	pbMsg := &ccipocr3pb.Message{
		Header: &ccipocr3pb.RampMessageHeader{
			MessageId:           msg.Header.MessageID[:],
			SourceChainSelector: uint64(msg.Header.SourceChainSelector),
			DestChainSelector:   uint64(msg.Header.DestChainSelector),
			SequenceNumber:      uint64(msg.Header.SequenceNumber),
			Nonce:               msg.Header.Nonce,
			MessageHash:         msg.Header.MsgHash[:],
			OnRamp:              msg.Header.OnRamp,
			TxHash:              msg.Header.TxHash,
		},
		Sender:         msg.Sender,
		Data:           msg.Data,
		Receiver:       msg.Receiver,
		ExtraArgs:      msg.ExtraArgs,
		FeeToken:       msg.FeeToken,
		FeeTokenAmount: intToPbBigInt(msg.FeeTokenAmount.Int),
		FeeValueJuels:  intToPbBigInt(msg.FeeValueJuels.Int),
	}

	// Convert token amounts
	for _, ta := range msg.TokenAmounts {
		pbMsg.TokenAmounts = append(pbMsg.TokenAmounts, &ccipocr3pb.RampTokenAmount{
			SourcePoolAddress: ta.SourcePoolAddress,
			DestTokenAddress:  ta.DestTokenAddress,
			ExtraData:         ta.ExtraData,
			Amount:            intToPbBigInt(ta.Amount.Int),
		})
	}

	return pbMsg
}

// Convert protobuf Message to ccipocr3.Message
func pbToMessage(pb *ccipocr3pb.Message) ccipocr3.Message {
	// Convert header with proper Bytes32 conversion
	var messageID, msgHash ccipocr3.Bytes32
	copy(messageID[:], pb.Header.MessageId)
	copy(msgHash[:], pb.Header.MessageHash)

	return ccipocr3.Message{
		Header: ccipocr3.RampMessageHeader{
			MessageID:           messageID,
			SourceChainSelector: ccipocr3.ChainSelector(pb.Header.SourceChainSelector),
			DestChainSelector:   ccipocr3.ChainSelector(pb.Header.DestChainSelector),
			SequenceNumber:      ccipocr3.SeqNum(pb.Header.SequenceNumber),
			Nonce:               pb.Header.Nonce,
			MsgHash:             msgHash,
			OnRamp:              pb.Header.OnRamp,
			TxHash:              pb.Header.TxHash,
		},
		Sender:         pb.Sender,
		Data:           pb.Data,
		Receiver:       pb.Receiver,
		ExtraArgs:      pb.ExtraArgs,
		FeeToken:       pb.FeeToken,
		FeeTokenAmount: pbToBigInt(pb.FeeTokenAmount),
		FeeValueJuels:  pbToBigInt(pb.FeeValueJuels),
		TokenAmounts:   pbToTokenAmounts(pb.TokenAmounts),
	}
}

func pbToTokenAmounts(pbAmounts []*ccipocr3pb.RampTokenAmount) []ccipocr3.RampTokenAmount {
	var amounts []ccipocr3.RampTokenAmount
	for _, pb := range pbAmounts {
		amounts = append(amounts, ccipocr3.RampTokenAmount{
			SourcePoolAddress: pb.SourcePoolAddress,
			DestTokenAddress:  pb.DestTokenAddress,
			ExtraData:         pb.ExtraData,
			Amount:            pbToBigInt(pb.Amount),
		})
	}
	return amounts
}

func pbToCurseInfo(pb *ccipocr3pb.CurseInfo) ccipocr3.CurseInfo {
	result := ccipocr3.CurseInfo{
		CursedSourceChains: make(map[ccipocr3.ChainSelector]bool),
		CursedDestination:  pb.CursedDestination,
		GlobalCurse:        pb.GlobalCurse,
	}
	for chainSel, cursed := range pb.CursedSourceChains {
		result.CursedSourceChains[ccipocr3.ChainSelector(chainSel)] = cursed
	}
	return result
}
