package limits

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
)

func ExampleResourceLimiter_Use() {
	ctx := context.Background()
	limiter := GlobalResourceLimiter[int](5)
	ch := make(chan struct{})
	go func() { // Do 2s of work with all 5 resources
		free, err := limiter.Use(ctx, 5)
		if err != nil {
			log.Fatalf("Failed to get resources: %v", err)
		}
		defer free(ctx)
		close(ch)
		time.Sleep(2 * time.Second)
	}()
	<-ch
	start := time.Now()
	// Blocks until goroutine frees resources
	free, err := limiter.Use(ctx, 1)
	defer free(ctx)
	elapsed := time.Since(start)
	if err != nil {
		log.Fatalf("Failed to get resources: %v", err)
	}
	fmt.Printf("Got resources after waiting: ~%s\n", elapsed.Round(time.Second))

	// Output:
	// Got resources after waiting: ~2s
}

func ExampleResourceLimiter_TryUse() {
	ctx := context.Background()
	limiter := GlobalResourceLimiter[int](5)
	free, err := limiter.Use(ctx, 5)
	if err != nil {
		log.Fatalf("Failed to get resources: %v", err)
	}
	defer free(ctx)

	// Returns immediately
	free2, err := limiter.TryUse(ctx, 1)
	if err != nil {
		if errors.Is(err, ErrorLimited[int]{}) {
			fmt.Printf("Try failed: %v\n", err)
			return
		}
		log.Fatalf("Failed to get resources: %v", err)
	}
	defer free2(ctx)

	// Output:
	// Try failed: global resource limited: cannot use 1, already using 5/5
}

func ExampleMultiResourceLimiter() {
	ctx := context.Background()
	ctx = contexts.WithOrg(ctx, "orgID")
	ctx = contexts.WithWorkflow(ctx, "workflowID")
	ctx = contexts.WithUser(ctx, "userID")
	global := GlobalResourceLimiter[int](100)
	freeGlobal, err := global.TryUse(ctx, 95)
	if err != nil {
		log.Fatal(err)
	}
	org := OrgResourceLimiter[int](50)
	freeOrg, err := org.TryUse(ctx, 45)
	if err != nil {
		log.Fatal(err)
	}
	user := UserResourceLimiter[int](20)
	freeUser, err := user.TryUse(ctx, 15)
	if err != nil {
		log.Fatal(err)
	}
	workflow := WorkflowResourceLimiter[int](10)
	freeWorkflow, err := workflow.TryUse(ctx, 5)
	if err != nil {
		log.Fatal(err)
	}
	multi := MultiResourceLimiter[int]{global, org, user, workflow}
	tryWork := func() error {
		free, err := multi.TryUse(ctx, 10)
		if err != nil {
			return err
		}
		defer free(ctx)
		return nil
	}

	fmt.Println(tryWork())
	freeGlobal(ctx)
	fmt.Println(tryWork())
	freeOrg(ctx)
	fmt.Println(tryWork())
	freeUser(ctx)
	fmt.Println(tryWork())
	freeWorkflow(ctx)
	fmt.Println(tryWork())
	// Output:
	// global resource limited: cannot use 10, already using 95/100
	// organization[orgID] resource limited: cannot use 10, already using 45/50
	// user[userID] resource limited: cannot use 10, already using 15/20
	// workflow[workflowID] resource limited: cannot use 10, already using 5/10
	// <nil>
}

//TODO all modes, overridden each way

//TODO validate metrics
