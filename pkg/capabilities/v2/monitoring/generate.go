//go:generate protoc -I .. --go_out=.. --go_opt=paths=source_relative monitoring/execution_context.proto monitoring/events.proto
package monitoring
