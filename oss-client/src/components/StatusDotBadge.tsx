import { useAccent } from '@/hooks/usePalette'
import { AccentColor } from '@/types/colors'
import clsx from 'clsx'

type StatusDotProps = {
  color: AccentColor
  size?: number // px
  className?: string
  style?: React.CSSProperties
  filled?: boolean // true for filled, false for stroke-only
}

type StatusDotBadgeProps = {
  color?: AccentColor
  className?: string
  displayText: string
  filled?: boolean // true for filled, false for stroke-only
}

export function StatusDot({
  color,
  size = 8,
  className,
  style,
  filled = true,
}: StatusDotProps) {
  const accent = useAccent(color)

  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 8 8"
      xmlns="http://www.w3.org/2000/svg"
      focusable="false"
      className={className}
      style={{
        minWidth: '8px',
        display: 'inline-block',
        verticalAlign: 'text-bottom',
        ...style,
      }}
    >
      <circle
        cx="4"
        cy="4"
        r={filled ? 4 : 3.5}
        fill={filled ? accent : 'transparent'}
        stroke={filled ? 'transparent' : accent}
        strokeWidth={filled ? 0 : 1}
      />
    </svg>
  )
}

export const StatusDotBadge = ({
  color,
  displayText,
  className,
  filled = true,
}: StatusDotBadgeProps) => {
  return (
    <div
      className={clsx(
        `flex h-5 items-center gap-1 rounded-lg px-1.25 text-nowrap dark:bg-(--system-gray-2)`,
        className,
      )}
    >
      {color && <StatusDot color={color} filled={filled} />}
      <span>{displayText}</span>
    </div>
  )
}
