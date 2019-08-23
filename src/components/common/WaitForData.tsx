import * as React from 'react';

import { isFunction } from 'common/typeCheckers';
import { DataError } from 'components/Errors';
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
    /** The string to display as the header of the error content */
    errorTitle?: string;
    /** Indicates if the data has successfully loaded at least once */
    hasLoaded: boolean;
    /** Any error returned from the last fetch attempt, will trigger an error visual */
    lastError?: Error | null;
    /** True if a fetch request is currently in progress */
    loading: boolean;
    /** An optional custom component to render when in the loading state.
     * Setting this will override `spinnerVariant`
     */
    loadingComponent?: React.ComponentType;
    /** Which size spinner to use, defaults to 'large'. Pass 'none' to disable
     * the loading animation entirely.
     */
    spinnerVariant?: SpinnerVariant;
    /** A callback that will initiaite a fetch of the underlying resource. This
     * is wired to a "Retry" button when showing the error visual.
     */
    fetch?(): any;
}

/** A wrapper component which will change its rendering behavior based on
 * loading state. If there is an error on the first fetch, it will show an error
 * component. Otherwise, it will render its children or a loading spinner based
 * on the value of `hasLoaded`. The loading state props for this component are
 * usually obtained from a `Fetchable` object.
 */
export const WaitForData: React.FC<WaitForDataProps> = ({
    children,
    errorTitle = defaultErrorTitle,
    hasLoaded,
    lastError,
    loading,
    loadingComponent: LoadingComponent,
    spinnerVariant = 'large',
    fetch
}) => {
    // If an error occurs, only show it if we haven't successfully retrieved a
    // value yet
    if (lastError && !hasLoaded) {
        return (
            <DataError
                error={lastError}
                errorTitle={errorTitle}
                retry={fetch}
            />
        );
    }

    if (hasLoaded) {
        return (
            <ErrorBoundary>
                <>{isFunction(children) ? children() : children}</>
            </ErrorBoundary>
        );
    }

    if (!loading) {
        return null;
    }

    return LoadingComponent ? (
        <LoadingComponent />
    ) : (
        renderSpinner(spinnerVariant)
    );
};
