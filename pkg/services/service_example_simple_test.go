package services_test

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"

	. "github.com/smartcontractkit/chainlink-common/pkg/internal/example"
)

type simple struct {
	services.Service
}

func NewSimple(lggr logger.Logger) *simple {
	return &simple{Service: services.Config{Name: "Example"}.NewService(lggr)}
}

func ExampleNewService() {
	lggr, err := Logger()
	if err != nil {
		fmt.Println("Failed to create logger:", err)
		return
	}
	s := NewSimple(lggr)
	if err = s.Start(context.Background()); err != nil {
		fmt.Println("Failed to start service:", err)
		return
	}
	if err = s.Close(); err != nil {
		fmt.Println("Failed to close service:", err)
		return
	}

	// Output:
	// INFO	Example	Starting
	// INFO	Example	Started
	// INFO	Example	Closing
	// INFO	Example	Closed
}
