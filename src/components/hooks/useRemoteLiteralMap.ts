import { useAPIContext } from 'components/data/apiContext';
import { LiteralMap } from 'models/Common/types';
import { FetchableData } from './types';
import { useFetchableData } from './useFetchableData';

/** Fetches a LiteralMap from a given URL */
export function useRemoteLiteralMap(url: string): FetchableData<LiteralMap> {
  const { getRemoteLiteralMap } = useAPIContext();
  return useFetchableData<LiteralMap, string>(
    {
      debugName: 'RemoteLiteralMap',
      defaultValue: {} as LiteralMap,
      doFetch: (blobUrl) => getRemoteLiteralMap(blobUrl),
    },
    url,
  );
}
