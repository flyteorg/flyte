import { Badge } from '@/components/Badge'
import { formatNumber } from '@/lib/numberUtils'

export const ActionBadge = ({ childCount }: { childCount: number }) => (
  <Badge
    className="h-[13px] !rounded-[3px] !px-[2px] !py-0 font-mono !text-[10px] !leading-none font-semibold"
    color="gray"
  >
    {' '}
    {formatNumber(childCount)}
  </Badge>
)
