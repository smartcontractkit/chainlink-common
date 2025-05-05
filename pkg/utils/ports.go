package utils

import (
	"net"
	"strconv"
	"testing"

	"github.com/smartcontractkit/freeport"
)

func IsPortOpen(t *testing.T, port string) bool {
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		t.Log("error in checking port: ", err.Error())
		return false
	}
	defer l.Close()
	return true
}

// Deprecated: use https://github.com/smartcontractkit/freeport GetOne
func MustRandomPort(t *testing.T) string {
	return strconv.Itoa(freeport.GetOne(t))
}
