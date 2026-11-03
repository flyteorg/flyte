package schedule

import (
	"fmt"

	"github.com/robfig/cron/v3"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
)

// ParseCron parses either form of cron schedule, including its timezone.
func ParseCron(schedule *task.Schedule) (cron.Schedule, error) {
	if schedule == nil {
		return nil, nil
	}

	var expression string
	switch value := schedule.GetExpression().(type) {
	case *task.Schedule_CronExpression:
		expression = value.CronExpression
	case *task.Schedule_Cron:
		expression = value.Cron.GetExpression()
		if timezone := value.Cron.GetTimezone(); timezone != "" {
			expression = fmt.Sprintf("CRON_TZ=%s %s", timezone, expression)
		}
	default:
		return nil, nil
	}

	parsed, err := cron.ParseStandard(expression)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression %q: %w", expression, err)
	}
	return parsed, nil
}
