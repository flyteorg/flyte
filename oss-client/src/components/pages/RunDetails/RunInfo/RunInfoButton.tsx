import { Button } from '@/components/Button'
import { PaneRightIcon } from '@/components/icons/PaneRightIcon'

export const RunInfoButton = ({ onClick }: { onClick: () => void }) => {
  return (
    <Button
      className="flex cursor-pointer items-center font-semibold text-nowrap !text-(--system-gray-5) hover:!text-(--system-white)"
      outline
      onClick={onClick}
    >
      <PaneRightIcon />
      Run info
    </Button>
  )
}
