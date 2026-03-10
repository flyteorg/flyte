import { useExternalLogUrls } from '@/components/pages/RunDetails/hooks/useExternalLogUrls'
import { ExternalLinkUrl } from '@/components/ExternalLinkUrl'
import { ActionAttempt } from '@/gen/flyteidl2/workflow/run_definition_pb'
import React from 'react'
import { TaskLog_LinkType } from '@/gen/flyteidl2/core/execution_pb'

export const ActionDashboardLinks: React.FC<{
  attempt?: ActionAttempt | null
}> = ({ attempt }) => {
  const urls = useExternalLogUrls(attempt?.logInfo, TaskLog_LinkType.DASHBOARD)

  return urls.length ? (
    <div>
      <h3 className="py-2 text-sm/5 font-bold">Links</h3>
      <div className="mt-2 flex flex-row flex-wrap gap-2">
        {urls.map((url, index) => (
          <ExternalLinkUrl {...url} key={index} />
        ))}
      </div>
    </div>
  ) : null
}
