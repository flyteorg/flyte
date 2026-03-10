import { CronScheduleInput } from './CronScheduleInput'
import { FixedRateInput } from './FixedRateInput'

export type AutomationType = {
  id: Automation
  title: string
}

export type Automation = 'cron-schedule' | 'fixed-rate' | 'webhook' | 'artifact'

export const automationTypes: AutomationType[] = [
  {
    id: 'cron-schedule',
    title: 'Cron schedule',
  },
  {
    id: 'fixed-rate',
    title: 'Fixed rate',
  },
  // {
  //   id: 'webhook',
  //   title: 'Webhook',
  // },
  // {
  //   id: 'artifact',
  //   title: 'Artifact',
  // },
]

export const AutomationTypeFields = ({
  automationType,
}: {
  automationType: Automation
}) => {
  switch (automationType) {
    case 'cron-schedule': {
      return <CronScheduleInput />
    }
    case 'fixed-rate': {
      return <FixedRateInput />
    }
    default: {
      return <div>Not implemented</div>
    }
  }
}
