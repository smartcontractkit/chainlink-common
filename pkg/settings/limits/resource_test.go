package limits

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
)

func ExampleResourcePoolLimiter_Wait() {
	ctx := context.Background()
	limiter := GlobalResourcePoolLimiter[int](5)
	ch := make(chan struct{})
	go func() { // Do 2s of work with all 5 resources
		free, err := limiter.Wait(ctx, 5)
		if err != nil {
			log.Fatalf("Failed to get resources: %v", err)
		}
		defer free()
		close(ch)
		time.Sleep(2 * time.Second)
	}()
	<-ch
	start := time.Now()
	// Blocks until goroutine frees resources
	free, err := limiter.Wait(ctx, 1)
	defer free()
	elapsed := time.Since(start)
	if err != nil {
		log.Fatalf("Failed to get resources: %v", err)
	}
	fmt.Printf("Got resources after waiting: ~%s\n", elapsed.Round(time.Second))

	// Output:
	// Got resources after waiting: ~2s
}

func ExampleResourceLimiter_Use() {
	ctx := context.Background()
	limiter := GlobalResourcePoolLimiter[int](5)
	free, err := limiter.Wait(ctx, 5)
	if err != nil {
		log.Fatalf("Failed to get resources: %v", err)
	}
	defer free()

	// Returns immediately
	err = limiter.Use(ctx, 1)
	if err != nil {
		if errors.Is(err, ErrorResourceLimited[int]{}) {
			fmt.Printf("Try failed: %v\n", err)
			return
		}
		log.Fatalf("Failed to get resources: %v", err)
	}
	defer limiter.Free(ctx, 1)

	// Output:
	// Try failed: resource limited: cannot use 1, already using 5/5
}

func ExampleMultiResourcePoolLimiter() {
	ctx := contexts.WithCRE(context.Background(), contexts.CRE{Org: "orgID", Owner: "ownerID", Workflow: "workflowID"})
	global := GlobalResourcePoolLimiter[int](100)
	freeGlobal, err := global.Wait(ctx, 95)
	if err != nil {
		log.Fatal(err)
	}
	org := OrgResourcePoolLimiter[int](50)
	freeOrg, err := org.Wait(ctx, 45)
	if err != nil {
		log.Fatal(err)
	}
	user := OwnerResourcePoolLimiter[int](20)
	freeUser, err := user.Wait(ctx, 15)
	if err != nil {
		log.Fatal(err)
	}
	workflow := WorkflowResourcePoolLimiter[int](10)
	freeWorkflow, err := workflow.Wait(ctx, 5)
	if err != nil {
		log.Fatal(err)
	}
	multi := MultiResourcePoolLimiter[int]{global, org, user, workflow}
	tryWork := func() error {
		free, err := multi.Use(ctx, 10)
		if err != nil {
			return err
		}
		defer free()
		return nil
	}

	fmt.Println(tryWork())
	freeGlobal()
	fmt.Println(tryWork())
	freeOrg()
	fmt.Println(tryWork())
	freeUser()
	fmt.Println(tryWork())
	freeWorkflow()
	fmt.Println(tryWork())
	// Output:
	// resource limited: cannot use 10, already using 95/100
	// resource limited for org[orgID]: cannot use 10, already using 45/50
	// resource limited for owner[ownerID]: cannot use 10, already using 15/20
	// resource limited for workflow[workflowID]: cannot use 10, already using 5/10
	// <nil>
}
