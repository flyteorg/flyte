import React, { ComponentProps, memo, useState } from 'react'
import { CheckIcon } from '@heroicons/react/24/outline'
import { Button } from '@/components/Button'
import { CopyIcon } from './icons/CopyIcon'

type IconSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl'

interface CopyButtonProps {
  value: string
  delayMs?: number // Delay in milliseconds before resetting the copied state
  size?: IconSize
  plain?: boolean
  outline?: boolean
  title?: string
  className?: string
  color?: string
}

const iconSizes = {
  xs: '*:data-[slot=icon]:!size-2.5',
  sm: '*:data-[slot=icon]:!size-3',
  md: '*:data-[slot=icon]:!size-4',
  lg: '*:data-[slot=icon]:!size-6',
  xl: '*:data-[slot=icon]:!size-8',
}

const CopyButtonComponent: React.FC<
  CopyButtonProps & Partial<ComponentProps<typeof Button>>
> = ({
  value,
  delayMs,
  size = 'md',
  plain = true,
  outline,
  title = 'Copy to clipboard',
  className,
  color,
}) => {
  const { copiedValue, handleCopy } = useCopyToClipboard({ delayMs })

  return (
    <Button
      onClick={(e: React.MouseEvent<HTMLButtonElement, MouseEvent>) =>
        handleCopy(e, value)
      }
      size={size}
      plain={plain}
      outline={outline}
      title={title}
      className={`${className ?? ''} ${iconSizes[size]}`}
      color={color}
    >
      {copiedValue ? (
        <CheckIcon
          className="stroke-(--accent-graphic-green) transition-colors duration-200"
          data-slot="icon"
        />
      ) : (
        <CopyIcon data-slot="icon" />
      )}
    </Button>
  )
}

export const useCopyToClipboard = ({
  delayMs = 1000,
}: {
  delayMs?: number
}) => {
  const [copiedValue, setCopiedValue] = useState<string | null>(null)

  const handleCopy = (
    event: React.MouseEvent<HTMLButtonElement, MouseEvent>,
    value: string,
  ) => {
    event.stopPropagation()
    event.preventDefault()
    navigator.clipboard.writeText(value)
    setCopiedValue(value)
    setTimeout(() => setCopiedValue(null), delayMs)
  }

  return { copiedValue, handleCopy }
}

const areEqual = (prevProps: CopyButtonProps, nextProps: CopyButtonProps) =>
  prevProps.value === nextProps.value

export const CopyButton = memo(CopyButtonComponent, areEqual)
