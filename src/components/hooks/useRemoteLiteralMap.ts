import { useAPIContext } from 'components/data/apiContext';
import { LiteralMap } from 'models';
import { FetchableData } from './types';
import { useFetchableData } from './useFetchableData';

/** Fetches a LiteralMap from a given URL */
export function useRemoteLiteralMap(url: string): FetchableData<LiteralMap> {
    const { getRemoteLiteralMap } = useAPIContext();
    // TODO: caching of these objects (can they change?)
    return useFetchableData<LiteralMap, string>(
        {
            debugName: 'RemoteLiteralMap',
            defaultValue: {} as LiteralMap,
            doFetch: blobUrl => getRemoteLiteralMap(blobUrl)
        },
        url
    );
}
