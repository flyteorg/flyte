import type { EnrichedIdentity as EnrichedIdentity2 } from '@/gen/common/identity_pb'
import { EnrichedIdentity } from '@/gen/flyteidl2/common/identity_pb'
import {
  EnrichedIdentitySchema,
  UserSchema,
  UserSpecSchema,
} from '@/gen/common/identity_pb'
import { create } from '@bufbuild/protobuf'

/**
 * Extracts the email prefix (part before @) from an email address.
 *
 * @param email - The email address to extract the prefix from
 * @returns The email prefix (part before @) or undefined if invalid/empty
 */
export function getEmailPrefix(email: string | undefined): string | undefined {
  if (!email) return undefined

  const trimmedEmail = email.trim()
  if (!trimmedEmail) return undefined

  // If email doesn't contain "@", treat entire string as prefix (edge case)
  if (!trimmedEmail.includes('@')) {
    return trimmedEmail || undefined
  }

  const prefix = trimmedEmail.split('@')[0]?.trim()
  // Return undefined if prefix is empty (e.g., "@example.com")
  return prefix || undefined
}

/**
 * Creates an EnrichedIdentity from useIdentity() data.
 * This is used to convert the current user's identity data into the EnrichedIdentity format
 * required by UserIdentityInfo component.
 */
export function createEnrichedIdentityFromUserInfo(
  userInfo:
    | {
        givenName?: string
        familyName?: string
        email?: string
        preferredUsername?: string
      }
    | null
    | undefined,
): EnrichedIdentity2 | undefined {
  if (!userInfo) return undefined

  const email = userInfo.email || userInfo.preferredUsername
  const userSpec = create(UserSpecSchema, {
    firstName: userInfo.givenName || '',
    lastName: userInfo.familyName || '',
    email: email || '',
    organization: '',
    userHandle: '',
    groups: [],
    photoUrl: '',
  })

  const user = create(UserSchema, {
    spec: userSpec,
    roles: [],
    policies: [],
  })

  return create(EnrichedIdentitySchema, {
    principal: {
      case: 'user',
      value: user,
    },
  })
}

/**
 * Resolves user name fields with email fallback.
 * If firstName and lastName are both missing, extracts the email prefix to use as firstName.
 * Returns firstName, lastName, and the constructed userIdentityString.
 * This is the single source of truth for user name resolution logic.
 */
export function resolveUserNameFields(
  firstName: string | undefined,
  lastName: string | undefined,
  email: string | undefined,
  subject?: string,
): {
  firstName: string | undefined
  lastName: string | undefined
  userIdentityString: string
} {
  // Start with the provided firstName and lastName
  let resolvedFirstName = firstName
  const resolvedLastName = lastName

  // If both firstName and lastName are missing, use email prefix as firstName (leave lastName undefined)
  if (!resolvedFirstName && !resolvedLastName) {
    const emailPrefix = getEmailPrefix(email)
    if (emailPrefix) {
      resolvedFirstName = emailPrefix
    }
  }

  // Construct userIdentityString: if resolvedFirstName was set from email prefix above,
  // it will already be included in nameString, so no need for additional fallback.
  // Falls back to raw subject ID before "Unknown" for cases where only subject is available
  // (e.g., when identity enrichment is disabled in self-hosted deployments).
  const nameString =
    `${resolvedFirstName || ''} ${resolvedLastName || ''}`.trim()
  const userIdentityString = nameString || subject || 'Unknown'

  return {
    firstName: resolvedFirstName,
    lastName: resolvedLastName,
    userIdentityString,
  }
}

/**
 * Gets the user identity string for display purposes.
 * For users: constructs from firstName and lastName, with email prefix fallback if names are missing.
 * For applications: returns the application name.
 * Uses resolveUserNameFields internally to ensure consistent behavior.
 */
export function getUserIdentityString(
  userIdentity?: EnrichedIdentity | EnrichedIdentity2,
) {
  switch (userIdentity?.principal?.case) {
    case 'user': {
      const spec = userIdentity.principal.value.spec
      const subject = userIdentity.principal.value.id?.subject
      const { userIdentityString } = resolveUserNameFields(
        spec?.firstName,
        spec?.lastName,
        spec?.email,
        subject,
      )
      return userIdentityString
    }
    case 'application':
      return userIdentity.principal.value.spec?.name
    default:
      return 'Unknown'
  }
}
