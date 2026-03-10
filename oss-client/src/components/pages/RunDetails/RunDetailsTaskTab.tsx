import {
  DescriptionListWrapper,
  Section,
} from '@/components/DescriptionListWrapper'
import { LicensedEditionPlaceholder } from '@/components/pages/SettingsUserManagement/LicensedEditionPlaceholder'
import { useSelectedActionId } from '@/components/pages/RunDetails/hooks/useSelectedItem'
import { useSelectedAttemptStore } from '@/components/pages/RunDetails/state/AttemptStore'
import { TabSection } from '@/components/TabSection'
import {
  GPUAccelerator_DeviceClass,
  RuntimeMetadata_RuntimeType,
} from '@/gen/flyteidl2/core/tasks_pb'

import { TaskSpec } from '@/gen/flyteidl2/task/task_definition_pb'
import { useWatchActionDetails } from '@/hooks/useWatchActionDetails'
import { getDurationStringFromProto } from '@/lib/dateUtils'
import {
  flattenCustom,
  getAcceleratorPartitionSizeValue,
  getTaskTargetImage,
  getTaskTargetResources,
  getTaskTargetArgValue,
  getTaskTargetLabels,
  getTaskTargetAnnotations,
} from '@/lib/taskUtils'
import { isNumber } from 'lodash'
import isNil from 'lodash/isNil'
import React, { useMemo } from 'react'
import stringify from 'safe-stable-stringify'

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
      items: flattenCustom(custom, 5),
      fullWidthItemBorders: true,
    },
  ]
}

export const RunDetailsTaskTab: React.FC = ({}) => {
  const selectedActionId = useSelectedActionId()
  const selectedActionDetails = useWatchActionDetails(selectedActionId)
  const selectedAttempt = useSelectedAttemptStore((s) => s.selectedAttempt)
  const { spec } = selectedActionDetails.data || {}
  const { taskTemplate } = (spec?.value as TaskSpec) || {}

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
            { name: 'Cluster', value: selectedAttempt?.cluster },
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
          ].filter(Boolean),
        },
      ]
    }
    return { sections, rawJson: taskTemplate }
  }, [taskTemplate, selectedAttempt?.cluster])

  const copyButtonContent = stringify(taskTemplate || {}, null, 2)

  return (
    <div className="flex w-full min-w-0 flex-col gap-6 p-8 pt-2.5">
      <TabSection
        copyButtonContent={copyButtonContent}
        heading="Task"
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
