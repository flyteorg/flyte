'use client'

import {
  DescriptionListWrapper,
  Section,
} from '@/components/DescriptionListWrapper'
import { LicensedEditionPlaceholder } from '@/components/pages/SettingsUserManagement/LicensedEditionPlaceholder'
import { DetailsDescription } from '@/components/DetailsDescription'
import { TabSection } from '@/components/TabSection'

import {
  GPUAccelerator_DeviceClass,
  RuntimeMetadata_RuntimeType,
} from '@/gen/flyteidl2/core/tasks_pb'
import { useOrg } from '@/hooks/useOrg'
import { useTaskDetails } from '@/hooks/useTaskDetails'
import { getDurationStringFromProto } from '@/lib/dateUtils'
import {
  flattenCustom,
  getAcceleratorPartitionSizeValue,
  getTaskTargetAnnotations,
  getTaskTargetArgValue,
  getTaskTargetImage,
  getTaskTargetLabels,
  getTaskTargetResources,
} from '@/lib/taskUtils'
import { ArrowUpRightIcon } from '@heroicons/react/16/solid'
import { isNumber } from 'lodash'
import isNil from 'lodash/isNil'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import React, { useMemo } from 'react'
import stringify from 'safe-stable-stringify'
import { TaskDetailsPageParams } from './types'

const mapKeyValuePairs = (obj: { [key: string]: string } | undefined) => {
  return obj
    ? Object.entries(obj)
        .map(([key, value]) => `"${key}": "${value}"`)
        .join('\n')
    : undefined
}

const getCustomSection = (
  custom: Record<string, unknown> | undefined,
): Section[] => {
  if (custom == null || Object.keys(custom).length === 0) return []
  return [
    {
      id: 'custom',
      name: 'Custom',
      items: flattenCustom(custom, 0),
      fullWidthItemBorders: true,
    },
  ]
}

export const TaskDetailsTaskTab: React.FC<{
  latestVersion?: string
  version?: string
}> = ({ latestVersion, version }) => {
  const params = useParams<TaskDetailsPageParams>()
  const org = useOrg()

  const versionToRender =
    version ||
    // if no version is provided, we are showing the latest version only
    latestVersion
  const taskDetails = useTaskDetails({
    name: params.name,
    version: versionToRender!,
    project: params.project,
    domain: params.domain,
    org,
    enabled: !!versionToRender,
  })

  const { taskTemplate } = taskDetails.data?.details?.spec || {}

  const { sections, rawJson } = useMemo(() => {
    let sections: Section[] = []

    if (taskTemplate) {
      const {
        config,
        custom,
        id,
        metadata,
        securityContext,
        extendedResources,
        target,
        type,
      } = taskTemplate
      const resources = getTaskTargetResources(target)
      const labelsObj = getTaskTargetLabels(target)
      const annotationsObj = getTaskTargetAnnotations(target)
      sections = [
        {
          id: 'Main',
          name: '',
          items: [
            { name: 'Name', value: id?.name },
            { name: 'Version', value: id?.version, copyBtn: true },
            { name: 'Task type', value: type, copyBtn: true },
            {
              name: 'Image',
              value: getTaskTargetImage(target),
              url:
                metadata?.imageBuildRun?.domain &&
                metadata?.imageBuildRun?.project &&
                metadata?.imageBuildRun?.name
                  ? `/v2/domain/${metadata?.imageBuildRun?.domain}/project/${metadata?.imageBuildRun?.project}/runs/${metadata?.imageBuildRun?.name}`
                  : undefined,
              copyBtn: true,
            },
            { name: 'Config', value: mapKeyValuePairs(config) },
            { name: 'Domain', value: id?.domain },
            { name: 'Project', value: id?.project },
            { name: 'Organization', value: id?.org },
            {
              name: 'Labels',
              value:
                labelsObj !== undefined
                  ? mapKeyValuePairs(labelsObj)
                  : undefined,
            },
            {
              name: 'Annotations',
              value:
                annotationsObj !== undefined
                  ? mapKeyValuePairs(annotationsObj)
                  : undefined,
            },
          ],
        },
        ...(securityContext?.secrets?.length
          ? [
              {
                id: 'secrets',
                name: 'Secrets',
                items: securityContext.secrets.map(({ key }) => ({
                  name: 'Key',
                  value: key,
                })),
              },
            ]
          : []),
        {
          id: 'resources',
          name: 'Resources',
          items: [
            { name: 'CPU requests', value: resources.cpuRequests },
            { name: 'CPU limit', value: resources.cpuLimit },
            { name: 'Memory requests', value: resources.memoryRequests },
            { name: 'Memory limit', value: resources.memoryLimit },
            {
              name: 'Ephemeral storage requests',
              value: resources.ephemeralStorageRequests,
            },
            {
              name: 'GPU accelerator device',
              value: extendedResources?.gpuAccelerator?.device || '-',
            },
            {
              name: 'GPU accelerator device class',
              value: isNumber(extendedResources?.gpuAccelerator?.deviceClass)
                ? GPUAccelerator_DeviceClass[
                    extendedResources?.gpuAccelerator?.deviceClass
                  ] || '-'
                : '-',
            },
            {
              name: 'GPU accelerator partition size',
              value: getAcceleratorPartitionSizeValue(
                extendedResources?.gpuAccelerator?.partitionSizeValue,
              ),
            },
            { name: 'GPU requests', value: resources.gpuRequests },
          ],
        },
        ...getCustomSection(custom as Record<string, unknown>),
        {
          id: 'configuration',
          name: 'Configuration',
          items: [
            { name: 'Cache Serializable', value: metadata?.cacheSerializable },
            {
              name: 'Destination',
              value: getTaskTargetArgValue(target, '--dest'),
            },
            { name: 'Discoverable', value: metadata?.discoverable },
            { name: 'Discovery Version', value: metadata?.discoveryVersion },
            { name: 'Generates Deck', value: metadata?.generatesDeck },
            { name: 'Eager', value: metadata?.isEager },
            {
              name: 'Image cache',
              value: getTaskTargetArgValue(target, '--image-cache'),
            },
            {
              name: 'Interruptible',
              value: metadata?.interruptibleValue.value,
            },
            { name: 'Retries', value: metadata?.retries?.retries },
            { name: 'Runtime', value: metadata?.runtime?.flavor },
            {
              name: 'Runtime type',
              value: !isNil(metadata?.runtime?.type)
                ? RuntimeMetadata_RuntimeType[metadata.runtime.type]
                : '',
            },
            { name: 'Runtime version', value: metadata?.runtime?.version },
            { name: 'Tags', value: mapKeyValuePairs(metadata?.tags ?? {}) },
            {
              name: 'Timeout',
              value: getDurationStringFromProto(metadata?.timeout),
            },
            {
              name: 'tgz',
              value: getTaskTargetArgValue(target, '--tgz'),
            },
          ],
        },
      ]
    }
    return { sections, rawJson: taskTemplate }
  }, [taskTemplate])

  const copyButtonContent = stringify(taskTemplate || {}, null, 2)
  const sourceLink =
    taskDetails.data?.details?.spec?.documentation?.sourceCode?.link ?? ''
  const { documentation } = taskDetails.data?.details?.spec || {}
  return (
    <div className="flex w-full min-w-0 flex-col gap-6 p-8 pt-2.5">
      <DetailsDescription
        shortDescription={documentation?.shortDescription}
        longDescription={documentation?.longDescription}
      />
      <TabSection copyButtonContent={sourceLink} heading="Source">
        <DescriptionListWrapper
          sections={[
            {
              id: 'repository',
              name: 'Repository',
              value: sourceLink ? (
                <div className="flex items-center">
                  <Link
                    href={sourceLink}
                    className="contents hover:[&_span]:underline"
                    target="_blank"
                  >
                    <span className="truncate">{sourceLink}</span>
                    <ArrowUpRightIcon className="size-4 min-w-4 dark:text-(--system-gray-5)" />
                  </Link>
                </div>
              ) : (
                '-'
              ),
            },
          ]}
          isRawView={false}
        />
      </TabSection>
      <TabSection
        copyButtonContent={copyButtonContent}
        heading="Spec"
        defaultView="raw"
        sectionContent={({ isRawView }) =>
          isRawView ? (
            <DescriptionListWrapper
              rawJson={rawJson}
              sections={sections}
              isRawView={true}
            />
          ) : (
            <LicensedEditionPlaceholder title="Formatted view" fullWidth hideBorder />
          )
        }
        showRawJsonToggle
      />
    </div>
  )
}
