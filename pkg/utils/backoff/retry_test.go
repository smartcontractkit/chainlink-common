package backoff

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"
)

type testTimer struct {
	timer *time.Timer
}

func (t *testTimer) Start(duration time.Duration) {
	t.timer = time.NewTimer(0)
}

func (t *testTimer) Stop() {
	if t.timer != nil {
		t.timer.Stop()
	}
}

func (t *testTimer) C() <-chan time.Time {
	return t.timer.C
}

func TestRetry(t *testing.T) {
	const successOn = 3
	var i = 0

	// This function is successful on "successOn" calls.
	f := func() (bool, error) {
		i++
		log.Printf("function is called %d. time\n", i)

		if i == successOn {
			log.Println("OK")
			return true, nil
		}

		log.Println("error")
		return false, errors.New("error")
	}

	_, err := Retry(context.Background(), f, WithBackOff(NewExponentialBackOff()), withTimer(&testTimer{}))
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if i != successOn {
		t.Errorf("invalid number of retries: %d", i)
	}
}

func TestRetryWithData(t *testing.T) {
	const successOn = 3
	var i = 0

	// This function is successful on "successOn" calls.
	f := func() (int, error) {
		i++
		log.Printf("function is called %d. time\n", i)

		if i == successOn {
			log.Println("OK")
			return 42, nil
		}

		log.Println("error")
		return 1, errors.New("error")
	}

	res, err := Retry(context.Background(), f, WithBackOff(NewExponentialBackOff()), withTimer(&testTimer{}))
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if i != successOn {
		t.Errorf("invalid number of retries: %d", i)
	}
	if res != 42 {
		t.Errorf("invalid data in response: %d, expected 42", res)
	}
}

func TestRetryContext(t *testing.T) {
	var cancelOn = 3
	var i = 0

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(context.Canceled)

	expectedErr := errors.New("custom error")

	// This function cancels context on "cancelOn" calls.
	f := func() (bool, error) {
		i++
		log.Printf("function is called %d. time\n", i)

		// cancelling the context in the operation function is not a typical
		// use-case, however it allows to get predictable test results.
		if i == cancelOn {
			cancel(expectedErr)
		}

		log.Println("error")
		return false, fmt.Errorf("error (%d)", i)
	}

	_, err := Retry(ctx, f, WithBackOff(&ZeroBackOff{}), withTimer(&testTimer{}))
	if err == nil {
		t.Errorf("error is unexpectedly nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if i != cancelOn {
		t.Errorf("invalid number of retries: %d", i)
	}
}
