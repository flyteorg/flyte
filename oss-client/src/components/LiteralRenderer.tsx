import { LiteralType } from '@/gen/flyteidl2/core/types_pb'
import { NamedLiteral } from '@/gen/flyteidl2/task/common_pb'
import { literalToDisplayValue } from '@/lib/literalToDisplayValue'
import { isMarkdown } from '@/lib/stringUtils'
import { literalTypeToString } from '@/lib/variableToString'
import React from 'react'
import ReactMarkdown from 'react-markdown'
import rehypeHighlight from 'rehype-highlight'

export interface LiteralRendererProps {
  literal: NamedLiteral
  type: LiteralType
}

export const LiteralRenderer: React.FC<LiteralRendererProps> = ({
  literal,
  type,
}) => {
  const { name, value } = literal
  const typeString = type ? literalTypeToString(type) : 'unknown'
  const valueString = literalToDisplayValue(value)
  const isMarkdownRef = isMarkdown(valueString)
  return (
    <div className="flex w-full items-center px-3 py-3 text-zinc-900 dark:text-white dark:text-zinc-100">
      <div className="flex w-1/2 items-baseline justify-start gap-1 text-sm">
        <span className="align-middle leading-5 font-medium">{name}</span>
        {/* TODO: get required info from the type */}
        <span className="text-zinc-400 dark:text-zinc-400">
          ({typeString})*
        </span>
      </div>
      <div
        className={`flex w-1/2 justify-start text-sm dark:text-white ${isMarkdownRef ? 'flex-col' : 'font-semibold'}`}
      >
        {isMarkdownRef ? (
          <div className="prose-sm prose max-w-none dark:prose-invert">
            <ReactMarkdown rehypePlugins={[rehypeHighlight]}>
              {valueString}
            </ReactMarkdown>
          </div>
        ) : (
          valueString
        )}
      </div>
    </div>
  )
}

export default LiteralRenderer
