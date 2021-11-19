package telemetry

// go:generate protoc --go_out=. --go_opt=paths=source_relative --go-wsrpc_out=. --go-wsrpc_opt=paths=source_relative telemetry.proto
// go:generate mockery --name=WSRPCConnection --inpackage --case=underscore
