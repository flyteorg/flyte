import { CircularProgress } from '@material-ui/core';
import { makeStyles } from '@material-ui/core/styles';
import classnames from 'classnames';
import { useDelayedValue } from 'components/hooks/useDelayedValue';
import * as React from 'react';
import { loadingSpinnerDelayMs } from './constants';

const useStyles = makeStyles({
    container: {
        alignItems: 'center',
        display: 'flex',
        justifyContent: 'center',
        '&.large': { height: '300px' },
        '&.medium': { height: '80px' },
        '&.small': { height: '40px' }
    }
});

type SizeValue = 'small' | 'medium' | 'large';
interface LoadingSpinnerProps {
    size?: SizeValue;
}

const spinnerSizes: Record<SizeValue, number> = {
    small: 24,
    medium: 48,
    large: 96
};

/** Renders a loading spinner after 1000ms. Size options are 'small', 'medium', and 'large' */
export const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
    size = 'large'
}) => {
    const styles = useStyles();
    const shouldRender = useDelayedValue(false, loadingSpinnerDelayMs, true);
    return shouldRender ? (
        <div
            className={classnames(styles.container, size)}
            data-testid="loading-spinner"
        >
            <CircularProgress size={spinnerSizes[size]} />
        </div>
    ) : null;
};

/** `LoadingSpinner` with a pre-bound size of `small` */
export const SmallLoadingSpinner: React.FC = () => (
    <LoadingSpinner size="small" />
);
/** `LoadingSpinner` with a pre-bound size of `medium` */
export const MediumLoadingSpinner: React.FC = () => (
    <LoadingSpinner size="medium" />
);
/** `LoadingSpinner` with a pre-bound size of `large` */
export const LargeLoadingSpinner: React.FC = () => (
    <LoadingSpinner size="large" />
);
