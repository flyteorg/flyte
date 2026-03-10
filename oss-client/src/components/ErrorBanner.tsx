'use client'

import { Badge, colors } from '@/components/Badge'
import {
  AbortInfo,
  ActionAttempt,
  ActionDetails,
  ErrorInfo,
  ErrorInfo_Kind,
} from '@/gen/flyteidl2/workflow/run_definition_pb'
import { isAbortInfoResult, isErrorInfoResult } from '@/lib/runTypeUtils'
import { getUserIdentityString } from '@/lib/userIdentityUtils'
import { python } from '@codemirror/lang-python'
import { XMarkIcon } from '@heroicons/react/24/outline'
import { vscodeLight } from '@uiw/codemirror-theme-vscode'
import CodeMirror, { EditorView } from '@uiw/react-codemirror'
import clsx from 'clsx'
import { useEffect, useRef } from 'react'
import { Button } from './Button'
import { CopyButton } from './CopyButton'
import { ExplainErrorButton } from './ExplainErrorButton'

export interface BannerConfig {
  badges: string[]
  textColor: string
  borderColor: string
  badgeColor: keyof typeof colors
  message: string
  backgroundColor: string
}

const getErrorInfoConfig = (error: ErrorInfo): BannerConfig => {
  const baseConfig: Pick<
    BannerConfig,
    'textColor' | 'borderColor' | 'badgeColor' | 'backgroundColor'
  > = {
    textColor: 'text-[#F87171]',
    borderColor: 'border-[#F87171]/20',
    badgeColor: 'red',
    backgroundColor: 'dark:bg-[#2F0D0D]',
  }

  const badges = (() => {
    switch (error.kind) {
      case ErrorInfo_Kind.USER:
        return ['UserError']
      case ErrorInfo_Kind.SYSTEM:
        return ['SystemError']
      case ErrorInfo_Kind.UNSPECIFIED:
      default:
        return ['UnknownError']
    }
  })()

  return {
    ...baseConfig,
    badges,
    message: error.message,
  }
}

const getAbortInfoConfig = (abort: AbortInfo): BannerConfig => {
  const userIdentityString = getUserIdentityString(abort.abortedBy)
  return {
    badges: ['Execution Aborted'],
    textColor: 'text-red-400',
    borderColor: 'border-[#F87171]/20',
    badgeColor: 'red',
    backgroundColor: 'dark:bg-[#2F0D0D]',
    message: `Aborted by ${userIdentityString}. ${abort.reason}`,
  }
}

interface BannerProps {
  config: BannerConfig
  org?: string
  domain?: string
  project?: string
  runName?: string
  actionId?: string
}

const Banner = ({
  config,
  org,
  domain,
  project,
  runName,
  actionId,
}: BannerProps) => {
  const scrollRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const scrollToBottom = () => {
      if (scrollRef.current) {
        scrollRef.current.scrollTop = scrollRef.current.scrollHeight
      }
    }
    // Use requestAnimationFrame to ensure DOM is updated
    const frameId = requestAnimationFrame(scrollToBottom)

    // Cleanup: cancel the animation frame if component unmounts before it executes
    return () => {
      cancelAnimationFrame(frameId)
    }
  }, [config])

  return (
    <div
      className={clsx(
        'w-full overflow-hidden rounded-2xl border p-4',
        config.borderColor,
        config.backgroundColor,
      )}
    >
      <div className="flex flex-1 items-center justify-between gap-3 overflow-hidden">
        <div className="flex items-center gap-2">
          {config.badges.map((badge) => (
            <Badge key={badge} color={config.badgeColor}>
              {badge}
            </Badge>
          ))}
        </div>
        <div className="flex items-center gap-1">
          <ExplainErrorButton
            org={org}
            domain={domain}
            project={project}
            runName={runName}
            actionId={actionId}
          />
          <CopyButton value={config.message} />
        </div>
      </div>

      {config.message ? (
        <div className={clsx('mt-1', config.borderColor)}>
          <div
            ref={scrollRef}
            className="max-h-[300px] overflow-y-auto [&_.cm-editor]:!bg-transparent [&_.cm-gutters]:!border-none [&_.cm-gutters]:!bg-transparent"
          >
            <CodeMirror
              readOnly
              value={config.message}
              theme={vscodeLight}
              extensions={[python(), EditorView.lineWrapping]}
              basicSetup={{
                lineNumbers: false,
                foldGutter: false,
                highlightActiveLine: false,
                highlightActiveLineGutter: false,
              }}
              style={{ fontSize: '12px' }}
            />
          </div>
        </div>
      ) : null}
    </div>
  )
}

export interface ErrorBannerProps {
  attempt?: ActionAttempt | null
  result?: ActionDetails['result']
  org?: string
  domain?: string
  project?: string
  runName?: string
  actionId?: string
}

export const ErrorBanner = ({
  attempt,
  result,
  org,
  domain,
  project,
  runName,
  actionId,
}: ErrorBannerProps) => {
  const bannerProps = { org, domain, project, runName, actionId }

  if (attempt && attempt.errorInfo) {
    return (
      <Banner config={getErrorInfoConfig(attempt.errorInfo)} {...bannerProps} />
    )
  }

  if (isErrorInfoResult(result)) {
    return <Banner config={getErrorInfoConfig(result.value)} {...bannerProps} />
  }

  if (isAbortInfoResult(result)) {
    return <Banner config={getAbortInfoConfig(result.value)} {...bannerProps} />
  }

  return null
}

export const ErrorBannerForm = ({
  message,
  onClose,
}: {
  message: string
  onClose: () => void
}) => {
  const config = getErrorInfoConfig({
    message,
    kind: ErrorInfo_Kind.USER,
  } as ErrorInfo)

  return (
    <div
      className={clsx(
        'flex w-full items-center justify-between gap-2 rounded-2xl border p-4',
        config.borderColor,
        config.backgroundColor,
        config.textColor,
      )}
    >
      <span className={`${config.textColor} text-sm font-medium`}>
        {message}
      </span>
      <Button plain onClick={onClose} size="xxs" aria-label="Close">
        <XMarkIcon data-slot="icon" aria-hidden="true" />
      </Button>
    </div>
  )
}
