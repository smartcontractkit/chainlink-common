//go:generate go run ./gen

// Package pb holds the plain-gRPC CapabilitiesRegistry shim protocol.
//
// Capability type and capability configuration are imported from
// capabilities/pb rather than redeclared, so generation needs pkg/ on the
// include path; ./gen sets that up.
package pb
