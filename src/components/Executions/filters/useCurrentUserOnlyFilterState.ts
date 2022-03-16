import { FilterOperationName } from 'models/AdminEntity/types';
import { UserProfile } from 'models/Common/types';
import { useState } from 'react';
import { useUserProfile } from 'components/hooks/useUserProfile';
import { BooleanFilterState } from './types';
import { useFilterButtonState } from './useFilterButtonState';

function getUserId(profile: UserProfile): string {
  return profile.sub ? profile.sub : '';
}

/** Maintains the state for a button to be used for filtering by user.
 */
export function useCurrentUserOnlyFilterState(): BooleanFilterState {
  const profile = useUserProfile();
  const userId = profile.value ? getUserId(profile.value) : '';
  const [active, setActive] = useState(true);

  const button = useFilterButtonState();

  const getFilter = () => {
    return active && userId
      ? [
          {
            value: userId,
            key: 'user',
            operation: FilterOperationName.EQ,
          },
        ]
      : [];
  };

  const onReset = () => setActive(true);

  return {
    active,
    button,
    hidden: !userId,
    label: 'Only my executions',
    getFilter,
    setActive,
    onReset,
    type: 'boolean',
    status: profile.state.value.toString(),
  };
}
