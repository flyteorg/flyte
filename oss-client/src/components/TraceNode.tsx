'use client'
import React, { useState } from 'react'

export interface TraceNodeData {
  name: string
  icon: string
  duration: number
  children?: TraceNodeData[]
}

interface TraceNodeProps {
  node: TraceNodeData
  depth?: number
  isLast?: boolean
  isRoot?: boolean
  showLine?: boolean
}

export const TraceNode: React.FC<TraceNodeProps> = ({
  node,
  depth = 0,
  isLast = false,
}) => {
  const [expanded, setExpanded] = useState(true)
  const hasChildren = node.children && node.children.length > 0

  return (
    <div className="relative flex">
      {/* Left gutter with lines */}
      <div className="flex flex-col items-center" style={{ width: 20 * depth }}>
        {depth > 0 && (
          <div
            className={`w-px bg-gray-600`}
            style={{
              height: isLast ? '16px' : '100%',
              marginTop: '0px',
              marginBottom: isLast ? '0px' : '-4px',
            }}
          />
        )}
      </div>

      {/* Connector + node content */}
      <div className="relative">
        {/* Horizontal line from gutter to icon */}
        {depth > 0 && (
          <div className="absolute top-[14px] left-[-12px] h-px w-3 bg-gray-600" />
        )}

        <div
          className="flex cursor-pointer items-center space-x-2 py-1"
          onClick={() => setExpanded(!expanded)}
        >
          <span>{node.icon}</span>
          <span>{node.name}</span>
          <span className="text-sm">{node.duration.toFixed(2)}s</span>
        </div>

        {/* Children */}
        {expanded &&
          hasChildren &&
          node.children!.map((child, idx) => (
            <TraceNode
              key={idx}
              node={child}
              depth={depth + 1}
              isLast={idx === node.children!.length - 1}
            />
          ))}
      </div>
    </div>
  )
}
