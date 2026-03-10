import { useMemo } from 'react'
import { RadioSwitch, type RadioSwitchItem } from '@/components/RadioSwitch'

export enum AppLogType {
  APP = 'APP',
  SCALING = 'SCALING',
}

export const LogSwitch: React.FC<{
  onChange: (type: AppLogType) => void
  currentValue: AppLogType
}> = ({ onChange, currentValue }) => {
  const items = useMemo<RadioSwitchItem<AppLogType>[]>(
    () => [
      {
        label: 'App Logs',
        value: AppLogType.APP,
      },
      {
        label: 'Scaling Events',
        value: AppLogType.SCALING,
      },
    ],
    [],
  )

  return (
    <RadioSwitch
      items={items}
      currentValue={currentValue}
      onChange={onChange}
    />
  )
}
