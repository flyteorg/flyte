import { ChevronDownIcon } from '@heroicons/react/24/outline'
import { FC, useEffect, useRef, useState } from 'react'
import { MarkdownContent } from '@/components/MarkdownRenderer'

const MAX_HEIGHT = 160 // 8 lines of standard text

export const DetailsDescription: FC<{
  shortDescription?: string | undefined
  longDescription?: string | undefined
}> = ({ shortDescription, longDescription }) => {
  const [isExpanded, setIsExpanded] = useState(false)
  const contentRef = useRef<HTMLDivElement>(null)
  const [height, setHeight] = useState(MAX_HEIGHT)
  const [isOverflowing, setIsOverflowing] = useState(false) // if content is taller than MAX_HEIGHT

  useEffect(() => {
    if (longDescription) {
      const fullHeight = contentRef.current?.scrollHeight ?? 0
      setIsOverflowing(fullHeight > MAX_HEIGHT)
      setHeight(isExpanded ? fullHeight : MAX_HEIGHT)
    }
  }, [longDescription, isExpanded])

  // Don't render anything if there's no description
  if (!shortDescription && !longDescription) {
    return null
  }

  return (
    <div className="flex max-w-[584px] flex-col">
      <p className="mb-1 text-sm font-bold dark:text-(--system-white)">
        Description
      </p>
      {shortDescription ? (
        <div className="mb-3">
          <MarkdownContent text={shortDescription} />
        </div>
      ) : null}
      {longDescription ? (
        <div
          ref={contentRef}
          style={isOverflowing ? { height } : undefined}
          className="relative overflow-hidden transition-[height] duration-300 ease-in-out"
        >
          <MarkdownContent text={longDescription} />
        </div>
      ) : null}
      {isOverflowing && (
        <button
          className="flex cursor-pointer items-center gap-1 text-sm"
          onClick={() => setIsExpanded((prev) => !prev)}
        >
          <span>{isExpanded ? 'Read less' : 'Read more'}</span>
          <ChevronDownIcon
            className={`size-4 transition-all duration-300 dark:text-(--system-gray-5) ${isExpanded ? 'rotate-180' : ''}`}
          />
        </button>
      )}
    </div>
  )
}
