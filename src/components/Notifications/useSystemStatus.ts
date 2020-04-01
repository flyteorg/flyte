import { useAPIContext } from 'components/data/apiContext';
import { useFetchableData } from 'components/hooks';
import { defaultSystemStatus } from 'models/Common/constants';

/** Hook for fetching the current system status. Defaults to a safe value
 * indicating normal system status.
 */
export function useSystemStatus() {
    const { getSystemStatus } = useAPIContext();
    return useFetchableData({
        defaultValue: defaultSystemStatus,
        doFetch: getSystemStatus
    });
}
