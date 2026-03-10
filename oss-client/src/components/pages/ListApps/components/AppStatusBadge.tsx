import { CheckIcon } from '@/components/icons/CheckIcon'
import { CircleDotIcon } from '@/components/icons/CircleDotIcon'
import { CircleIcon } from '@/components/icons/CircleIcon'
import { ExclamationIcon } from '@/components/icons/ExclamationIcon'
import { MultiEllipsisIcon } from '@/components/icons/MultiEllipsisIcon'
import { QuestionMarkIcon } from '@/components/icons/QuestionMarkIcon'
import { Status_DeploymentStatus } from '@/gen/flyteidl2/app/app_definition_pb'
import { useAccent, usePalette } from '@/hooks/usePalette'
import { AccentColor } from '@/types/colors'
import { XMarkIcon } from '@heroicons/react/16/solid'
import { PauseIcon } from '@heroicons/react/20/solid'
import clsx from 'clsx'

const getAppStatusDisplayText = (status: Status_DeploymentStatus): string[] => {
  switch (status) {
    case Status_DeploymentStatus.ACTIVE:
      return ['Active']
    case Status_DeploymentStatus.ASSIGNED:
      return ['Assigned']
    case Status_DeploymentStatus.DEPLOYING:
      return ['Deploying']
    case Status_DeploymentStatus.FAILED:
      return ['Failed']
    case Status_DeploymentStatus.PENDING:
      return ['Pending']
    case Status_DeploymentStatus.SCALING_DOWN:
      return ['Active ', 'Scaling down']
    case Status_DeploymentStatus.SCALING_UP:
      return ['Active', 'Scaling up']
    case Status_DeploymentStatus.STOPPED:
      return ['Stopped']
    case Status_DeploymentStatus.UNASSIGNED:
      return ['Inactive']
    default:
    case Status_DeploymentStatus.UNSPECIFIED:
      return ['Unspecified']
  }
}

const getColorForStatus = (status: Status_DeploymentStatus): AccentColor => {
  switch (status) {
    case Status_DeploymentStatus.ACTIVE:
    case Status_DeploymentStatus.SCALING_DOWN:
    case Status_DeploymentStatus.SCALING_UP:
      return 'green'
    case Status_DeploymentStatus.ASSIGNED:
    case Status_DeploymentStatus.DEPLOYING:
      return 'purple'
    case Status_DeploymentStatus.PENDING:
      return 'yellow'
    case Status_DeploymentStatus.STOPPED:
      return 'orange'
    case Status_DeploymentStatus.FAILED:
      return 'red'
    case Status_DeploymentStatus.UNASSIGNED:
    case Status_DeploymentStatus.UNSPECIFIED:
    case Status_DeploymentStatus.STARTED:
    default:
      return 'gray'
  }
}

const getIconForStatus = (
  status: Status_DeploymentStatus,
): React.ReactElement => {
  switch (status) {
    case Status_DeploymentStatus.SCALING_DOWN:
    case Status_DeploymentStatus.SCALING_UP:
    case Status_DeploymentStatus.ACTIVE: {
      return <CheckIcon />
    }
    case Status_DeploymentStatus.DEPLOYING: {
      return <MultiEllipsisIcon width={11} />
    }
    case Status_DeploymentStatus.ASSIGNED: {
      return <CircleDotIcon width={11} />
    }
    case Status_DeploymentStatus.FAILED: {
      return <ExclamationIcon width={11} />
    }
    case Status_DeploymentStatus.STOPPED: {
      return <XMarkIcon width={11} />
    }
    case Status_DeploymentStatus.PENDING: {
      return <PauseIcon width={11} />
    }
    case Status_DeploymentStatus.UNSPECIFIED: {
      return <QuestionMarkIcon width={11} />
    }
    case Status_DeploymentStatus.UNASSIGNED: {
      return <CircleIcon width={11} />
    }
    default:
      return <div></div>
  }
}

const Badge = ({
  children,
  accentHex,
  iconColor,
}: {
  children: React.ReactNode
  accentHex: string
  iconColor: string
}) => {
  return (
    <div
      className="flex items-center justify-center rounded-sm"
      style={{
        backgroundColor: accentHex,
        color: iconColor,
        height: 14,
        width: 14,
      }}
    >
      {children}
    </div>
  )
}

export const AppStatusBadge = ({
  className,
  status,
}: {
  className?: string
  status: Status_DeploymentStatus
}) => {
  const accentColor = getColorForStatus(status)
  const accentHex = useAccent(accentColor)
  const icon = getIconForStatus(status)
  const statusDisplay = getAppStatusDisplayText(status)
  const pallete = usePalette()
  const iconColor = pallete.system.black
  return (
    <div className={clsx('flex items-center gap-2', className)}>
      <Badge accentHex={accentHex} iconColor={iconColor}>
        {icon}
      </Badge>
      <div
        className="flex flex-col gap-1"
        style={{ color: accentHex }}
        data-testid="app-status"
      >
        {statusDisplay.map((status, i) => {
          const isSecondaryStatus = i !== 0
          const secondaryStyles: React.CSSProperties = {
            color: pallete.accent.gray,
            lineHeight: '12px',
          }
          return (
            <div
              className={clsx(
                'text-2xs leading-[12px]',
                !isSecondaryStatus && 'font-semibold',
              )}
              key={status}
              style={isSecondaryStatus ? secondaryStyles : {}}
            >
              {status}
            </div>
          )
        })}
      </div>
    </div>
  )
}
