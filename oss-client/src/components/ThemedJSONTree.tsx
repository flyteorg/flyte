import { JSONTree, ValueRenderer } from 'react-json-tree'

export interface ThemedJSONTreeProps {
  data: unknown
  expandLevel?: number
}

const valueRenderer: ValueRenderer = (displayValue, value) => {
  if (typeof value === 'bigint') {
    return (
      <span
        className="text-pink-400 dark:text-pink-300"
        style={{
          fontFamily: 'monospace',
        }}
      >
        {value.toString()}
      </span>
    )
  }

  return (
    <span
      className="text-green-700 dark:text-green-400"
      style={{
        fontFamily: 'monospace',
      }}
    >
      {String(displayValue)}
    </span>
  )
}

function removeTypeName(obj: unknown): unknown {
  if (Array.isArray(obj)) {
    return obj.map(removeTypeName)
  } else if (obj && typeof obj === 'object') {
    return Object.fromEntries(
      Object.entries(obj)
        .filter(([key]) => key !== '$typeName')
        .map(([key, value]) => [key, removeTypeName(value)]),
    )
  }
  return obj
}

const transparentTheme = {
  base00: 'transparent',
}

export function ThemedJSONTree({ data, expandLevel = 3 }: ThemedJSONTreeProps) {
  const cleanedData = removeTypeName(data)

  return (
    <div
      className="rounded"
      style={{ fontSize: '13px', lineHeight: '1.4', fontFamily: 'monospace' }}
    >
      <JSONTree
        data={cleanedData}
        hideRoot
        shouldExpandNodeInitially={(_, __, c) => c <= expandLevel}
        valueRenderer={valueRenderer}
        theme={transparentTheme}
      />
    </div>
  )
}

export default ThemedJSONTree
