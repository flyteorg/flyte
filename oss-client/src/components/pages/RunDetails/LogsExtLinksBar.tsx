import { ArrowTopRightIcon } from '@/components/icons/ArrowTopRightIcon'
import { Link } from '@/components/Link'
import { useExternalLogUrls } from '@/components/pages/RunDetails/hooks/useExternalLogUrls'
import { ExternalLinkUrlProps } from '@/components/pages/RunDetails/types'
import { ActionAttempt } from '@/gen/flyteidl2/workflow/run_definition_pb'

const ExternalLinkUrl: React.FC<ExternalLinkUrlProps> = ({
  icon: Icon,
  name,
  url,
  ready,
}) =>
  url ? (
    ready ? (
      <Link href={url} target="_blank" rel="noopener noreferrer">
        <div className="gap flex cursor-pointer flex-row items-center gap-1 text-xs/5 font-medium tracking-[0.25px]">
          {Icon && <Icon className="h-3.5 w-3.5" />}
          <span>{name}</span>
          <ArrowTopRightIcon className="h-2 w-2" />
        </div>
      </Link>
    ) : (
      <div className="gap flex cursor-not-allowed flex-row items-center gap-1 text-xs/5 opacity-50">
        {Icon && <Icon className="h-3.5 w-3.5" />}
        <span>{name}</span>
        <ArrowTopRightIcon className="h-2 w-2" />
      </div>
    )
  ) : null

const LogsExtLinksBar: React.FC<{
  logInfo?: ActionAttempt['logInfo']
}> = ({ logInfo }) => {
  const urls = useExternalLogUrls(logInfo)

  return (
    <div className="flex flex-wrap items-center gap-x-5 text-(--system-gray-5) hover:text-(--system-white)">
      {urls.map((linkProps, key) => (
        <ExternalLinkUrl {...linkProps} key={key} />
      ))}
    </div>
  )
}

export default LogsExtLinksBar
