package sdk_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

func TestPromiseFromResult(t *testing.T) {
	p := sdk.PromiseFromResult("hello", nil)

	val, err := p.Await()
	assert.NoError(t, err)
	assert.Equal(t, "hello", val)
}

func TestPromiseFromResultError(t *testing.T) {
	expectedErr := errors.New("failure")
	p := sdk.PromiseFromResult[string]("", expectedErr)

	val, err := p.Await()
	assert.ErrorIs(t, err, expectedErr)
	assert.Equal(t, "", val)
}

func TestBasicPromiseResolvesOnlyOnce(t *testing.T) {
	counter := 0
	p := sdk.NewBasicPromise(func() (int, error) {
		counter++
		return 42, nil
	})

	val1, err1 := p.Await()
	val2, err2 := p.Await()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, 42, val1)
	assert.Equal(t, 42, val2)
	assert.Equal(t, 1, counter, "await should be called only once")
}

func TestBasicPromiseError(t *testing.T) {
	p := sdk.NewBasicPromise(func() (int, error) {
		return 0, errors.New("something went wrong")
	})

	_, err := p.Await()
	assert.Error(t, err)
}

func TestThenChainsCorrectly(t *testing.T) {
	p := sdk.PromiseFromResult(3, nil)
	chained := sdk.Then(p, func(i int) (string, error) {
		return string(rune('A' + i)), nil
	})

	result, err := chained.Await()
	assert.NoError(t, err)
	assert.Equal(t, "D", result)
}

func TestThenPropagatesError(t *testing.T) {
	expectedErr := errors.New("boom")
	p := sdk.PromiseFromResult[int](0, expectedErr)

	chained := sdk.Then(p, func(i int) (string, error) {
		return "should not happen", nil
	})

	_, err := chained.Await()
	assert.ErrorIs(t, err, expectedErr)
}

func TestThenHandlesFnError(t *testing.T) {
	p := sdk.PromiseFromResult(123, nil)
	fnErr := errors.New("failed")
	chained := sdk.Then(p, func(i int) (string, error) {
		return "", fnErr
	})

	_, err := chained.Await()
	assert.ErrorIs(t, err, fnErr)
}
