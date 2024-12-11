package services

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsError(t *testing.T) {
	anError := errors.New("an error")
	anotherError := errors.New("another error")
	testCases := []struct {
		Name           string
		Report         map[string]error
		Target         error
		ExpectedResult bool
	}{
		{
			Name:           "nil map",
			Report:         nil,
			Target:         anError,
			ExpectedResult: false,
		},
		{
			Name:           "report contains service, but it's healthy",
			Report:         map[string]error{"service": nil},
			Target:         anError,
			ExpectedResult: false,
		},
		{
			Name:           "service is not healthy, but it's not caused by target error",
			Report:         map[string]error{"service": anotherError},
			Target:         anError,
			ExpectedResult: false,
		},
		{
			Name:           "service is not healthy and contains wrapper target",
			Report:         map[string]error{"service": fmt.Errorf("wrapped error: %w", anError)},
			Target:         anError,
			ExpectedResult: true,
		},
		{
			Name:           "service is not healthy due to multiple errors including target",
			Report:         map[string]error{"service": errors.Join(anError, anotherError)},
			Target:         anError,
			ExpectedResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actualResult := ContainsError(tc.Report, tc.Target)
			assert.Equal(t, tc.ExpectedResult, actualResult)
		})
	}
}
