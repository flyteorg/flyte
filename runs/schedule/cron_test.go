package schedule

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
)

func TestValidateCronExpression(t *testing.T) {
	tests := []struct {
		name     string
		schedule *task.Schedule
		wantErr  bool
	}{
		{
			name: "structured cron with timezone",
			schedule: &task.Schedule{
				Expression: &task.Schedule_Cron{Cron: &task.Cron{
					Expression: "0 9 * * *",
					Timezone:   "America/Los_Angeles",
				}},
			},
		},
		{
			name: "invalid structured cron",
			schedule: &task.Schedule{
				Expression: &task.Schedule_Cron{Cron: &task.Cron{Expression: "invalid"}},
			},
			wantErr: true,
		},
		{
			name: "legacy cron",
			schedule: &task.Schedule{
				Expression: &task.Schedule_CronExpression{CronExpression: "0 9 * * *"},
			},
		},
		{
			name: "invalid legacy cron",
			schedule: &task.Schedule{
				Expression: &task.Schedule_CronExpression{CronExpression: "invalid"},
			},
			wantErr: true,
		},
		{
			name: "empty structured cron",
			schedule: &task.Schedule{
				Expression: &task.Schedule_Cron{Cron: &task.Cron{}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &task.TriggerAutomationSpec{
				Automation: &task.TriggerAutomationSpec_Schedule{
					Schedule: tt.schedule,
				},
			}
			err := ValidateCronExpression("trigger", spec)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
