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
	t.Run("full identity from created_by", func(t *testing.T) {
		m := &models.Action{CreatedBy: mustMarshalIdentity(t, fullIdentity("00u1", "Kevin", "Su", "kevin@union.ai"))}
		eb := actionMetadataFromModel(m).GetExecutedBy().GetUser()
		assert.Equal(t, "00u1", eb.GetId().GetSubject())
		assert.Equal(t, "Kevin", eb.GetSpec().GetFirstName())
		assert.Equal(t, "Su", eb.GetSpec().GetLastName())
		assert.Equal(t, "kevin@union.ai", eb.GetSpec().GetEmail())
	})

	t.Run("subject-only identity from created_by", func(t *testing.T) {
		m := &models.Action{CreatedBy: mustMarshalIdentity(t, subjectOnlyIdentity("00u2"))}
		eb := actionMetadataFromModel(m).GetExecutedBy().GetUser()
		assert.Equal(t, "00u2", eb.GetId().GetSubject())
		assert.Nil(t, eb.GetSpec())
	})

	t.Run("corrupt created_by yields nil", func(t *testing.T) {
		m := &models.Action{CreatedBy: []byte("not a valid proto\xff\xfe")}
		assert.Nil(t, actionMetadataFromModel(m).GetExecutedBy())
	})

	t.Run("no identity yields nil executed_by", func(t *testing.T) {
		assert.Nil(t, actionMetadataFromModel(&models.Action{}).GetExecutedBy())
	})
}
