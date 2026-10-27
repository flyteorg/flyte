package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

func TestMatchesIdentity(t *testing.T) {
	identity, err := NewIdentityContext("", "sub", "app", time.Now(), sets.NewString(),
		&service.UserInfoResponse{Email: "dev@example.com"}, nil)
	assert.NoError(t, err)

	assert.True(t, MatchesIdentity(identity, []string{"sub"}))
	assert.True(t, MatchesIdentity(identity, []string{"app"}))
	assert.True(t, MatchesIdentity(identity, []string{"dev@example.com"}))
	assert.False(t, MatchesIdentity(identity, []string{"other"}))
	assert.False(t, MatchesIdentity(identity, nil))
	bare, err := NewIdentityContext("", "sub", "", time.Now(), sets.NewString(), nil, nil)
	assert.NoError(t, err)
	assert.False(t, MatchesIdentity(bare, []string{""}))
}

func TestHasGroup(t *testing.T) {
	identity, err := NewIdentityContext("", "sub", "", time.Now(), sets.NewString(), nil,
		map[string]interface{}{"groups": []interface{}{"/dcp_team"}})
	assert.NoError(t, err)

	assert.True(t, HasGroup(identity, "groups", []string{"dcp_team"}))
	assert.True(t, HasGroup(identity, "groups", []string{"/dcp_team"}))
	assert.False(t, HasGroup(identity, "groups", []string{"other"}))
	assert.False(t, HasGroup(identity, "", []string{"dcp_team"}))
	assert.False(t, HasGroup(identity, "realm_access", []string{"dcp_team"}))
}

func TestClaimStrings(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, ClaimStrings([]interface{}{"a", "b"}))
	assert.Equal(t, []string{"a", "b"}, ClaimStrings([]string{"a", "b"}))
	assert.Equal(t, []string{"a"}, ClaimStrings("a"))
	assert.Empty(t, ClaimStrings([]interface{}{1}))
	assert.Nil(t, ClaimStrings(nil))
	assert.Nil(t, ClaimStrings(42))
}
