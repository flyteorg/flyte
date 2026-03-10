'use client'

import { CreateProjectWorkflow } from '@/components/pages/Projects/components/CreateProjectWorkflow.tsx'
import { ProjectArchiveRestoreDialog } from '@/components/pages/Projects/components/ProjectArchiveRestoreDialog'
import { ProjectsHeaderDropdown } from '@/components/pages/Projects/components/ProjectsHeaderDropdown'
import { ProjectsItemDropdown } from '@/components/pages/Projects/components/ProjectsItemDropdown'
import { useProjectArchiveRestoreDialog } from '@/components/pages/Projects/components/useProjectArchiveRestoreDialog.ts'
import { SearchBar } from '@/components/SearchBar'
import { Table, TableState } from '@/components/Tables'
import { Tooltip } from '@/components/Tooltip'
import {
  Project,
  ProjectState,
} from '@/gen/flyteidl2/project/project_service_pb'

import { useLatestProjectDomainPairs } from '@/hooks/useLatestProjects'
import { useProjects } from '@/hooks/useProjects'
import { useDomainStore } from '@/lib/DomainStore'
import { highlightMatches } from '@/lib/highlightMatches'
import { createColumnHelper } from '@tanstack/react-table'
import { useQueryState } from 'nuqs'
import React, { useMemo, useState } from 'react'

const Badge = ({ children }: { children: React.ReactNode }) => {
  return (
    <div className="w-fit rounded-md border border-(--system-gray-6) p-1.5 text-2xs/4 whitespace-nowrap">
      {children}
    </div>
  )
}

export function ProjectsPage() {
  const selectedDomainId = useDomainStore().selectedDomain?.id
  const selectedDomain = useDomainStore().selectedDomain
  const [searchQuery, setSearchQuery] = useState('')
  const { setLatestProjectDomain } = useLatestProjectDomainPairs()
  const projectQuery = useProjects()
  const columnHelper = createColumnHelper<Project>()

  const [showArchived] = useQueryState('showArchived')
  const archiveDialogProps = useProjectArchiveRestoreDialog()

  const projectsForCurrentDomain = useMemo(() => {
    if (projectQuery.data) {
      return projectQuery.data.projects.filter((p) =>
        p.domains.find((pd) => pd.id === selectedDomainId),
      )
    }
  }, [projectQuery.data, selectedDomainId])

  const filteredProjects = useMemo(() => {
    if (projectsForCurrentDomain) {
      return projectsForCurrentDomain
        .filter((p) => p.name.toLowerCase().includes(searchQuery.toLowerCase()))
        .filter((p) =>
          showArchived ? true : p.state !== ProjectState.ARCHIVED,
        )
    }
  }, [projectsForCurrentDomain, searchQuery, showArchived])

  const columns = React.useMemo(() => {
    return [
      columnHelper.accessor('name', {
        maxSize: 500,
        cell: (info) => {
          return (
            <div className="flex flex-col">
              <div className="text-sm/5 font-semibold text-(--system-white)">
                {highlightMatches(info.row.original.name, searchQuery)}
              </div>
              <div className="truncate text-sm/5 whitespace-nowrap text-(--system-gray-5)">
                {highlightMatches(info.row.original.description, searchQuery)}
              </div>
            </div>
          )
        },
        header: () => <div className="text-sm/5">Name</div>,
      }),
      columnHelper.accessor('labels', {
        maxSize: 300,
        cell: (info) => {
          const v = info.getValue()

          if (v?.values && Object.keys(v.values).length > 0) {
            const visibleLabels = Object.entries(v.values).slice(0, 2)
            const remainingLabels = Object.entries(v.values).slice(2)
            return (
              <div className="flex gap-2">
                {visibleLabels.map(([k, v]) => (
                  <div key={`${k}-${v}`}>
                    <Badge aria-hidden={true}>
                      {k}:{v}
                    </Badge>
                  </div>
                ))}
                {remainingLabels.length > 0 && (
                  <Tooltip
                    content={
                      <div className="p-1">
                        {remainingLabels.map(([k, v]) => (
                          <div key={`${k}-${v}`}>
                            {k}:{v}
                          </div>
                        ))}
                      </div>
                    }
                  >
                    <div className="rounded-md border border-(--system-gray-5) p-1.5 text-2xs/4">
                      +{remainingLabels.length}
                    </div>
                  </Tooltip>
                )}
              </div>
            )
          }
          return <div className="text-sm text-(--system-gray-5)">None</div>
        },
        header: () => <div className="text-sm/5">Labels</div>,
      }),
      columnHelper.accessor('id', {
        maxSize: 120,
        enableResizing: false,
        cell: (info) => (
          <div className="truncate text-sm/5 whitespace-nowrap text-(--system-gray-5)">
            {highlightMatches(info.getValue(), searchQuery)}
          </div>
        ),
        header: () => <div className="text-sm/5">ID</div>,
      }),
      columnHelper.accessor('state', {
        maxSize: 50,
        header: '',
        cell: (info) => (
          <div
            className="flex items-center justify-end text-(--system-gray-5)"
            onClick={(e) => {
              e.preventDefault()
              e.stopPropagation()
            }}
          >
            <ProjectsItemDropdown
              project={info.row.original}
              setActiveProject={archiveDialogProps.setActiveProject}
            />
          </div>
        ),
      }),
    ]
  }, [columnHelper, searchQuery, archiveDialogProps.setActiveProject])

  return (
    <>
      <div className="flex items-center justify-between gap-2 px-10 pt-6 pb-5">
        <div className="flex flex-col">
          <h1 className="text-xl font-medium">Projects</h1>
          <span className="text-2xs font-semibold">
            {filteredProjects?.length ?? 0} total
          </span>
        </div>

        <SearchBar
          placeholder="Search projects"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          // to prevent the browser from autofilling the input
          autoComplete="off"
        />
      </div>

      <div className="flex items-center justify-between gap-2 px-10 pb-6">
        <div className="flex-1" aria-hidden="true"></div>
        <div className="flex items-center gap-2">
          <CreateProjectWorkflow />
          <ProjectsHeaderDropdown />
        </div>
      </div>

      <ProjectArchiveRestoreDialog {...archiveDialogProps} />
      <div className="bg-primary flex h-full min-h-0 flex-grow flex-col gap-3">
        <TableState
          dataLabel="projects"
          data={filteredProjects}
          searchQuery={searchQuery}
          isError={projectQuery.isError}
          isLoading={projectQuery.isLoading}
          subtitle="Create a project from the CLI"
          // TODO
          // subtitle="Create a project to see it here"
          // content={ CreateProjectButton }
        >
          {(data) => (
            <Table
              data={data}
              columns={columns}
              getRowHref={(runRow) => {
                return `/domain/${selectedDomainId}/project/${runRow.id}/runs`
              }}
              onRowClick={(project) => {
                if (selectedDomain) {
                  setLatestProjectDomain({ domain: selectedDomain, project })
                }
              }}
            />
          )}
        </TableState>
      </div>
    </>
  )
}
