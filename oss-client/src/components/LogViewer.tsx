import { CopyButton } from '@/components/CopyButton'
import { RunLogType } from '@/components/pages/RunDetails/types'
import {
  LogLine,
  LogLineOriginator,
} from '@/gen/flyteidl2/logs/dataplane/payload_pb'
import { getLogDateString } from '@/lib/dateUtils'
import { handleDownload } from '@/lib/download'
import { logsToString } from '@/lib/logUtils'
import { stringToColor } from '@/lib/stringToColor'
import { ArrowDownTrayIcon } from '@heroicons/react/16/solid'
import { ArrowPathIcon } from '@heroicons/react/24/outline'
import { useVirtualizer } from '@tanstack/react-virtual'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Button } from './Button'
import { LogViewerRow, SourceOption } from './LogViewerRow'
import { PopoverMenu, type MenuItem } from './Popovers'
import { SearchBar } from './SearchBar'
import { Tooltip } from './Tooltip'
import { AppLogType } from './pages/AppDetails/LogSwitch'

/** Minimum width (px) for the log viewer container and content to avoid layout thrashing / resize loops. */
export const LOG_VIEWER_MIN_WIDTH_PX = 700

interface LogViewerProps {
  enableSourceFilter?: boolean
  logs?: LogLine[]
  waiting?: boolean
  done?: boolean
  error?: Error | null
  logType?: RunLogType | AppLogType
  shouldSkipIcon?: boolean
}

const getDisplayNameByType = (type: RunLogType | AppLogType | undefined) => {
  if (type === RunLogType.K8S || type === AppLogType.SCALING) {
    return 'events'
  } else if (type === RunLogType.RUN || AppLogType.APP) {
    return 'logs'
  }
  return 'logs'
}

const getLongNameByType = (type: RunLogType | AppLogType | undefined) => {
  if (type === RunLogType.K8S) {
    return 'Kubernetes events'
  } else if (type === AppLogType.SCALING) {
    return 'scaling events'
  } else if (type === RunLogType.RUN || AppLogType.APP) {
    return 'logs'
  }
  return 'logs'
}

const EmptyStateMessage: React.FC<{
  logType?: RunLogType | AppLogType
}> = ({ logType }) => (
  <div className="flex flex-col items-center">
    <h2 className="mb-2 text-xl font-semibold text-zinc-500 dark:text-zinc-400">
      No {getDisplayNameByType(logType)}
    </h2>
    <p className="text-base text-zinc-500 dark:text-zinc-400">
      We didn&apos;t find any {getLongNameByType(logType)} for the specified
      source
    </p>
  </div>
)

const LogViewerRenderer = ({
  logs = [],
  searchQuery,
  logType,
  showEmpty,
  showStreaming,
  shouldSkipIcon,
}: LogViewerProps & {
  selectedSource: string
  searchQuery: string
  showEmpty?: boolean
  showStreaming?: boolean
  shouldSkipIcon?: boolean
}) => {
  const logsContainerRef = useRef<HTMLDivElement>(null)
  const virtualizerContainerRef = useRef<HTMLDivElement>(null)
  const scrollRafRef = useRef<number | null>(null)
  const shouldAutoScrollRef = useRef<boolean>(true)
  const isProgrammaticScrollRef = useRef<boolean>(false)

  // Stable measureElement callback
  const measureElement = useCallback(
    (el: Element | null) => el?.getBoundingClientRect().height ?? 24,
    [],
  )

  // Virtualizer for rendering only visible log rows
  const rowVirtualizer = useVirtualizer({
    count: logs.length,
    getScrollElement: () => logsContainerRef.current,
    estimateSize: () => 24, // Minimum row height based on min-h-[24px] in LogViewerRow
    // Measure actual height of each row for variable height support
    measureElement,
    overscan: 10, // Render 10 extra items outside viewport
  })

  // Check if user is at the bottom (within 10px threshold for more leniency)
  const isAtBottom = useCallback(() => {
    const container = logsContainerRef.current
    if (!container) return false

    const { scrollTop, scrollHeight, clientHeight } = container
    return scrollHeight - scrollTop - clientHeight < 10
  }, [])

  // Scroll to bottom - called when content size changes
  const scrollToBottom = useCallback(() => {
    const container = logsContainerRef.current
    if (!container || !shouldAutoScrollRef.current) return

    const totalSize = rowVirtualizer.getTotalSize()
    const containerHeight = container.clientHeight
    const targetScrollTop = Math.max(0, totalSize - containerHeight)

    // Mark as programmatic scroll
    isProgrammaticScrollRef.current = true
    container.scrollTop = targetScrollTop

    // Reset flag after scroll completes
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        isProgrammaticScrollRef.current = false
      })
    })
  }, [rowVirtualizer])

  // Track user scroll to determine if we should auto-scroll
  useEffect(() => {
    const container = logsContainerRef.current
    if (!container) return

    const handleScroll = () => {
      // Ignore scroll events from programmatic scrolling
      if (isProgrammaticScrollRef.current) return

      // Only update shouldAutoScroll based on user-initiated scrolls
      shouldAutoScrollRef.current = isAtBottom()
    }

    container.addEventListener('scroll', handleScroll, { passive: true })
    return () => container.removeEventListener('scroll', handleScroll)
  }, [isAtBottom])

  // Use ResizeObserver to detect when virtualizer container size changes
  // This fires when new items are measured and automatically scrolls to bottom
  useEffect(() => {
    const container = virtualizerContainerRef.current
    if (!container) return

    const resizeObserver = new ResizeObserver(() => {
      // Only scroll if user is at the bottom
      if (!shouldAutoScrollRef.current) return

      // Cancel any pending scroll
      if (scrollRafRef.current) {
        cancelAnimationFrame(scrollRafRef.current)
      }

      // Scroll after DOM updates
      scrollRafRef.current = requestAnimationFrame(() => {
        scrollToBottom()
        scrollRafRef.current = null
      })
    })

    resizeObserver.observe(container)
    return () => {
      resizeObserver.disconnect()
      if (scrollRafRef.current) {
        cancelAnimationFrame(scrollRafRef.current)
      }
    }
  }, [scrollToBottom])

  // Show empty state when there are no logs and showEmpty is true
  if (showEmpty && logs.length === 0) {
    return (
      <div className="relative h-full min-h-0">
        <div
          className={`flex h-full flex-col overflow-hidden bg-(--system-black)`}
        >
          <div className="w-1 rounded-tl-lg"></div>
          <div className="flex flex-1 items-center justify-center">
            <EmptyStateMessage logType={logType} />
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="relative h-full min-h-0" data-testid="logviewer">
      <div
        className={`flex h-full flex-col overflow-hidden bg-(--system-black)`}
      >
        <div className="w-1 rounded-tl-lg"></div>
        <div
          ref={logsContainerRef}
          className="min-h-0 flex-1 overflow-auto pb-4 [scrollbar-gutter:stable] [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-corner]:bg-transparent [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-zinc-400 dark:[&::-webkit-scrollbar-thumb]:bg-zinc-600 [&::-webkit-scrollbar-thumb:hover]:bg-zinc-500 dark:[&::-webkit-scrollbar-thumb:hover]:bg-zinc-500 [&::-webkit-scrollbar-track]:bg-transparent"
          style={{
            overflowAnchor: 'none',
            scrollbarWidth: 'thin',
            scrollbarColor: 'rgb(161 161 170) transparent', // zinc-400 - works reasonably in both modes
          }}
        >
          <div
            ref={virtualizerContainerRef}
            style={{
              height: `${rowVirtualizer.getTotalSize()}px`,
              width: '100%',
              minWidth: LOG_VIEWER_MIN_WIDTH_PX,
              position: 'relative',
            }}
          >
            {rowVirtualizer.getVirtualItems().map((virtualRow) => {
              const log = logs[virtualRow.index]
              if (!log) return null
              return (
                <div
                  key={virtualRow.key}
                  ref={rowVirtualizer.measureElement}
                  data-index={virtualRow.index}
                  style={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    width: '100%',
                    transform: `translateY(${virtualRow.start}px)`,
                  }}
                >
                  <LogViewerRow
                    log={log}
                    searchQuery={searchQuery}
                    skipOriginatorIcon={shouldSkipIcon}
                  />
                </div>
              )
            })}
          </div>

          {showStreaming ? (
            <div className="absolute bottom-0 left-0 w-full">
              <div className="mx-auto mt-2 flex w-4/5 items-center">
                <div className="flex-grow border-t border-dotted border-[#FCB51D]"></div>
                <span className="mx-4 flex items-center text-sm whitespace-nowrap text-[#FCB51D]">
                  <ArrowPathIcon className="mr-2 inline-block h-4 w-4 animate-spin" />
                  streaming
                </span>
                <div className="flex-grow border-t border-dotted border-[#FCB51D]"></div>
              </div>
            </div>
          ) : null}
        </div>
      </div>
    </div>
  )
}

const SourcesDropdown: React.FC<{
  selectedSource: SourceOption
  setSelectedSource: (source: SourceOption) => void
  sources: string[]
}> = ({ setSelectedSource, selectedSource, sources }) => {
  const menuItems: MenuItem[] = useMemo(() => {
    return sources.map((source) => ({
      id: source,
      label: (
        <div className="flex items-center gap-x-2 text-sm/5">
          {source !== 'All sources' && (
            <div
              className={`mr-0 h-[8px] w-[8px] rounded-[2px] ${stringToColor(source)}`}
            />
          )}
          <span className="font-medium text-(--system-gray-5)">{source}</span>
        </div>
      ),
      onClick: () => setSelectedSource(source),
      selected: source === selectedSource,
    }))
  }, [selectedSource, setSelectedSource, sources])

  return (
    <div className="min-h-0 min-w-[95px] text-(--system-gray-5)">
      <PopoverMenu label={selectedSource} items={menuItems}></PopoverMenu>
    </div>
  )
}

const LogsSearch: React.FC<{
  searchQuery: string
  setSearchQuery: (query: string) => void
}> = ({ searchQuery, setSearchQuery }) => (
  <SearchBar
    value={searchQuery}
    onChange={(e) => setSearchQuery(e.target.value)}
    onKeyDown={(e) => {
      if (e.key === 'Escape') {
        setSearchQuery('')
      }
    }}
    className="!w-[258px]"
  />
)

export const LogViewer = ({
  enableSourceFilter = true,
  logs = [],
  done = false,
  waiting,
  error,
  logType,
  shouldSkipIcon = false,
}: LogViewerProps) => {
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedSource, setSelectedSource] =
    useState<SourceOption>('All sources')
  const [showError, setShowError] = useState(false)
  const [showEmpty, setShowEmpty] = useState(false)
  const [showStreaming, setShowStreaming] = useState(false)
  const [showWaiting, setShowWaiting] = useState(false)

  // Dynamically get unique sources from logs
  // Memoize to prevent recalculation when logs array reference changes but content is the same
  const uniqueSources = useMemo(
    () =>
      Array.from(
        new Set(logs.map((log) => LogLineOriginator[log.originator])),
      ).sort(),
    [logs],
  )
  const sources: SourceOption[] = useMemo(
    () => ['All sources', ...uniqueSources],
    [uniqueSources],
  )
  const logsToCopy = useMemo(() => logsToString(logs), [logs])

  // no source filtering for K8s logs
  const rendererLogSource = useMemo(
    () => (logType !== RunLogType.K8S ? selectedSource : 'All sources'),
    [logType, selectedSource],
  )

  const shouldDisplayControls = useMemo(
    () =>
      !waiting &&
      !error &&
      (logs?.length > 0 ||
        searchQuery.length > 0 ||
        selectedSource !== 'All sources'),
    [waiting, error, logs?.length, searchQuery, selectedSource],
  )

  // Delay showing error screen by 500ms for streaming logs
  useEffect(() => {
    let timeout: NodeJS.Timeout | undefined
    if (error) {
      timeout = setTimeout(() => setShowError(true), 500)
    } else {
      setShowError(false)
    }
    return () => {
      if (timeout) clearTimeout(timeout)
    }
  }, [error])

  // Delay showing streaming indicator by 500ms to avoid flashing
  useEffect(() => {
    let timeout: NodeJS.Timeout | undefined
    if (!done) {
      timeout = setTimeout(() => setShowStreaming(true), 500)
    } else {
      setShowStreaming(false)
    }
    return () => {
      if (timeout) clearTimeout(timeout)
    }
  }, [done])

  // Delay showing waiting screen by 500ms to avoid flashing
  useEffect(() => {
    let timeout: NodeJS.Timeout | undefined
    if (waiting) {
      timeout = setTimeout(() => setShowWaiting(true), 500)
    } else {
      setShowWaiting(false)
    }
    return () => {
      if (timeout) clearTimeout(timeout)
    }
  }, [waiting])

  // Filter logs based on search query and selected source
  const filteredLogs = useMemo(() => {
    if (!logs) return []

    const searchLower = searchQuery.toLowerCase()

    return logs.filter((log) => {
      const originator =
        LogLineOriginator[
          log.originator || LogLineOriginator.UNKNOWN
        ]?.toString()
      const message = log.message || ''
      const timestamp = getLogDateString(log.timestamp)

      // Check source filter
      const matchesSource =
        selectedSource === 'All sources' || originator === selectedSource
      if (!matchesSource) return false

      // Check search filter
      if (!searchQuery) return true

      return (
        message.toLowerCase().includes(searchLower) ||
        timestamp.includes(searchQuery) ||
        originator?.toLowerCase().includes(searchLower)
      )
    })
  }, [logs, searchQuery, selectedSource])

  // Check if there are active filters (search or source selection)
  const hasActiveFilters = useMemo(
    () => searchQuery.length > 0 || selectedSource !== 'All sources',
    [searchQuery, selectedSource],
  )

  // Delay showing empty screen by 500ms for streaming logs
  useEffect(() => {
    let timeout: NodeJS.Timeout | undefined
    if (filteredLogs.length === 0 && !error && !waiting && done) {
      timeout = setTimeout(() => setShowEmpty(true), 500)
    } else {
      setShowEmpty(false)
    }
    return () => {
      if (timeout) clearTimeout(timeout)
    }
  }, [filteredLogs.length, error, waiting, done])

  // Show waiting state after delay
  if (showWaiting && waiting) {
    return (
      <div className="flex flex-1 items-center justify-center rounded-lg">
        <div className="flex flex-col items-center">
          <h2 className="mb-2 text-xl font-semibold text-zinc-500 dark:text-zinc-400">
            Waiting
          </h2>
          <p className="text-base text-zinc-500 dark:text-zinc-400">
            Logs are not available yet
          </p>
        </div>
      </div>
    )
  }

  // Show error screen after delay
  if (showError && error) {
    return (
      <div className="flex flex-1 items-center justify-center rounded-lg">
        <div className="flex flex-col items-center">
          <h2 className="mb-2 text-xl font-bold text-zinc-500 dark:text-zinc-400">
            Error
          </h2>
          <p className="text-base text-zinc-500 dark:text-zinc-400">
            We&apos;re having trouble loading the logs
          </p>
        </div>
      </div>
    )
  }

  // Show empty screen after delay - only early return if no logs and no active filters
  // If there are active filters, show empty state within the main component so controls remain visible
  if (showEmpty && !hasActiveFilters && logs.length === 0) {
    return (
      <div className={`flex flex-1 items-center justify-center rounded-lg`}>
        <EmptyStateMessage logType={logType} />
      </div>
    )
  }
  return (
    <div className="flex h-full min-h-0 flex-col gap-3">
      {shouldDisplayControls && (
        <div className="flex min-h-0 shrink-0 items-center justify-between gap-4">
          {enableSourceFilter && (
            <SourcesDropdown
              selectedSource={selectedSource}
              setSelectedSource={setSelectedSource}
              sources={sources}
            />
          )}

          <div className="gap flex flex-row items-center gap-1.25">
            {
              <LogsSearch
                searchQuery={searchQuery}
                setSearchQuery={setSearchQuery}
              />
            }

            <CopyButton value={logsToCopy} />
            <Tooltip content="Download logs" placement="bottom">
              <Button
                size="md"
                plain
                disabled={logsToCopy.length === 0}
                className="!bg-transparent enabled:hover:dark:[&_[data-slot=icon]]:!text-white"
                onClick={() => handleDownload(logsToCopy, 'logs.log')}
              >
                <ArrowDownTrayIcon data-slot="icon" />
              </Button>
            </Tooltip>
          </div>
        </div>
      )}

      <LogViewerRenderer
        searchQuery={searchQuery}
        selectedSource={rendererLogSource}
        shouldSkipIcon={shouldSkipIcon}
        waiting={waiting}
        error={error}
        done={done}
        logs={filteredLogs}
        logType={logType}
        showEmpty={showEmpty && hasActiveFilters}
        showStreaming={showStreaming}
      />
    </div>
  )
}
