import { useFetchableData } from 'components/hooks';
import { getUserProfile, UserProfile } from 'models';

/** State hook that returns the user information if logged in, null otherwise */
export function useUserProfile() {
    return useFetchableData<UserProfile | null>({
        debugName: 'UserProfile',
        defaultValue: null,
        doFetch: getUserProfile,
        useCache: true
    });
}
