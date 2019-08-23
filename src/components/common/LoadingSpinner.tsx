import * as classnames from 'classnames';
import * as React from 'react';

import { CircularProgress } from '@material-ui/core';

import { makeStyles } from '@material-ui/core/styles';

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

/** Renders a loading spinner. Size options are 'small', 'medium', and 'large' */
export const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
    size = 'large'
}) => {
    const styles = useStyles();
    return (
        <div className={classnames(styles.container, size)}>
            <CircularProgress size={spinnerSizes[size]} />
        </div>
    );
};
