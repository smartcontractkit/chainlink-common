package testutils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	basicactionmock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/basicaction/basic_actionmock"
	basictriggermock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/basictrigger/basic_triggermock"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/testutils"
)

func TestRegisterCapability(t *testing.T) {
	r := &testutils.Registry{}
	c := &basictriggermock.BasicCapability{}

	err := r.RegisterCapability(c)
	assert.NoError(t, err)

	err = r.RegisterCapability(c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestGetCapability(t *testing.T) {
	r := &testutils.Registry{}
	c1 := &basictriggermock.BasicCapability{}
	c2 := &basicactionmock.BasicActionCapability{}

	err := r.RegisterCapability(c1)
	require.NoError(t, err)

	// register a second capability to make sure that the same capability isn't always returned
	err = r.RegisterCapability(c2)
	require.NoError(t, err)

	got, err := r.GetCapability(c1.ID())
	require.NoError(t, err)
	assert.Equal(t, c1, got)

	got, err = r.GetCapability(c2.ID())
	require.NoError(t, err)
	assert.Equal(t, c2, got)

	notReal := "not" + c1.ID()
	_, err = r.GetCapability(notReal)
	require.Equal(t, err, testutils.NoCapability(notReal))
}
