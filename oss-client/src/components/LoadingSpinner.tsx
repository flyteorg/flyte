import React, { useEffect, useState } from 'react'

const colors = {
  union: 'border-(--union)',
  zinc: 'border-zinc-500',
}

const sizes = {
  sm: 'h-4 w-4 border-2',
  md: 'h-10 w-10 border-4',
}
interface LoadingSpinnerProps {
  delay?: number
  size?: 'sm' | 'md'
  color?: 'union' | 'zinc'
  text?: string
}

export const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
  delay = 500,
  size = 'md',
  color = 'zinc',
  text,
}) => {
  const [show, setShow] = useState(false)

  useEffect(() => {
    const timer = setTimeout(() => setShow(true), delay)
    return () => clearTimeout(timer)
  }, [delay])

  if (!show) return null

  return (
    <div className="flex h-full w-full min-w-full items-center justify-center">
      <span
        className={`${colors[color]} ${sizes[size]} inline-block animate-spin rounded-full border-t-transparent`}
      />
      {text ? (
        <span className="ml-2 text-sm tracking-wide dark:text-(--system-gray-5)">
          {text}
        </span>
      ) : null}
    </div>
  )
}
