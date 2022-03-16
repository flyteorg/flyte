import { StaticContext } from 'react-router';
import useReactRouter from 'use-react-router';

export interface LocationState {
  backLink?: string;
}

/** A hook for fetching potential state values from the `location` object for the
 * current <Route />. `LocationState` defines the values that can exist.
 */
export function useLocationState(): LocationState {
  const { state = {} } = useReactRouter<{}, StaticContext, LocationState>().location;
  return state;
}
