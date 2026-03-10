import { Button, ButtonProps } from '@/components/Button'
import { ButtonConfig, SimpleDialog } from '@/components/SimpleDialog'
import {
  App,
  Status_DeploymentStatus,
} from '@/gen/flyteidl2/app/app_definition_pb'
import { useStartApp, useStopApp } from '@/hooks/useApps'
import { getStatus } from '@/lib/appUtils'
import { useMemo, useState } from 'react'

export const AppStartStopModal = ({ app }: { app: App }) => {
  const [isConfirmationDialogOpen, setConfirmationDialogOpen] = useState(false)

  const status = getStatus(app.status?.conditions)
  const isUnspecifiedOrStopped =
    Status_DeploymentStatus.UNASSIGNED === status ||
    Status_DeploymentStatus.STOPPED === status

  const startApp = useStartApp({ app })
  const stopApp = useStopApp({ app })

  const dialogButtons: ButtonConfig[] = useMemo(() => {
    return [
      {
        displayText: 'Stop App',
        color: 'rose',
        outline: true,
        onClick: async () => {
          stopApp.mutate()
          setConfirmationDialogOpen(false)
        },
      },
      {
        color: 'dark/zinc',
        displayText: 'Cancel',
        onClick: () => setConfirmationDialogOpen(false),
        outline: true,
        plain: true,
      },
    ]
  }, [stopApp])

  const [buttonColor, buttonText]: [ButtonProps['color'], string] =
    isUnspecifiedOrStopped ? ['union', 'Start App'] : ['rose', 'Stop App']

  return (
    <>
      <SimpleDialog
        buttons={dialogButtons}
        headerText={`Stop ${app.metadata?.id?.name ?? 'app'}?`}
        content={
          <>
            <div className="text-sm dark:text-(--system-gray-5)">
              Are you sure you want to stop this app? It will no longer be
              accessible from the HTTP endpoint.
            </div>
            <hr className="mt-5 dark:text-(--system-gray-5)" />
          </>
        }
        isOpen={isConfirmationDialogOpen}
        setIsOpen={setConfirmationDialogOpen}
      />
      <Button
        className="min-w-36"
        color={buttonColor}
        onClick={async () => {
          if (isUnspecifiedOrStopped) {
            startApp.mutate()
          } else {
            setConfirmationDialogOpen(true)
          }
        }}
        disabled={isConfirmationDialogOpen}
      >
        {buttonText}
      </Button>
    </>
  )
}
