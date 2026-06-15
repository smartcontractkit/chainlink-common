package starkkey

import "math/big"

// curveOrder is the STARK curve subgroup order (N).
//
// starknet.go v0.17 dropped the exported curve.Curve type when the curve pkg
// migrated to gnark-crypto (https://github.com/NethermindEth/starknet.go/issues/710).
// We still need N for OCR2 canonical signatures (s <= N/2).
//
// https://docs.starknet.io/learn/protocol/cryptography
var curveOrder = func() *big.Int {
	n, ok := new(big.Int).SetString("3618502788666131213697322783095070105526743751716087489154079457884512865583", 10)
	if !ok {
		panic("invalid stark curve order")
	}
	return n
}()
