import { ArrowsRightIcon } from '@/components/icons/ArrowsRightIcon'
import { ChevronRightIcon } from '@/components/icons/ChevronRightIcon'
import { LiveTimestamp } from '@/components/LiveTimestamp'
import { MenuItem, PopoverMenu } from '@/components/Popovers'
import { CacheStatusIcon } from '@/components/StatusIcons'
import { StatusIcon } from '@/components/StatusIcons/StatusIcon'
import { Tooltip } from '@/components/Tooltip'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import {
  getActionDisplayString,
  getCachingStatus,
  getTaskType,
} from '@/lib/actionUtils'
import { mapPhaseToDisplayString } from '@/lib/mapPhaseToDisplayString'
import { formatNumber } from '@/lib/numberUtils'
import { Timestamp } from '@bufbuild/protobuf/wkt'
import clsx from 'clsx'
import isNil from 'lodash/isNil'
import React, { useCallback, useMemo } from 'react'
import { useSelectedItem } from '../hooks/useSelectedItem'
import { useRunStore } from '../state/RunStore'
import { FlatRunNode } from '../state/types'
import { ActionBadge } from './ActionBadge'
import { ActionBar } from './ActionBar'
import { DisplayText } from './DisplayText'
import { GroupFolderItem } from './GroupFolderItem'
import { RunNodeTooltip } from './RunNodeTooltip'
import { useGroupTimestamp } from './useTimestamps'
import { getChildCount, getTimelineWidth } from './util'
import { getLocation } from '@/lib/windowUtils'
import { useLayoutStore } from '../state/LayoutStore'
import { CatalogCacheStatus } from '@/gen/flyteidl2/core/catalog_pb'

type RunNodeProps = {
  flatRunNode: FlatRunNode
  sidebarWidth: number
}

const ICON_SIZE = 16

const makeAttemptsIcon = (attempts: number | undefined) => {
  if (!attempts || attempts === 1) return ''
  return (
    <Tooltip content={`${attempts} attempts`} placement="bottom">
      <span className="flex items-center text-sm text-(--system-gray-5)">
        <ArrowsRightIcon />
        {attempts}
      </span>
    </Tooltip>
  )
}

const SidebarItem = ({
  attempts,
  childCount,
  endTime,
  failedCount,
  isSelected,
  displayText,
  phase,
  startTime,
  cacheStatus,
  taskType,
  maxWidth,
}: {
  attempts: number | undefined
  childCount: number | undefined
  endTime: Timestamp | undefined
  failedCount: number
  isSelected: boolean
  phase: ActionPhase | undefined
  displayText: string | undefined
  startTime: Timestamp | undefined
  cacheStatus: CatalogCacheStatus | undefined
  taskType: string | undefined
  maxWidth: number
}) => {
  const attemptsIcon = makeAttemptsIcon(attempts)
  // Don't show cache status if it's disabled to avoid clutter
  const showCacheStaus = cacheStatus !== CatalogCacheStatus.CACHE_DISABLED
  const liveTimestamp = (
    <LiveTimestamp
      decimalPlaces={1}
      timestamp={startTime}
      endTimestamp={endTime}
      className="text-2xs leading-5 font-semibold tracking-[0px] whitespace-nowrap dark:text-(--system-gray-6)"
      minWidth={0}
    />
  )

  return (
    <div className={clsx('flex w-full items-center gap-1 pl-4')}>
      {!isNil(phase) ? (
        <Tooltip content={mapPhaseToDisplayString[phase]}>
          <StatusIcon phase={phase} taskType={taskType} disableTooltip />
        </Tooltip>
      ) : null}
      <DisplayText
        displayText={displayText}
        isSelected={isSelected}
        phase={phase}
        maxWidth={maxWidth}
      />
      {childCount ? <ActionBadge childCount={childCount} /> : null}
      {failedCount ? (
        <span className="text-3xs font-semibold text-(--accent-red)">
          {formatNumber(failedCount)}
        </span>
      ) : null}
      <span className="ml-1 flex items-center gap-2">
        {showCacheStaus && (
          <CacheStatusIcon status={cacheStatus} size="md" disableTooltip />
        )}

        {startTime && liveTimestamp}
        {attemptsIcon && <span className="">{attemptsIcon}</span>}
      </span>
    </div>
  )
}

const MIN_CONTENT_PX = 200

export function RunNode({
  flatRunNode,
  sidebarWidth,
}: RunNodeProps) {
  const { depth, id, isGroup, node } = flatRunNode
  const mode = useLayoutStore((s) => s.mode)
  const shouldShowTimeline = mode === 'default' || mode === 'full-action-log'
  const onToggle = useRunStore((s) => s.toggleCollapsed)
  const cacheStatus = getCachingStatus(node?.action)
  const { setSelectedItem, selectedItem } = useSelectedItem()
  const taskName = getActionDisplayString(node?.action)
  const isExpanded = useRunStore((s) => !s.collapsedItems.has(id))
  const isThisNodeSelected = !!(selectedItem?.id && selectedItem?.id === id)
  const paddingLeft = depth * 20 + (isGroup ? 20 : 0)
  // clamp the width of the display content for correct text truncation
  const displayWidth = sidebarWidth - MIN_CONTENT_PX
  const timestamps = useGroupTimestamp({ node, isGroup })
  const childCount = getChildCount(node.childrenPhaseCounts)
  const failedCount = node?.childrenPhaseCounts
    ? node.childrenPhaseCounts[ActionPhase.FAILED]
    : 0

  const copyLinkCallback = useCallback(() => {
    const domain = node.action?.id?.run?.domain
    const host = getLocation().hostname.replace('localhost.', '')
    const project = node.action?.id?.run?.project
    const link = `https://${encodeURI(
      `${host}/v2/domain/${domain}/project/${project}/runs/${node.action?.id?.run?.name}?i=${id}`,
    )}`
    navigator.clipboard.writeText(link)
  }, [id, node])

  const hasChevron =
    depth !== 0 &&
    (node.children?.length > 0 ||
      Object.keys(node.groupChildren || {}).length > 0)

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (node?.action?.id) {
      setSelectedItem(id)
    }
  }

  const overflowMenuItems: MenuItem[] = useMemo(() => {
    if (isGroup) {
      return [
        {
          id: 'copy-group-link',
          label: 'Copy group link',
          onClick: copyLinkCallback,
        },
      ]
    } else {
      const items: MenuItem[] = [
        {
          id: 'copy-action-link',
          label: 'Copy action link',
          onClick: copyLinkCallback,
        },
      ]
      return items
    }
  }, [copyLinkCallback, isGroup])

  return (
    <Tooltip
      openDelay={700}
      content={
        <RunNodeTooltip
          attempts={node.action?.status?.attempts}
          cacheStatus={cacheStatus}
          endTime={timestamps.endTime}
          node={node}
          startTime={timestamps.startTime}
          taskName={isGroup ? node.action?.metadata?.group : taskName}
        />
      }
      placement={mode === 'full-action-log' ? 'bottom' : 'right-start'}
    >
      <div
        className={`relative grid h-[28px] cursor-pointer grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-1 rounded-md px-4 pr-0 transition-colors focus-visible:bg-(--system-gray-4) dark:focus-visible:bg-(--system-gray-2) ${
          isThisNodeSelected
            ? 'bg-(--system-gray-4) dark:bg-(--system-gray-2)'
            : 'hover:bg-(--system-gray-3) hover:dark:bg-(--system-gray-3)'
        }`}
        style={{
          paddingLeft,
        }}
        onClick={handleClick}
      >
        <div className="relative flex items-center gap-1">
          {hasChevron && (
            <div
              className={clsx(
                'absolute flex w-4 items-center justify-center pt-[2px]',
                isGroup ? 'left-[-24px]' : 'left-[-4px]',
              )}
            >
              {
                <button
                  className="cursor-pointer p-1"
                  data-testid="timeline-group-chevron"
                  onClick={(e) => {
                    e.stopPropagation()
                    onToggle(id)
                  }}
                >
                  <ChevronRightIcon
                    className={`transition-all`}
                    style={{
                      transform: isExpanded ? 'rotate(90deg)' : 'rotate(0)',
                    }}
                    width={ICON_SIZE}
                  />
                </button>
              }
            </div>
          )}

          <div className="flex w-full flex-1 items-center gap-1">
            {isGroup ? (
              <GroupFolderItem
                isSelected={isThisNodeSelected}
                displayText={node.action?.metadata?.group}
                maxWidth={displayWidth}
                node={node}
              />
            ) : (
              <SidebarItem
                attempts={node?.action?.status?.attempts}
                childCount={childCount}
                displayText={taskName}
                endTime={timestamps.endTime}
                failedCount={failedCount}
                isSelected={isThisNodeSelected}
                phase={node?.action?.status?.phase}
                startTime={timestamps.startTime}
                cacheStatus={cacheStatus}
                taskType={getTaskType(node.action?.metadata)}
                maxWidth={displayWidth}
              />
            )}
          </div>
        </div>
        {shouldShowTimeline && (
          <div
            className="pointer-events-none min-w-0 overflow-hidden"
            style={{
              width: getTimelineWidth(mode, sidebarWidth),
              right: 27,
              transition: '.1s all ease-in-out',
            }}
          >
            <ActionBar
              endTime={timestamps.endTime}
              isGroup={isGroup}
              phase={node?.action?.status?.phase}
              startTime={timestamps.startTime}
            />
          </div>
        )}
        <div className="flex h-[28px] w-6 flex-none items-center justify-center">
          <PopoverMenu
            items={overflowMenuItems}
            overflowProps={{ orientation: 'horizontal' }}
            variant="overflow"
          />
        </div>
      </div>
    </Tooltip>
  )
}
