package limits

import (
	"context"
	"fmt"
	"time"
)

func ExampleTimeLimiter_WithTimeout() {
	tl := NewTimeLimiter(time.Second)
	fn := func(ctx context.Context) {
		if ctx.Err() != nil {
			fmt.Println(ctx.Err())
			return
		}
		fmt.Println("done")
	}
	ctx, cancel, err := tl.WithTimeout(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cancel()
	fn(ctx)
	cancel()
	fn(ctx)
	// Output:
	// done
	// context canceled
}
