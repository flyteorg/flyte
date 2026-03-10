'use client'

import { ChevronRightIcon } from '@/components/icons/ChevronRightIcon'
import { DocumentIcon } from '@/components/icons/DocumentIcon'
import { FolderIcon } from '@/components/icons/FolderIcon'
import {
  findFirstFile,
  getFileContent,
  getLanguageFromPath,
  type DirectoryNode,
} from '@/lib/tgzUtils'
import { CodeViewer } from '@/components/CodeViewer'
import React, { useEffect, useMemo, useRef, useState } from 'react'

type DirectoryTreeViewProps = {
  node: DirectoryNode
  selectedFile: DirectoryNode | null
  onSelectFile: (file: DirectoryNode) => void
}

export const DirectoryTreeView: React.FC<DirectoryTreeViewProps> = ({
  node,
  selectedFile,
  onSelectFile,
}) => {
  const [isTreeCollapsed, setIsTreeCollapsed] = useState(true)
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const autoSelectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const hasAutoExpandedRef = useRef(false)

  // Determine display node: skip root if it has empty name or only one child
  // TODO: is this the best way to do this?
  const displayNode = useMemo(() => {
    if (!node) return null
    if (!node.name || node.name === '') return node
    if (node.children?.length === 1) return node.children[0]
    return node
  }, [node])

  // Reset expansion state when the root changes
  useEffect(() => {
    hasAutoExpandedRef.current = false
    setExpanded(new Set())
  }, [displayNode])

  // Auto-expand only once per root load; afterwards user choices persist
  useEffect(() => {
    if (!displayNode) return
    if (hasAutoExpandedRef.current) return

    const collectAllDirectoryPaths = (
      root: DirectoryNode,
      paths: Set<string> = new Set(),
    ): Set<string> => {
      if (root.isDirectory) {
        paths.add(root.path)
        root.children?.forEach((child) =>
          collectAllDirectoryPaths(child, paths),
        )
      }
      return paths
    }

    setExpanded(collectAllDirectoryPaths(displayNode))
    hasAutoExpandedRef.current = true
  }, [displayNode])

  // Auto-select first file if none selected
  useEffect(() => {
    if (!displayNode) return
    if (autoSelectTimeoutRef.current) {
      clearTimeout(autoSelectTimeoutRef.current)
      autoSelectTimeoutRef.current = null
    }

    if (selectedFile) return

    autoSelectTimeoutRef.current = setTimeout(() => {
      if (selectedFile) return
      const firstFile = findFirstFile(displayNode)
      if (firstFile) {
        onSelectFile(firstFile)
      }
      autoSelectTimeoutRef.current = null
    }, 100)

    return () => {
      if (autoSelectTimeoutRef.current) {
        clearTimeout(autoSelectTimeoutRef.current)
        autoSelectTimeoutRef.current = null
      }
    }
  }, [displayNode, selectedFile, onSelectFile])

  const toggleExpand = (path: string) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(path)) {
        next.delete(path)
      } else {
        next.add(path)
      }
      return next
    })
  }

  const renderFileTree = (
    node: DirectoryNode,
    depth: number = 0,
  ): React.ReactNode => {
    const isDirectory = node.isDirectory ?? false
    const isExpanded = expanded.has(node.path)
    const hasChildren = (node.children?.length ?? 0) > 0
    const isSelected = selectedFile?.path === node.path
    const isEmptyRoot = depth === 0 && (!node.name || node.name === '')

    // Skip rendering root if it has empty name - render children at depth 0
    if (isEmptyRoot && hasChildren) {
      return <>{node.children.map((child) => renderFileTree(child, 0))}</>
    }

    const itemClasses = isSelected
      ? 'bg-blue-100 text-blue-900 dark:bg-blue-500/30 dark:text-white'
      : 'text-zinc-700 hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800/50'

    return (
      <div key={node.path} className="select-none">
        <div
          className={`flex h-7 cursor-pointer items-center gap-1.5 rounded text-sm transition-colors ${itemClasses}`}
          style={{ paddingLeft: `${depth * 22 + 4}px`, paddingRight: '8px' }}
          onClick={() => {
            if (isDirectory) {
              toggleExpand(node.path)
            } else {
              onSelectFile(node)
            }
          }}
        >
          {isDirectory ? (
            <>
              <div className="flex h-4 w-4 flex-shrink-0 items-center justify-center">
                <ChevronRightIcon
                  className={`h-3 w-3 flex-shrink-0 text-zinc-500 transition-transform dark:!fill-zinc-200 ${
                    isExpanded ? 'rotate-90' : ''
                  }`}
                />
              </div>
              <FolderIcon
                className="h-4 w-4 flex-shrink-0 text-blue-500 dark:text-blue-400"
                fill="none"
                stroke="currentColor"
              />
              <span className="min-w-0 flex-1 truncate">{node.name}</span>
            </>
          ) : (
            <>
              <div className="h-4 w-4 flex-shrink-0" />
              <DocumentIcon
                className="h-4 w-4 flex-shrink-0 text-zinc-600 dark:text-zinc-300"
                fill="none"
                stroke="currentColor"
              />
              <span className="min-w-0 flex-1 truncate">{node.name}</span>
            </>
          )}
        </div>
        {isDirectory && isExpanded && hasChildren && (
          <div>
            {node.children.map((child) => renderFileTree(child, depth + 1))}
          </div>
        )}
      </div>
    )
  }

  if (!displayNode) {
    return (
      <div className="text-sm text-(--system-gray-5)">
        No root node provided
      </div>
    )
  }

  const fileContent = selectedFile ? getFileContent(selectedFile) : null
  const headerClasses =
    'flex h-8 flex-shrink-0 items-center border-b border-zinc-200 bg-zinc-50 px-3 dark:border-zinc-700 dark:bg-zinc-900'

  return (
    <>
      <style
        dangerouslySetInnerHTML={{
          __html: `
          .directory-tree-scrollbar::-webkit-scrollbar,
          .codemirror-scrollbar .cm-scroller::-webkit-scrollbar,
          .monaco-scrollable-element > .scrollbar {
            width: 8px !important;
            height: 8px !important;
          }
          .directory-tree-scrollbar::-webkit-scrollbar-track,
          .codemirror-scrollbar .cm-scroller::-webkit-scrollbar-track,
          .monaco-scrollable-element > .scrollbar > .slider {
            background: transparent !important;
          }
          .directory-tree-scrollbar::-webkit-scrollbar-thumb,
          .codemirror-scrollbar .cm-scroller::-webkit-scrollbar-thumb,
          .monaco-scrollable-element > .scrollbar > .slider {
            background: rgb(212 212 212) !important;
            border-radius: 9999px !important;
          }
          .directory-tree-scrollbar::-webkit-scrollbar-thumb:hover,
          .codemirror-scrollbar .cm-scroller::-webkit-scrollbar-thumb:hover,
          .monaco-scrollable-element > .scrollbar > .slider:hover {
            background: rgb(163 163 163) !important;
          }
          .directory-tree-scrollbar::-webkit-scrollbar-corner,
          .codemirror-scrollbar .cm-scroller::-webkit-scrollbar-corner {
            background: transparent;
          }
          .dark .directory-tree-scrollbar::-webkit-scrollbar-thumb,
          .dark .codemirror-scrollbar .cm-scroller::-webkit-scrollbar-thumb,
          .dark .monaco-scrollable-element > .scrollbar > .slider {
            background: rgb(63 63 70) !important;
          }
          .dark .directory-tree-scrollbar::-webkit-scrollbar-thumb:hover,
          .dark .codemirror-scrollbar .cm-scroller::-webkit-scrollbar-thumb:hover,
          .dark .monaco-scrollable-element > .scrollbar > .slider:hover {
            background: rgb(82 82 91) !important;
          }
        `,
        }}
      />
      <div className="relative flex h-full w-full flex-1 overflow-hidden">
        {/* Collapsible sidebar */}
        <div
          className={`flex flex-shrink-0 border-r border-zinc-200 bg-zinc-50 transition-all duration-200 dark:border-zinc-700 dark:bg-zinc-900 ${
            isTreeCollapsed ? 'w-8' : 'w-64'
          }`}
        >
          {isTreeCollapsed ? (
            <div className="flex w-full items-start justify-center pt-2">
              <button
                onClick={() => setIsTreeCollapsed(false)}
                className="flex h-7 w-7 items-center justify-center rounded text-zinc-700 transition-colors hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800/50"
                title="Expand file explorer"
              >
                <ChevronRightIcon className="h-4 w-4" />
              </button>
            </div>
          ) : (
            <div className="flex w-full flex-col overflow-hidden">
              <div className={headerClasses}>
                <div className="flex flex-1 items-center gap-1.5 text-sm text-zinc-700 dark:text-zinc-300">
                  <FolderIcon
                    className="h-4 w-4 flex-shrink-0 text-blue-500 dark:text-blue-400"
                    fill="none"
                    stroke="currentColor"
                  />
                  <span className="min-w-0 flex-1 truncate font-medium">
                    Files
                  </span>
                </div>
                <button
                  onClick={() => setIsTreeCollapsed(true)}
                  className="flex h-6 w-6 items-center justify-center rounded text-zinc-700 transition-colors hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800/50"
                  title="Collapse file explorer"
                >
                  <ChevronRightIcon className="h-3 w-3 rotate-90" />
                </button>
              </div>
              <div className="directory-tree-scrollbar flex-1 overflow-auto">
                <div className="min-w-max p-2">
                  {renderFileTree(displayNode, 0)}
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Content area */}
        <div className="flex min-w-0 flex-1 flex-col">
          {selectedFile ? (
            <>
              <div className={headerClasses}>
                <DocumentIcon
                  className="mr-2 h-4 w-4 text-zinc-600 dark:text-zinc-300"
                  fill="none"
                  stroke="currentColor"
                />
                <span className="truncate text-sm font-medium text-zinc-700 dark:text-zinc-200">
                  {selectedFile.path}
                </span>
              </div>
              <div className="flex min-h-0 flex-1">
                {fileContent !== null && selectedFile?.path ? (
                  <CodeViewer
                    content={fileContent}
                    language={getLanguageFromPath(selectedFile.path)}
                    theme="light"
                    filePath={selectedFile.path}
                  />
                ) : (
                  <div className="flex h-full items-center justify-center p-8">
                    <div className="text-center text-sm text-(--system-gray-5) dark:text-(--system-gray-4)">
                      <p className="mb-2">Loading …</p>
                      <p className="text-xs">
                        Waiting for file content or language/theme to be ready.
                      </p>
                    </div>
                  </div>
                )}
              </div>
            </>
          ) : (
            <div className="flex flex-1 items-center justify-center text-(--system-gray-5) dark:text-zinc-500">
              <div className="text-center">
                <p className="mb-1 text-sm">No file selected</p>
                <p className="text-xs">
                  Click on a file in the explorer to view its contents
                </p>
              </div>
            </div>
          )}
        </div>
      </div>
    </>
  )
}
