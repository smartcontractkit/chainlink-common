package timeutil_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/timeutil"
)

func ExampleSleep() {
	ctx := context.Background()
	if !timeutil.Sleep(ctx.Done(), time.Second) {
		log.Fatal("context done")
	}
	fmt.Printf("Slept for %s\n", time.Second)

	// Output:
	// Slept for 1s
}
