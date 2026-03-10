import { Domain } from '@/gen/flyteidl2/project/project_service_pb'
import { create } from 'zustand'

export type DomainState = {
  domains: Domain[]
  selectedDomain: Domain | null
  setDomains: (domains: Domain[]) => void
  setSelectedDomain: (d: Domain) => void
  setSelectedDomainById: (id: string) => void
}

export const useDomainStore = create<DomainState>((set, get) => ({
  domains: [],
  selectedDomain: null,
  setDomains: (domains) => set({ domains }),
  setSelectedDomain: (domain: Domain) => set({ selectedDomain: domain }),
  setSelectedDomainById: (id: string) => {
    const domain = get().domains.find((d) => d.id === id) || null
    set({ selectedDomain: domain })
  },
}))
