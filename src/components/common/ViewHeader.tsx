import Typography from '@material-ui/core/Typography';
import * as React from 'react';

import { makeStyles, Theme } from '@material-ui/core/styles';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        marginBottom: theme.spacing(1)
    },
    title: {
        marginBottom: theme.spacing(2)
    }
}));

export interface ViewHeaderProps {
    title: string;
    subtitle?: string;
}
export const ViewHeader: React.FC<ViewHeaderProps> = ({ title, subtitle }) => {
    const styles = useStyles();
    return (
        <header className={styles.container}>
            <Typography variant="h5" className={styles.title}>
                {title}
            </Typography>
            {!!subtitle && <Typography variant="body1">{subtitle}</Typography>}
        </header>
    );
};
