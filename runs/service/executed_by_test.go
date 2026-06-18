package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

func mustMarshalIdentity(t *testing.T, id *common.EnrichedIdentity) []byte {
	t.Helper()
	b, err := proto.Marshal(id)
	require.NoError(t, err)
	return b
}

func fullIdentity(sub, first, last, email string) *common.EnrichedIdentity {
	return &common.EnrichedIdentity{Principal: &common.EnrichedIdentity_User{User: &common.User{
		Id:   &common.UserIdentifier{Subject: sub},
		Spec: &common.UserSpec{FirstName: first, LastName: last, Email: email},
	}}}
}

func TestActionMetadataFromModel_ExecutedBy(t *testing.T) {
	t.Run("full identity from executed_by", func(t *testing.T) {
		m := &models.Action{ExecutedBy: mustMarshalIdentity(t, fullIdentity("00u1", "Kevin", "Su", "kevin@union.ai"))}
		eb := actionMetadataFromModel(m).GetExecutedBy().GetUser()
		assert.Equal(t, "00u1", eb.GetId().GetSubject())
		assert.Equal(t, "Kevin", eb.GetSpec().GetFirstName())
		assert.Equal(t, "Su", eb.GetSpec().GetLastName())
		assert.Equal(t, "kevin@union.ai", eb.GetSpec().GetEmail())
	})

	t.Run("falls back to subject-only from created_by", func(t *testing.T) {
		m := &models.Action{}
		m.CreatedBy.Valid, m.CreatedBy.String = true, "00u2"
		eb := actionMetadataFromModel(m).GetExecutedBy().GetUser()
		assert.Equal(t, "00u2", eb.GetId().GetSubject())
		assert.Nil(t, eb.GetSpec())
	})

	t.Run("corrupt executed_by falls back to created_by", func(t *testing.T) {
		m := &models.Action{ExecutedBy: []byte("not a valid proto\xff\xfe")}
		m.CreatedBy.Valid, m.CreatedBy.String = true, "00u3"
		assert.Equal(t, "00u3", actionMetadataFromModel(m).GetExecutedBy().GetUser().GetId().GetSubject())
	})

	t.Run("no identity yields nil executed_by", func(t *testing.T) {
		assert.Nil(t, actionMetadataFromModel(&models.Action{}).GetExecutedBy())
	})
}
