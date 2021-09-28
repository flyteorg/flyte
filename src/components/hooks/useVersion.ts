import { useAPIContext } from 'components/data/apiContext';
import { GetVersionResponse } from 'models/Common/types';
import { useFetchableData } from './useFetchableData';

/** State hook that returns the version information */
export function useVersion() {
    const { getVersion } = useAPIContext();
    return useFetchableData<GetVersionResponse | null>({
        debugName: 'Version',
        defaultValue: null,
        doFetch: getVersion,
        useCache: true
    });
}
