import { useAPIContext } from 'components/data/apiContext';
import { useFetchableData } from 'components/hooks';
import { UserProfile } from 'models';

/** State hook that returns the user information if logged in, null otherwise */
export function useUserProfile() {
    const { getUserProfile } = useAPIContext();
    return useFetchableData<UserProfile | null>({
        debugName: 'UserProfile',
        defaultValue: null,
        doFetch: getUserProfile,
        useCache: true
    });
}
