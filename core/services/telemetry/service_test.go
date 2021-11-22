package telemetry

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-relay/core/services/telemetry/generated"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/wsrpc"
	"github.com/stretchr/testify/require"
)

func TestTelemetry(t *testing.T) {
	_, serverPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	serverPublicKey := serverPrivateKey.Public().(ed25519.PublicKey)
	_, clientPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	clientPublicKey := clientPrivateKey.Public().(ed25519.PublicKey)

	server := wsrpc.NewServer(wsrpc.Creds(serverPrivateKey, []ed25519.PublicKey{clientPublicKey}))
	defer server.Stop()

	backend := &telemetryServer{[][]byte{}, sync.Mutex{}}
	generated.RegisterTelemetryServer(server, backend)

	listener, err := net.Listen("tcp", "127.0.0.1:1337")
	require.NoError(t, err)
	defer listener.Close()
	go server.Serve(listener)

	log := logger.Default.Named("telemetry")
	serverURL, err := url.Parse("ws://127.0.0.1:1337")
	require.NoError(t, err)
	service := NewService(serverURL, clientPrivateKey, serverPublicKey, log)
	client, err := service.Start()
	require.NoError(t, err)
	defer service.Stop()

	<-time.After(5 * time.Second)

	// Simulate 10 instances of OCR all sharing the same client publishing telemetry.
	var wg sync.WaitGroup
	wg.Add(10)
	var i uint8
	for i = 0; i < 10; i++ {
		go func(i uint8) {
			defer wg.Done()
			var j uint8
			for j = 0; j < 10; j++ {
				client.Send(&generated.TelemetryRequest{
					Telemetry: []byte{i*10 + j},
					Address:   strconv.Itoa(int(i)),
				})
				fmt.Printf("node %d sent msg %d\n", i, j)
			}
		}(i)
	}
	wg.Wait()

	<-time.After(2 * time.Second)

	sort.Slice(backend.buffer, func(i, j int) bool {
		return bytes.Compare(backend.buffer[i], backend.buffer[j]) < 0
	})

	// TODO (dru) add assertions
}

// *telemetryServer is an instance of generated.TelemetryServer used in tests.
type telemetryServer struct {
	buffer [][]byte
	mu     sync.Mutex
}

func (t *telemetryServer) Telemetry(_ context.Context, req *generated.TelemetryRequest) (*generated.TelemetryResponse, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buffer = append(t.buffer, req.Telemetry)
	return &generated.TelemetryResponse{}, nil
}
