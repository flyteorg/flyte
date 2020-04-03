import Paper from '@material-ui/core/Paper';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { WaitForData } from 'components/common';
import { SystemStatus } from 'models/Common/types';
import * as React from 'react';
import { useSystemStatus } from './useSystemStatus';

const useStyles = makeStyles((theme: Theme) => ({
    statusPaper: {
        bottom: theme.spacing(3),
        margin: theme.spacing(3),
        position: 'absolute'
    },
    statusContentContainer: {
        display: 'flex'
    },
    statusClose: {
        flex: '0 0 auto'
    },
    statusIcon: {
        flex: '0 0 auto'
    },
    statusMessage: {
        flex: '1 1 auto'
    }
}));

const RenderSystemStatusBanner: React.FC<{
    status: SystemStatus;
    onClose: () => void;
}> = ({ status, onClose }) => {
    const styles = useStyles();
    return (
        <Paper className={styles.statusPaper} elevation={2} square={true}>
            <div className={styles.statusContentContainer}>
                <div className={styles.statusMessage}>{status.message}</div>
            </div>
        </Paper>
    );
};

export const SystemStatusBanner: React.FC<{}> = () => {
    const systemStatus = useSystemStatus();
    const [dismissed, setDismissed] = React.useState(false);
    const onClose = () => setDismissed(true);
    if (dismissed) {
        return null;
    }
    return (
        <WaitForData {...systemStatus}>
            {systemStatus.value.message && (
                <RenderSystemStatusBanner
                    status={systemStatus.value}
                    onClose={onClose}
                />
            )}
        </WaitForData>
    );
};
