'use client'

import { RunDetailsPageParams } from '@/components/pages/RunDetails/types'
import {
  ActionIdentifier,
  ActionIdentifierSchema,
} from '@/gen/flyteidl2/common/identifier_pb'
import { create } from '@bufbuild/protobuf'
import { isEqual } from 'lodash'
import { useParams } from 'next/navigation'
import { useEffect, useState } from 'react'
import { useOrg } from './useOrg'

export const useActionIdentifier = (actionId?: string | null) => {
  const org = useOrg()
  const params = useParams<RunDetailsPageParams>()
  const [actionIdentifier, setActionIdentifier] = useState<ActionIdentifier>()

  useEffect(() => {
    if (!actionId) {
      return
    }
    setActionIdentifier((prev) => {
      const newIDentifier = org
        ? create(ActionIdentifierSchema, {
            name: actionId,
            run: {
              domain: params.domain,
              name: params.runId,
              org,
              project: params.project,
            },
          })
        : undefined

      if (isEqual(prev, newIDentifier)) {
        return prev
      }
      return newIDentifier
    })
    return
  }, [actionId, org, params.domain, params.project, params.runId])

  return actionIdentifier
}
