package limits

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

func ExampleRateLimiter_Allow() {
	ctx := context.Background()

	rl := NewRateLimiter(rate.Every(time.Second), 4)

	// Try 5
	var g errgroup.Group
	for range 5 {
		g.Go(func() error {
			if !rl.Allow(ctx) {
				return fmt.Errorf("rate limit exceeded")
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		fmt.Println("5:", err)
	} else {
		fmt.Println("5: success")
	}

	rl = NewRateLimiter(rate.Every(time.Second), 4) // reset

	// Try 4
	g = errgroup.Group{}
	for range 4 {
		g.Go(func() error {
			if !rl.Allow(ctx) {
				return fmt.Errorf("rate limit exceeded")
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		fmt.Println("4:", err)
	} else {
		fmt.Println("4: success")
	}

	// Output:
	// 5: rate limit exceeded
	// 4: success
}

func ExampleRateLimiter_AllowErr() {
	ctx := context.Background()

	rl := NewRateLimiter(rate.Every(time.Second), 4)

	// Try 5
	var g errgroup.Group
	for range 5 {
		g.Go(func() error {
			return rl.AllowErr(ctx)
		})
	}
	if err := g.Wait(); err != nil {
		fmt.Println("5:", err)
	} else {
		fmt.Println("5: success")
	}

	rl = NewRateLimiter(rate.Every(time.Second), 4) // reset

	// Try 4
	g = errgroup.Group{}
	for range 4 {
		g.Go(func() error {
			return rl.AllowErr(ctx)
		})
	}
	if err := g.Wait(); err != nil {
		fmt.Println("4:", err)
	} else {
		fmt.Println("4: success")
	}

	// Output:
	// 5: rate limited
	// 4: success
}

func ExampleRateLimiter_Wait() {
	ctx := context.Background()

	rl := NewRateLimiter(rate.Every(time.Second), 1)
	var wg sync.WaitGroup
	wg.Add(2)
	for range 2 {
		go func() {
			defer wg.Done()
			start := time.Now()
			err := rl.Wait(ctx)
			if err != nil {
				fmt.Println("rate limited:", err)
				return
			}
			fmt.Println("waited:", time.Since(start).Round(time.Second))
		}()
	}
	wg.Wait()

	// Output:
	// waited: 0s
	// waited: 1s
}

func ExampleMultiRateLimiter() {
	ctx := context.Background()

	global := NewRateLimiter(rate.Every(time.Second), 4)
	multiA := MultiRateLimiter{global, NewRateLimiter(rate.Every(time.Second), 4)}
	multiB := MultiRateLimiter{global, NewRateLimiter(rate.Every(time.Second), 4)}

	// Try burst limit of 4 from A
	var g errgroup.Group
	for range 4 {
		g.Go(func() error {
			return multiA.AllowErr(ctx)
		})
	}
	if err := g.Wait(); err != nil {
		fmt.Println("A:", err)
	} else {
		fmt.Println("A: success")
	}

	// Try burst limit of 4 from A & B a the same time

	g = errgroup.Group{}
	for range 4 {
		g.Go(func() error {
			return multiA.AllowErr(ctx)
		})
	}
	for range 4 {
		g.Go(func() error {
			return multiB.AllowErr(ctx)
		})
	}
	if err := g.Wait(); err != nil {
		fmt.Println("A&B:", err)
	} else {
		fmt.Println("A&B: success")
	}

	// Output:
	// A: success
	// A&B: rate limited
}
