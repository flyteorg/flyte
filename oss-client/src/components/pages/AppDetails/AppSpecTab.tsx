import { useMemo } from 'react'
import {
  DescriptionListWrapper,
  Section,
} from '@/components/DescriptionListWrapper'
import { LicensedEditionPlaceholder } from '@/components/pages/SettingsUserManagement/LicensedEditionPlaceholder'
import { TabSection } from '@/components/TabSection'
import { ExternalLinkUrl } from '@/components/ExternalLinkUrl'
import { getStatus } from '@/lib/appUtils'
import {
  App,
  Status_DeploymentStatus,
} from '@/gen/flyteidl2/app/app_definition_pb'
import {
  Resources,
  Resources_ResourceName,
} from '@/gen/flyteidl2/core/tasks_pb'
import stringify from 'safe-stable-stringify'

export const AppSpecTab = ({ app }: { app: App | undefined }) => {
  const description = app?.spec?.profile?.shortDescription
  const specJson = stringify(app?.spec)

  const links = app?.spec?.links || []
  const isActive =
    getStatus(app?.status?.conditions) === Status_DeploymentStatus.ACTIVE

  const aboutSections: Section[] = useMemo(() => {
    return [
      {
        id: 'about',
        name: '',
        items: [
          {
            name: 'Image',
            value:
              app?.spec?.appPayload.case === 'container'
                ? app?.spec?.appPayload.value.image
                : '-',
          },
          {
            name: 'Command',
            value:
              app?.spec?.appPayload.case === 'container'
                ? app?.spec?.appPayload.value.command.join(` `)
                : '-',
          },
          {
            name: 'HTTP Endpoints',
            value: app?.status?.ingress?.publicUrl,
            copyBtn: true,
          },
        ],
      },
    ]
  }, [app])

  const replicaSections: Section[] = [
    {
      id: '',
      name: '',
      items: [
        {
          name: 'Current',
          value: app?.status?.currentReplicas,
        },
        {
          name: 'Min',
          value: app?.spec?.autoscaling?.replicas?.min,
        },
        {
          name: 'Max',
          value: app?.spec?.autoscaling?.replicas?.max,
        },
      ],
    },
  ]

  const containerResources: Resources | undefined =
    app?.spec?.appPayload.case === 'container'
      ? app.spec?.appPayload?.value.resources
      : undefined

  const requestsSection: Section[] = [
    {
      id: 'Requests',
      name: '',
      items: [
        {
          name: 'Memory',
          value: containerResources?.requests.find(
            (r) => r.name === Resources_ResourceName.MEMORY,
          )?.value,
        },
        {
          name: 'CPU',
          value: containerResources?.requests.find(
            (r) => r.name === Resources_ResourceName.CPU,
          )?.value,
        },
        {
          name: 'GPU',
          value: containerResources?.requests.find(
            (r) => r.name === Resources_ResourceName.GPU,
          )?.value,
        },
        {
          name: 'Ephemeral Storage',
          value: containerResources?.requests.find(
            (r) => r.name === Resources_ResourceName.EPHEMERAL_STORAGE,
          )?.value,
        },
      ],
    },
  ]

  const limitsSection: Section[] = [
    {
      id: 'Limits',
      name: '',
      items: [
        {
          name: 'Memory',
          value: containerResources?.limits.find(
            (r) => r.name === Resources_ResourceName.MEMORY,
          )?.value,
        },
        {
          name: 'CPU',
          value: containerResources?.limits.find(
            (r) => r.name === Resources_ResourceName.CPU,
          )?.value,
        },
        {
          name: 'GPU',
          value: containerResources?.limits.find(
            (r) => r.name === Resources_ResourceName.GPU,
          )?.value,
        },
        {
          name: 'Ephemeral Storage',
          value: containerResources?.limits.find(
            (r) => r.name === Resources_ResourceName.EPHEMERAL_STORAGE,
          )?.value,
        },
      ],
    },
  ]

  return (
    <div className="flex w-full min-w-0 flex-col gap-6 [&>*:last-child]:mb-5">
      {description && (
        <div>
          <h3 className="text-sm font-bold">Description</h3>
          <p className="text-sm dark:text-(--system-gray-6)">{description}</p>
        </div>
      )}
      {links.length > 0 && isActive && (
        <div>
          <h3 className="mb-2 text-sm font-bold">Links</h3>
          <div className="flex flex-wrap gap-3">
            {links.map((l) => (
              <ExternalLinkUrl
                iconClassname="dark:text-(--system-gray-6)"
                key={l.path}
                name={l.title}
                url={`${app?.status?.ingress?.publicUrl}${l.path}`}
              ></ExternalLinkUrl>
            ))}
          </div>
        </div>
      )}

      <TabSection
        heading="About"
        copyButtonContent={specJson}
        defaultView="raw"
        sectionContent={({ isRawView }) =>
          isRawView ? (
            <DescriptionListWrapper
              rawJson={app?.spec || {}}
              isRawView={true}
              sections={[]}
            />
          ) : (
            <LicensedEditionPlaceholder title="Formatted view" fullWidth hideBorder />
          )
        }
        showRawJsonToggle
      />

      <TabSection heading="Replicas">
        <DescriptionListWrapper sections={replicaSections} isRawView={false} />
      </TabSection>

      <TabSection heading="Requests">
        <DescriptionListWrapper sections={requestsSection} isRawView={false} />
      </TabSection>

      <TabSection heading="Limits">
        <DescriptionListWrapper sections={limitsSection} isRawView={false} />
      </TabSection>
    </div>
  )
}
