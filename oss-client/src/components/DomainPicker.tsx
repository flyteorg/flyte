import { Domain } from '@/gen/flyteidl2/project/project_service_pb'
import { useLatestProjectDomainPairs } from '@/hooks/useLatestProjects'
import { useProjectData, useProjects } from '@/hooks/useProjects'
import { useDomainStore } from '@/lib/DomainStore'
import { getUniqueDomains } from '@/lib/getUniqueDomains'
import { getUpdatedDomainPath } from '@/lib/urlUtils'
import { AppRouterInstance } from 'next/dist/shared/lib/app-router-context.shared-runtime'
import { useParams, usePathname, useRouter } from 'next/navigation'
import { useCallback, useEffect, useMemo } from 'react'
import { MenuItem, PopoverMenu } from './Popovers'

const useDomainData = () => {
  const domains = useDomainStore((s) => s.domains)
  const setDomains = useDomainStore((s) => s.setDomains)
  const selectedDomain = useDomainStore((s) => s.selectedDomain)
  const setSelectedDomain = useDomainStore((s) => s.setSelectedDomain)
  const setSelectedDomainById = useDomainStore((s) => s.setSelectedDomainById)
  const params = useParams<{ domain: string; project: string }>()
  const { data } = useProjects()
  
  useEffect(() => {
    if (data?.projects.length) {
      const uniqueDomains = getUniqueDomains(data.projects)
      setDomains(uniqueDomains)
    }
  }, [data, setDomains])

  useEffect(() => {
    if (domains.length && !selectedDomain) {
      if (params.domain) {
        setSelectedDomainById(params.domain)
      } else {
        // fallback: select first domain
        setSelectedDomain(domains[0])
      }
    }
  }, [
    domains,
    selectedDomain,
    params.domain,
    setSelectedDomainById,
    setSelectedDomain,
  ])
  return {
    domains,
    currentDomainName: selectedDomain?.name,
    currentDomainParam: params.domain,
  }
}

type DomainPickerProps = {
  shouldDisableNavigation?: boolean
}

export const DomainPicker = (props: DomainPickerProps | undefined) => {
  const shouldDisableNavigation = !!props?.shouldDisableNavigation
  const pathName = usePathname()
  const router = useRouter()
  const { currentDomainName, currentDomainParam, domains } = useDomainData()

  if (!currentDomainName) {
    return null
  }

  return (
    <DomainPickerComponent
      currentDomainName={currentDomainName}
      currentDomainParam={currentDomainParam}
      domains={domains}
      pathName={pathName}
      router={router}
      shouldDisableNavigation={shouldDisableNavigation}
    />
  )
}

const capitalize = (str: string) => {
  if (!str) return str
  return str[0].toLocaleUpperCase() + str.slice(1)
}

export const DomainPickerComponent = ({
  currentDomainName,
  currentDomainParam,
  domains,
  pathName,
  router,
  shouldDisableNavigation,
}: {
  currentDomainName: string
  currentDomainParam: string
  domains: Domain[]
  pathName: string
  router: AppRouterInstance
  shouldDisableNavigation: boolean
}) => {
  const { setLatestProjectDomain } = useLatestProjectDomainPairs()
  const paramsProject = useParams<{ project: string }>().project
  const { data } = useProjectData({ id: paramsProject })
  const setSelectedDomainById = useDomainStore((s) => s.setSelectedDomainById)
  const handleUpdateDomain = useCallback(
    (newDomain: string) => {
      setSelectedDomainById(newDomain)
      if (data) {
        const matchingDomain = domains.find((d) => d.id === newDomain)
        if (matchingDomain) {
          setLatestProjectDomain({ domain: matchingDomain, project: data })
        }
      }
      const updatedPath = getUpdatedDomainPath(pathName, newDomain)
      router.push(updatedPath)
    },
    [
      pathName,
      data,
      domains,
      router,
      setSelectedDomainById,
      setLatestProjectDomain,
    ],
  )

  const menuItems: MenuItem[] = useMemo(() => {
    const items: MenuItem[] = []

    // Add domain items
    domains.forEach((d) => {
      const isSelected = currentDomainParam === d.id
      items.push({
        id: d.id,
        type: 'item',
        label: capitalize(d.name),
        selected: isSelected,
        onClick: () => handleUpdateDomain(d.id),
      })
    })

    return items
  }, [currentDomainParam, domains, handleUpdateDomain])

  return (
    <PopoverMenu
      disabled={shouldDisableNavigation}
      items={menuItems}
      showCheckboxes={false}
      showChevron={!shouldDisableNavigation}
      portal
      size="sm"
      variant="dropdown"
      placement="bottom-start"
      label={
        <div className="text-left text-xs font-semibold capitalize">
          {currentDomainName}
        </div>
      }
      triggerClassName="text-(--system-gray-5)  border-[1.5px] border-(--system-gray-4) !py-0"
    />
  )
}
