import { useAPIContext } from 'components/data/apiContext';
import { GetVersionResponse } from 'models/Common/types';
import { useFetchableData } from './useFetchableData';

/** State hook that returns the version information */
function useVersion() {
  const { getVersion } = useAPIContext();
  return useFetchableData<GetVersionResponse | null>({
    debugName: 'Version',
    defaultValue: null,
    doFetch: getVersion,
    useCache: true,
  });
}

export function useAdminVersion() {
  const version = useVersion();

  const controlPlaneVersion = version?.value?.controlPlaneVersion;
  const adminVersion = controlPlaneVersion
    ? controlPlaneVersion.Version?.slice(1, controlPlaneVersion.Version?.indexOf('-'))
    : null;

  return { adminVersion };
}
