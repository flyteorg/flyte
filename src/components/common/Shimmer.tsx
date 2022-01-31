import * as React from 'react';
import { makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(() => ({
    animate: {
        height: 10,
        animation: '$shimmer 4s infinite linear',
        background:
            'linear-gradient(to right, #eff1f3 4%, #e2e2e2 25%, #eff1f3 36%)',
        backgroundSize: '1000px 100%',
        borderRadius: 8,
        width: '100%'
    },
    '@keyframes shimmer': {
        '0%': {
            backgroundPosition: '-1000px 0'
        },
        '100%': {
            backgroundPosition: '1000px 0'
        }
    }
}));

interface ShimmerProps {
    height?: number;
}

export const Shimmer: React.FC<ShimmerProps> = () => {
    const styles = useStyles();

    return <div className={styles.animate}></div>;
};
