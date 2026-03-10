import { DescriptionListWrapper } from '@/components/DescriptionListWrapper'
import { DetailsDescription } from '@/components/DetailsDescription'
import { TabSection } from '@/components/TabSection'
import { useOrg } from '@/hooks/useOrg'
import { useGetTriggerDetails } from '@/hooks/useTriggers'
import { useParams } from 'next/navigation'
import { useMemo } from 'react'
import stringify from 'safe-stable-stringify'
import { TriggerDetailsPageParams } from './types'

export const TriggerSpecTab = () => {
  const params = useParams<TriggerDetailsPageParams>()
  const org = useOrg()

  const decodedName = useMemo(
    () => (params.name ? decodeURIComponent(params.name) : ''),
    [params.name],
  )

  const triggerDetailsQuery = useGetTriggerDetails({
    org,
    name: decodedName,
    projectId: params.project,
    domain: params.domain,
    taskName: params.taskName,
  })
  // TODO - get correct translated json once we have the type definitions coming back with triggers

  const copyButtonContent = stringify(
    triggerDetailsQuery.data?.trigger || {},
    null,
    2,
  )
  return (
    <div className="flex w-full min-w-0 flex-col gap-6 p-8 pt-2.5">
      <DetailsDescription
        longDescription={triggerDetailsQuery.data?.trigger?.spec?.description}
      />
      <TabSection copyButtonContent={copyButtonContent} heading="Spec">
        <DescriptionListWrapper
          rawJson={triggerDetailsQuery.data?.trigger || {}}
          isRawView
          sections={[]}
        />
      </TabSection>
    </div>
  )
}
