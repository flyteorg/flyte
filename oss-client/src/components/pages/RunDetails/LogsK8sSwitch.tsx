import React, { useMemo } from 'react'
import { KubernetesIcon } from '@/components/icons/KubernetesIcon'
import { RunLogType } from './types'
import { RadioSwitch, type RadioSwitchItem } from '@/components/RadioSwitch'

export const RunK8sSwitch: React.FC<{
  onChange: (type: RunLogType) => void
  currentValue: RunLogType
}> = ({ onChange, currentValue }) => {
  const items = useMemo<RadioSwitchItem<RunLogType>[]>(
    () => [
      {
        label: 'Run Logs',
        value: RunLogType.RUN,
      },
      {
        label: 'Kubernetes Events',
        icon: <KubernetesIcon className="h-3.5" />,
        value: RunLogType.K8S,
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
