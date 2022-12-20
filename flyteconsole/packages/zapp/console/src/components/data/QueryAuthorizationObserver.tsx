import { NotAuthorizedError } from 'errors/fetchErrors';
import * as React from 'react';
import { onlineManager, Query, useQueryClient } from 'react-query';
import { useFlyteApi } from '@flyteconsole/flyte-api';

/** Watches all queries to detect a NotAuthorized error, disabling future queries
 * and triggering the login refresh flow.
 * Note: Should be placed just below the QueryClient and ApiContext providers.
 */

// TODO: narusina - move this one to flyte-api too
export const QueryAuthorizationObserver: React.FC = () => {
  const queryCache = useQueryClient().getQueryCache();
  const apiContext = useFlyteApi();
  React.useEffect(() => {
    const unsubscribe = queryCache.subscribe((query?: Query | undefined) => {
      if (!query || !query.state.error) {
        return;
      }
      if (query.state.error instanceof NotAuthorizedError) {
        // Stop all in-progress and future requests
        onlineManager.setOnline(false);
        // Trigger auth flow
        apiContext.loginStatus.setExpired(true);
      }
    });
    return unsubscribe;
  }, [queryCache, apiContext]);
  return null;
};
