import { useFlyteApi, AdminEndpoint, getAxiosApiCall } from '@flyteconsole/flyte-api';
import { GetVersionResponse } from 'models/Common/types';
import { useFetchableData } from './useFetchableData';

/** State hook that returns the version information */
function useVersion() {
  const { getAdminApiUrl } = useFlyteApi();
  const versionPath = getAdminApiUrl(AdminEndpoint.Version);

  return useFetchableData<GetVersionResponse | null>({
    debugName: 'Version',
    defaultValue: null,
    doFetch: () => getAxiosApiCall(versionPath),
    useCache: true,
  });
}

export function useAdminVersion() {
  const version = useVersion();

  const cpVersion = version?.value?.controlPlaneVersion;

  // Remove letter "v" from version string, if it's in the beginning
  const adminVersion = cpVersion?.Version?.replace(/^v/, '') ?? null;

  return { adminVersion };
}
