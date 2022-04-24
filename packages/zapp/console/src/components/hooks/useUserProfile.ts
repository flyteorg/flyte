import { useAPIContext } from 'components/data/apiContext';
import { UserProfile } from 'models/Common/types';
import { useFetchableData } from './useFetchableData';

/** State hook that returns the user information if logged in, null otherwise */
export function useUserProfile() {
  const { getUserProfile } = useAPIContext();
  return useFetchableData<UserProfile | null>({
    debugName: 'UserProfile',
    defaultValue: null,
    doFetch: getUserProfile,
    useCache: true,
  });
}
