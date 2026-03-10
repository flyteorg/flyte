'use client'

import Link from 'next/link'
import { Button } from '@/components/Button'
import { Tooltip } from '@/components/Tooltip'
import { UNION_AI_UPGRADE_URL } from '@/lib/constants'

const LICENSED_EDITION_TOOLTIP_CONTENT = (
  <>
    Available in the licensed edition.{' '}
    <Link
      href={UNION_AI_UPGRADE_URL}
      target="_blank"
      rel="noopener noreferrer"
      className="font-medium text-(--union) underline underline-offset-2 hover:no-underline"
      onClick={(e) => e.stopPropagation()}
    >
      Upgrade →
    </Link>
  </>
)

const LICENSED_EDITION_TOOLTIP_CLASS =
  '!bg-(--system-gray-2) !text-(--system-gray-7) px-3 py-2 text-sm font-normal border border-(--system-gray-3)'

interface DisabledButtonWithTooltipProps {
  children: React.ReactNode
}

/**
 * Wraps a disabled button so it still shows a hover tooltip
 * (disabled elements may not receive pointer events).
 */
export function DisabledButtonWithTooltip({ children }: DisabledButtonWithTooltipProps) {
  return (
    <Tooltip
      content={LICENSED_EDITION_TOOLTIP_CONTENT}
      contentClassName={LICENSED_EDITION_TOOLTIP_CLASS}
      placement="bottom"
    >
      <span className="inline-block cursor-not-allowed" aria-disabled>
        {children}
      </span>
    </Tooltip>
  )
}
