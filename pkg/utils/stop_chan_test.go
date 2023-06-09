package utils

import (
	"context"
	"testing"
	"time"
)

func TestStopChan_NewCtx(t *testing.T) {
	sc := make(StopChan)

	ctx, cancel := sc.NewCtx()
	defer cancel()

	go func() {
		close(sc)
	}()

	select {
	case <-time.After(1 * time.Second):
		t.Fatal("context should be cancelled when StopChan is closed")
	case <-ctx.Done():
	}
}

func TestStopChan_CtxCancel(t *testing.T) {
	stopChan := make(StopChan)
	originalCtx, originalCancel := context.WithTimeout(context.Background(), time.Second)
	ctx, cancel := stopChan.CtxCancel(originalCtx, originalCancel)
	defer cancel()

	if ctx != originalCtx {
		t.Fatal("expected ctx to be the same as originalCtx but it wasn't")
	}

	go func() {
		close(stopChan)
	}()

	select {
	case <-ctx.Done():
	case <-time.After(1 * time.Second):
		t.Fatal("expected ctx to be cancelled but it wasn't")
	}
}
