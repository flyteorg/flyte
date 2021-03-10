package launchplan

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytestdlib/errors"

	"github.com/stretchr/testify/assert"
)

func TestRemoteError(t *testing.T) {
	t.Run("alreadyExists", func(t *testing.T) {
		e := errors.Wrapf(RemoteErrorAlreadyExists, fmt.Errorf("blah"), "error")
		assert.Error(t, e)
		assert.True(t, IsAlreadyExists(e))
	})

	t.Run("notfound", func(t *testing.T) {
		e := errors.Wrapf(RemoteErrorNotFound, fmt.Errorf("blah"), "error")
		assert.Error(t, e)
		assert.True(t, IsNotFound(e))
	})

	t.Run("alreadyExists", func(t *testing.T) {
		e := errors.Wrapf(RemoteErrorUser, fmt.Errorf("blah"), "error")
		assert.Error(t, e)
		assert.True(t, IsUserError(e))
	})

	t.Run("system", func(t *testing.T) {
		e := errors.Wrapf(RemoteErrorSystem, fmt.Errorf("blah"), "error")
		assert.Error(t, e)
		assert.False(t, IsAlreadyExists(e))
		assert.False(t, IsNotFound(e))
		assert.False(t, IsUserError(e))
	})
}
