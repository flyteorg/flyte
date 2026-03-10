import { ChainIcon } from '@/components/icons/ChainIcon'
import { CheckCircleIcon } from '@heroicons/react/20/solid'
import { Tooltip } from './Tooltip'
import { useCopyToClipboard } from './CopyButton'
import { CopyIcon } from './icons/CopyIcon'

const classes = {
  chain: `[&_svg]:size-3.5 dark:text-(--system-gray-6) dark:hover:text-(--system-gray-7)`,
  copy: `[&_svg]:size-3.5 dark:text-(--system-gray-5) dark:hover:text-(--system-white)`,
}

const icons = {
  chain: ChainIcon,
  copy: CopyIcon,
}

export type CopyButtonWithTooltipProps = {
  icon?: 'chain' | 'copy'
  value: string
  textInitial?: string
  textCopied?: string
  classNameBtn?: string
}

export function CopyButtonWithTooltip({
  icon = 'copy',
  value,
  textInitial,
  textCopied = 'Copied to clipboard',
  classNameBtn,
}: CopyButtonWithTooltipProps) {
  const { copiedValue, handleCopy } = useCopyToClipboard({})
  const IconComponent = icons[icon]

  return (
    <Tooltip
      placement="bottom"
      offsetProp={4}
      contentClassName="py-1.5 px-4 shadow-[0px_8px_8px_0px_rgba(0,0,0,0.4)] dark:!bg-(--system-gray-1)"
      content={
        <div className="flex items-center gap-2">
          <span className="w-4">
            {copiedValue ? (
              <CheckCircleIcon className="size-4 dark:text-(--system-gray-7)" />
            ) : (
              <IconComponent className={classes[icon]} />
            )}
          </span>
          <span className="dark:text-(--system-gray-7)">
            {copiedValue ? textCopied : (textInitial ?? 'Copy')}
          </span>
        </div>
      }
    >
      <button
        onClick={(e) => handleCopy(e, value)}
        className={`flex cursor-pointer items-center justify-center p-1 ${classes[icon]} ${classNameBtn}`}
        aria-label={textInitial ?? 'Copy to clipboard'}
      >
        <IconComponent aria-hidden="true" />
      </button>
    </Tooltip>
  )
}
