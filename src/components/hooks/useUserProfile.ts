import { useFetchableData } from 'components/hooks';
import { getUserProfile, UserProfile } from 'models';

export function useUserProfile() {
    return useFetchableData<UserProfile | null>({
        debugName: 'UserProfile',
        defaultValue: null,
        doFetch: getUserProfile,
        useCache: true
    });
}
