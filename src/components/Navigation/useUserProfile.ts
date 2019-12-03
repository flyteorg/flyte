import { useFetchableData } from 'components/hooks';
import { UserProfile } from 'models';

// TODO
async function fetchUserProfile() {
    return Promise.resolve(null);
}

export function useUserProfile() {
    return useFetchableData<UserProfile | null>({
        debugName: 'UserProfile',
        defaultValue: null,
        doFetch: fetchUserProfile,
        useCache: true
    });
}
