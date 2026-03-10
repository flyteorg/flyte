import React from 'react'
import { Link } from '@/components/Link'
import { ArrowTopRightIcon } from '@/components/icons/ArrowTopRightIcon'
import clsx from 'clsx'

type ExternalLinkUrlProps = {
  icon?: React.FC<{ className?: string }>
  iconClassname?: string
  name: string
  ready?: boolean
  url: string
}

export const ExternalLinkUrl: React.FC<ExternalLinkUrlProps> = ({
  icon: Icon,
  iconClassname,
  name,
  url,
  ready = true, // callers who don't specify ready will always see the ready state
}) =>
  url ? (
    ready ? (
      <Link href={url} target="_blank" rel="noopener noreferrer">
        <div className="gap flex cursor-pointer flex-row items-center gap-2 rounded-2xl border border-(--system-gray-2) bg-(--system-gray-5) p-3 text-2xs/3 dark:border-(--system-gray-3) dark:bg-(--system-black)">
          {Icon && <Icon className="h-4 w-4" />}
          <span>{name}</span>
          <span>
            <ArrowTopRightIcon
              className={clsx(
                'h-2.125 w-2',
                iconClassname ? iconClassname : null,
              )}
            />
          </span>
        </div>
      </Link>
    ) : (
      <div className="gap flex cursor-not-allowed flex-row items-center gap-2 rounded-2xl border border-(--system-gray-2) bg-(--system-gray-5) p-3 text-2xs/3 opacity-50 dark:border-(--system-gray-3) dark:bg-(--system-black)">
        {Icon && <Icon className="h-4 w-4" />}
        <span>{name}</span>
        <span className={iconClassname}>
          <ArrowTopRightIcon
            className={clsx(
              'h-2.125 w-2',
              iconClassname ? iconClassname : null,
            )}
          />
        </span>
      </div>
    )
  ) : null
