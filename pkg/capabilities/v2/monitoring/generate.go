// Run from the parent directory (pkg/capabilities/v2) so that proto file
// descriptors are registered as "monitoring/execution_context.proto" and
// "monitoring/events.proto", avoiding conflicts with the binary repo's
// "execution_context.proto" (capabilities/libs/monitoring).
//
//go:generate protoc -I .. --go_out=.. --go_opt=paths=source_relative monitoring/execution_context.proto monitoring/events.proto
package monitoring
