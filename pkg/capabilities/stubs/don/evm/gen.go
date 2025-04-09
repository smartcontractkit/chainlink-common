package evm

//go:generate protoc -I. -I../../../protoc/pkg/pb -I../crosschain --go_out=. --go_opt=paths=source_relative "--cre_out=mode=don,id=evm@1.0.0:." evm.proto
