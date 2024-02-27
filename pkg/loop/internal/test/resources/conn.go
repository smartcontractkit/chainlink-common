package resources_test

import "google.golang.org/grpc"

type MockConn struct {
	grpc.ClientConnInterface
}
