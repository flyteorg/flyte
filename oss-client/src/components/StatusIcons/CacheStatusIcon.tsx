import { Tooltip } from '@/components/Tooltip'
import { IconSize, iconSizeMap } from '@/components/StatusIcons/iconSize'
import { useMemo } from 'react'
import CacheErrorIcon from '@/components/icons/CacheErrorIcon'
import CacheReadIcon from '@/components/icons/CacheReadIcon'
import CacheWrittenIcon from '@/components/icons/CacheWrittenIcon'
import CacheDisabledIcon from '@/components/icons/CacheDisabledIcon'
import { CatalogCacheStatus } from '@/gen/flyteidl2/core/catalog_pb'

interface CacheStatusIconProps {
  className?: string
  status?: CatalogCacheStatus
  size?: IconSize
  disableTooltip?: boolean
}

export const mapCacheStatusToDisplayString: Record<CatalogCacheStatus, string> =
  {
    [CatalogCacheStatus.CACHE_DISABLED]: 'Disabled',
    [CatalogCacheStatus.CACHE_LOOKUP_FAILURE]: 'Lookup failure',
    [CatalogCacheStatus.CACHE_EVICTED]: 'Evicted',
    [CatalogCacheStatus.CACHE_HIT]: 'Hit',
    [CatalogCacheStatus.CACHE_MISS]: 'Miss',
    [CatalogCacheStatus.CACHE_POPULATED]: 'Populated',
    [CatalogCacheStatus.CACHE_PUT_FAILURE]: 'Put failure',
    [CatalogCacheStatus.CACHE_SKIPPED]: 'Skipped',
  }

export const CacheErrorStatuses = [
  CatalogCacheStatus.CACHE_LOOKUP_FAILURE,
  CatalogCacheStatus.CACHE_PUT_FAILURE,
]

export const CacheStatusIcon: React.FC<CacheStatusIconProps> = ({
  status,
  className,
  size = 'md',
  disableTooltip,
}) => {
  const isError = !!status && CacheErrorStatuses.includes(status)
  const iconClass = `${iconSizeMap[size]} ${className ?? ''} ${isError ? 'text-orange-400' : 'text-[#777777]'}`

  const icon = useMemo(() => {
    switch (status) {
      case CatalogCacheStatus.CACHE_PUT_FAILURE:
      case CatalogCacheStatus.CACHE_LOOKUP_FAILURE:
        return <CacheErrorIcon className={iconClass} />
      case CatalogCacheStatus.CACHE_HIT:
        return <CacheReadIcon className={iconClass} />
      case CatalogCacheStatus.CACHE_POPULATED:
        return <CacheWrittenIcon className={iconClass} />
      case CatalogCacheStatus.CACHE_DISABLED:
        return <CacheDisabledIcon className={iconClass} />
      default:
        return null
    }
  }, [iconClass, status])

  if (disableTooltip) {
    return icon
  }

  return status ? (
    <div onMouseOver={(e) => e.stopPropagation()}>
      <Tooltip
        content={mapCacheStatusToDisplayString[status]}
        placement="bottom"
      >
        {icon}
      </Tooltip>
    </div>
  ) : (
    icon
  )
}
