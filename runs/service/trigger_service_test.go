package service

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger"
	"github.com/flyteorg/flyte/v2/runs/repository/mocks"
)

func TestGetTriggerRevisionDetailsErrors(t *testing.T) {
	for _, tc := range []struct {
		name  string
		cause error
		want  ErrorKind
	}{
		{"missing revision", fmt.Errorf("read revision: %w", sql.ErrNoRows), ErrNotFound},
		{"database failure", errors.New("connection lost"), ErrInternal},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			repo := mocks.NewTriggerRepo(t)
			repo.EXPECT().GetTriggerRevision(ctx, "p", "d", "job", "nightly", uint64(2)).Return(nil, tc.cause)
			req := &trigger.GetTriggerRevisionDetailsRequest{Id: &common.TriggerIdentifier{
				Name:     &common.TriggerName{Project: "p", Domain: "d", TaskName: "job", Name: "nightly"},
				Revision: 2,
			}}

			resp, err := NewTriggerService(repo).GetTriggerRevisionDetails(ctx, req)
			require.Nil(t, resp)
			require.ErrorIs(t, err, tc.want)
			require.ErrorIs(t, err, tc.cause)
		})
	}
}
