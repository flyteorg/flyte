'use client'

import type { EnrichedIdentity as EnrichedIdentity2 } from '@/gen/common/identity_pb'
import { EnrichedIdentity } from '@/gen/flyteidl2/common/identity_pb'
import {
  getUserIdentityString,
  resolveUserNameFields,
} from '@/lib/userIdentityUtils'
import clsx from 'clsx'
import trim from 'lodash/trim'
import React from 'react'
import { RobotIcon } from './icons/RobotIcon'
import { UnknownUsersIcon } from './icons/UnknownUsersIcon'
import { Tooltip } from './Tooltip'

export function stringToNumber(str: string): number {
  let sum = 0
  for (let i = 0; i < str.length; i++) {
    sum += str.charCodeAt(i)
  }
  return sum % 10
}

export const getUserInitials = (
  firstName: string | undefined,
  lastName: string | undefined,
) => {
  const firstInitial = trim(firstName || '')?.charAt(0)
  const lastInitial = trim(lastName || '')?.charAt(0) || ''
  if (!firstInitial && !lastInitial) {
    return null
  }
  return `${firstInitial}${lastInitial}`.toLocaleUpperCase()
}

const bgStyles =
  'flex h-5 min-w-5 w-5 items-center text-white justify-center dark:text-black rounded-full'
export const bgColors: Record<number, string> = {
  0: 'dark:bg-red-400 bg-red-700',
  1: 'dark:bg-orange-400 bg-orange-500',
  2: 'dark:bg-yellow-400 bg-yellow-500',
  3: 'dark:bg-lime-400 bg-lime-700',
  4: 'dark:bg-teal-400 bg-teal-700',
  5: 'dark:bg-blue-400 bg-blue-700',
  6: 'dark:bg-indigo-400 bg-indigo-700',
  7: 'dark:bg-purple-400 bg-purple-700',
  8: 'dark:bg-fuchsia-400 bg-fuchsia-700',
  9: 'dark:bg-pink-400 bg-pink-700',
}

export const UserIcon = ({
  firstName,
  lastName,
  showUserName,
  userIdentityString,
  userNameClass,
}: {
  firstName: string | undefined
  lastName: string | undefined
  showUserName?: boolean
  userNameClass?: string
  userIdentityString: string | undefined
}) => {
  const initials = getUserInitials(firstName, lastName)
  const bgColor =
    bgColors[stringToNumber(userIdentityString?.trim() ?? '')] ?? bgColors[0]
  return (
    <div className="flex items-center gap-2">
      <div className={`${bgStyles} ${bgColor} text-3xs font-medium`}>
        {initials}
      </div>
      {showUserName && (
        <div
          className={clsx(
            'text-sm font-medium text-(--system-white)',
            userNameClass,
          )}
        >
          {userIdentityString}
        </div>
      )}
    </div>
  )
}

export interface UserIdentityInfoProps {
  executedBy?: EnrichedIdentity | EnrichedIdentity2
  showUserName?: boolean
  hoverContent?: React.ReactNode
  userNameClass?: string
}
export const UserIdentityInfo = ({
  executedBy,
  showUserName = true,
  hoverContent,
  userNameClass,
}: UserIdentityInfoProps) => {
  let userIdentityString: string
  let iconElement: React.ReactNode

  if (executedBy) {
    switch (executedBy.principal?.case) {
      case 'user': {
        const spec = executedBy.principal.value.spec
        const subject = executedBy.principal.value.id?.subject
        // Call resolveUserNameFields once - it provides firstName, lastName, and userIdentityString
        const {
          firstName,
          lastName,
          userIdentityString: resolvedUserIdentityString,
        } = resolveUserNameFields(spec?.firstName, spec?.lastName, spec?.email, subject)

        userIdentityString = resolvedUserIdentityString

        iconElement = (
          <div className="flex items-center gap-3">
            <UserIcon
              firstName={firstName}
              lastName={lastName}
              userIdentityString={resolvedUserIdentityString}
              showUserName={showUserName}
              userNameClass={userNameClass}
            />
          </div>
        )
        break
      }
      case 'application':
        userIdentityString = getUserIdentityString(executedBy) || 'Unknown'
        iconElement = (
          <div className="flex items-center gap-3">
            <div className={`${bgStyles} bg-zinc-700 dark:bg-zinc-400`}>
              <RobotIcon width="12" height="12" />
            </div>
            {showUserName && (
              <div className="text-sm font-medium text-(--system-white)">
                {userIdentityString}
              </div>
            )}
          </div>
        )
        break
      default:
        userIdentityString = getUserIdentityString(executedBy) || 'Unknown'
        iconElement = (
          <div className="flex items-center gap-3">
            <div className={`${bgStyles} bg-zinc-700 dark:bg-zinc-400`}>
              <UnknownUsersIcon width="12" height="12" />
            </div>
            {showUserName && (
              <div className="text-sm font-medium text-(--system-white)">
                {userIdentityString}
              </div>
            )}
          </div>
        )
    }
  } else {
    // No data provided
    userIdentityString = 'Unknown'
    iconElement = (
      <div className="flex items-center gap-3">
        <div className={`${bgStyles} bg-zinc-700 dark:bg-zinc-400`}>
          <UnknownUsersIcon width="12" height="12" />
        </div>
        {showUserName && (
          <div className="text-sm font-medium text-(--system-white)">
            {userIdentityString}
          </div>
        )}
      </div>
    )
  }

  // Add a wrapper with accessibility/test attributes
  const wrapperProps = {
    'aria-label': userIdentityString ? `User: ${userIdentityString}` : 'User',
    'data-testid': 'user-identity-info',
  }

  const content = <span {...wrapperProps}>{iconElement}</span>

  return hoverContent ? (
    <Tooltip content={hoverContent} placement="bottom">
      {content}
    </Tooltip>
  ) : (
    content
  )
}
