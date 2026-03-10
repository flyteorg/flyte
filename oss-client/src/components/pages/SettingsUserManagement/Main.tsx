'use client'

import { Button } from '@/components/Button'
import { LogsIcon } from '@/components/icons/LogsIcon'
import { RobotIcon } from '@/components/icons/RobotIcon'
import { NavPanelLayout } from '@/components/NavPanel/NavPanelLayout'
import { Tabs } from '@/components/Tabs'
import { useSelectedTab } from '@/hooks/useQueryParamState'
import { ShieldCheckIcon } from '@heroicons/react/24/outline'
import { Suspense, useMemo } from 'react'
import { DisabledButtonWithTooltip } from './DisabledButtonWithTooltip'
import { LicensedEditionPlaceholder } from './LicensedEditionPlaceholder'

const UserManagementTabs = {
  USERS: 'users',
  APPS: 'apps',
  ROLES: 'roles',
  POLICIES: 'policies',
} as const

export function SettingsUserManagementPage() {
  const { selectedTab, setSelectedTab } = useSelectedTab(
    UserManagementTabs.USERS,
    UserManagementTabs,
  )

  const tabs = useMemo(
    () => [
      {
        label: 'Users',
        icon: <LogsIcon width="16" height="16" />,
        path: UserManagementTabs.USERS,
        content: (
          <div className="mx-10 mt-2">
            <LicensedEditionPlaceholder title="Users" />
          </div>
        ),
      },
      {
        label: 'OAuth Apps',
        icon: <RobotIcon width="16" height="16" />,
        path: UserManagementTabs.APPS,
        content: (
          <div className="mx-10 mt-2">
            <LicensedEditionPlaceholder title="OAuth Apps" />
          </div>
        ),
      },
      {
        label: 'Roles',
        icon: <ShieldCheckIcon className="size-4" />,
        path: UserManagementTabs.ROLES,
        content: (
          <div className="mx-10 mt-2">
            <LicensedEditionPlaceholder title="Roles" />
          </div>
        ),
      },
      {
        label: 'Policies',
        icon: <ShieldCheckIcon className="size-4" />,
        path: UserManagementTabs.POLICIES,
        content: (
          <div className="mx-10 mt-2">
            <LicensedEditionPlaceholder title="Policies" />
          </div>
        ),
      },
    ],
    [],
  )

  const widgets = useMemo(
    () =>
      selectedTab === UserManagementTabs.USERS
        ? [
            {
              key: 'create-user',
              content: (
                <DisabledButtonWithTooltip>
                  <Button disabled color="union" size="sm">
                    Create user
                  </Button>
                </DisabledButtonWithTooltip>
              ),
            },
          ]
        : selectedTab === UserManagementTabs.POLICIES
          ? [
              {
                key: 'create-policy',
                content: (
                  <DisabledButtonWithTooltip>
                    <Button disabled color="union" size="sm">
                      Create policy
                    </Button>
                  </DisabledButtonWithTooltip>
                ),
              },
            ]
          : selectedTab === UserManagementTabs.ROLES
            ? [
                {
                  key: 'create-role',
                  content: (
                    <DisabledButtonWithTooltip>
                      <Button disabled color="union" size="sm">
                        Create role
                      </Button>
                    </DisabledButtonWithTooltip>
                  ),
                },
              ]
            : selectedTab === UserManagementTabs.APPS
              ? [
                  {
                    key: 'create-oauth-app',
                    content: (
                      <DisabledButtonWithTooltip>
                        <Button disabled color="union" size="sm">
                          Create OAuth app
                        </Button>
                      </DisabledButtonWithTooltip>
                    ),
                  },
                ]
              : undefined,
    [selectedTab],
  )

  return (
    <Suspense>
      <NavPanelLayout initialSize="wide" mode="embedded" type="settings">
        <main className="flex h-full w-full flex-col bg-(--system-gray-1)">
          <div className="flex h-full min-h-0 flex-col">
            <div className="px-10 pt-10 pb-5">
              <h1 className="text-xl font-medium">User Management</h1>
            </div>

            <div className="flex min-h-0 flex-1 flex-col pb-6">
              <Tabs
                currentTab={selectedTab}
                tabs={tabs}
                onClickTab={setSelectedTab}
                classes={{
                  tabListContainer: '!px-10',
                }}
                widgets={widgets}
              />
            </div>
          </div>
        </main>
      </NavPanelLayout>
    </Suspense>
  )
}

export default SettingsUserManagementPage
