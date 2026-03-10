import { Dispatch, SetStateAction, useMemo } from 'react'
import { usePathname } from 'next/navigation'
import { useRunDetails } from '@/hooks/useRunDetails'
import Drawer from '@/components/Drawer'
import { RunInfoContent } from './RunInfoContent'

type RunInfoDrawerProps = {
  isOpen: boolean
  setIsOpen: Dispatch<SetStateAction<boolean>>
}

export const RunInfoDrawer = ({ isOpen, setIsOpen }: RunInfoDrawerProps) => {
  const pathname = usePathname()
  const runId = useMemo(() => {
    const pathParts = pathname.split('/')
    return pathParts[pathParts.length - 1]
  }, [pathname])

  const runDetailsQuery = useRunDetails(runId)

  return (
    <Drawer
      isOpen={isOpen}
      setIsOpen={setIsOpen}
      size={758}
      title={'Run info'}
      tabs={<RunInfoContent runDetails={runDetailsQuery} />}
    />
  )
}
