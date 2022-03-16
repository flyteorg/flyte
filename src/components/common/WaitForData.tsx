import { log } from 'common/log';
import { isFunction } from 'common/typeCheckers';
import { DataError } from 'components/Errors/DataError';
import { FetchableState, fetchStates } from 'components/hooks/types';
import * as React from 'react';
import { ErrorBoundary } from './ErrorBoundary';
import { LoadingSpinner } from './LoadingSpinner';

const defaultErrorTitle = 'Failed to fetch data';

type SpinnerVariant = 'small' | 'medium' | 'large' | 'none';

function renderSpinner(variant: SpinnerVariant) {
  if (variant === 'none') {
    return null;
  }
  return <LoadingSpinner size={variant} />;
}

interface WaitForDataProps {
  children: (() => React.ReactNode) | React.ReactNode;
  /** Component to use for displaying errors. This will override `errorTitle` */
  errorComponent?: React.ComponentType<{ error?: Error; retry?(): any }>;
  /** The string to display as the header of the error content */
  errorTitle?: string;
  /** Any error returned from the last fetch attempt, will trigger an error visual */
  lastError?: Error | null;
  /** An optional custom component to render when in the loading state.
   * Setting this will override `spinnerVariant`
   */
  loadingComponent?: React.ComponentType;
  /** Which size spinner to use, defaults to 'large'. Pass 'none' to disable
   * the loading animation entirely.
   */
  spinnerVariant?: SpinnerVariant;
  /** Loading state (passed from Fetchable) */
  state: FetchableState<any>;
  /** A callback that will initiaite a fetch of the underlying resource. This
   * is wired to a "Retry" button when showing the error visual.
   */
  fetch?(): any;
}

/** A wrapper component which will change its rendering behavior based on
 * loading state. If there is an error on the first fetch, it will show an error
 * component. Otherwise, it will render its children or a loading spinner based
 * on the value of `state`. The loading state props for this component are
 * usually obtained from a `Fetchable` object.
 */
export const WaitForData: React.FC<WaitForDataProps> = ({
  children,
  errorComponent: ErrorComponent,
  errorTitle = defaultErrorTitle,
  lastError,
  loadingComponent: LoadingComponent,
  spinnerVariant = 'large',
  state,
  fetch,
}) => {
  switch (true) {
    case state.matches(fetchStates.IDLE): {
      return null;
    }
    case state.matches(fetchStates.LOADED):
    case state.matches(fetchStates.REFRESH_FAILED):
    case state.matches(fetchStates.REFRESHING):
    case state.matches(fetchStates.REFRESH_FAILED_RETRYING): {
      return (
        <ErrorBoundary>
          <>{isFunction(children) ? children() : children}</>
        </ErrorBoundary>
      );
    }
    // An error occurred and we haven't successfully fetched a value yet,
    // so the only option is to show the error
    case state.matches(fetchStates.FAILED): {
      const error = lastError || new Error('Unknown failure');
      return ErrorComponent ? (
        <ErrorComponent error={error} retry={fetch} />
      ) : (
        <DataError error={error} errorTitle={errorTitle} retry={fetch} />
      );
    }
    case state.matches(fetchStates.LOADING):
    case state.matches(fetchStates.FAILED_RETRYING): {
      return LoadingComponent ? <LoadingComponent /> : renderSpinner(spinnerVariant);
    }
    default:
      log.error(`Unexpected fetch state value: ${state.value}`);
      return null;
  }
};
