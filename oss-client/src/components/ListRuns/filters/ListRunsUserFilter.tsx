import { useEffect, useMemo, useState } from 'react'

import { PopoverMenu, type MenuItem } from '@/components/Popovers'
import { usersPayload } from './usersPayload'
import { ListUsersResponse } from '@/gen/identity/user_payload_pb'
import { User } from '@/gen/common/identity_pb'
import { useSelectedUsers } from '@/hooks/useQueryParamState'

async function getUsers() {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve(usersPayload)
    }, 500)
  })
}

export const ListRunsUserFilter = () => {
  const [userState, setUserstate] = useState<User[]>([])

  const { selectedUsers, toggleUser } = useSelectedUsers()

  useEffect(() => {
    const getData = async () => {
      const response = await getUsers()
      const typedResponse = response as ListUsersResponse
      if (typedResponse.users) {
        setUserstate(typedResponse.users)
      }
    }
    getData()
  }, [])

  const formattedMenuItems: MenuItem[] = useMemo(() => {
    return userState.map(
      (u): MenuItem => ({
        id: `${u.spec?.firstName} ${u.spec?.lastName}`,
        type: 'item',
        label: `${u.spec?.firstName} ${u.spec?.lastName}`,
        onClick: () => {
          if (u.id?.subject) {
            toggleUser(u.id?.subject)
          }
        },
        selected: !!(u.id?.subject && selectedUsers?.includes(u.id?.subject)),
      }),
    )
  }, [userState, selectedUsers, toggleUser])

  return (
    <PopoverMenu
      items={formattedMenuItems}
      label="Owner"
      size="sm"
      outline={!selectedUsers?.length}
    ></PopoverMenu>
  )
}
