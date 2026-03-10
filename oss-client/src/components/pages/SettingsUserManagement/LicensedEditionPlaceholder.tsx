'use client'

import Link from 'next/link'
import { ChartIcon } from '@/components/icons/ChartIcon'
import { UNION_AI_UPGRADE_URL } from '@/lib/constants'

interface LicensedEditionPlaceholderProps {
  /** Section title, e.g. "Users", "OAuth Apps", "Policies", "Roles" */
  title: string
  /** When true, card fills container width (use on Profile/Secrets); when false, card is content-sized (User Management). */
  fullWidth?: boolean
  /** When true, no border or shadow (e.g. formatted view inside tab sections). */
  hideBorder?: boolean
}

/**
 * Placeholder for user management features (licensed edition).
 * Light mode: white card, gray icon/text, "Upgrade" link.
 */
export function LicensedEditionPlaceholder({ title, fullWidth, hideBorder }: LicensedEditionPlaceholderProps) {
  return (
    <div
      className={`flex flex-col items-center justify-center rounded-xl bg-white px-10 py-12 text-center ${
        fullWidth ? 'min-h-[200px] w-full min-w-0' : ''
      } ${hideBorder ? '' : 'border border-(--system-gray-3) shadow-sm'}`}
    >
      <div className="flex items-center justify-center gap-2">
        <ChartIcon
          className="size-5 shrink-0 text-(--system-gray-5)"
          aria-hidden
        />
        <span className="text-base font-medium text-(--system-gray-7)">
          {title}
        </span>
      </div>
      <p className="mt-1 text-sm text-(--system-gray-5)">
        Available in the licensed edition
      </p>
      <Link
        href={UNION_AI_UPGRADE_URL}
        target="_blank"
        rel="noopener noreferrer"
        className="mt-4 text-sm font-medium text-(--union) underline underline-offset-2 hover:no-underline"
      >
        Upgrade
      </Link>
    </div>
  )
}
