package starkkey

import "math/big"

// curveOrder is the STARK curve subgroup order (N).
//
// starknet.go v0.17 dropped the exported curve.Curve type that previously
// exposed this value. We still need N for private-key sampling in
// GenerateKey and for OCR2 canonical signatures (s <= N/2).
var curveOrder = func() *big.Int {
	n, ok := new(big.Int).SetString("3618502788666131213697322783095070105526743751716087489154079457884512865583", 10)
	if !ok {
		panic("invalid stark curve order")
	}
	return n
}()
