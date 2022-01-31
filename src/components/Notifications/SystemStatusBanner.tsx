import { ButtonBase, Typography } from '@material-ui/core';
import Paper from '@material-ui/core/Paper';
import { makeStyles, Theme } from '@material-ui/core/styles';
import Close from '@material-ui/icons/Close';
import Info from '@material-ui/icons/Info';
import Warning from '@material-ui/icons/Warning';
import { Empty } from 'components/common/Empty';
import { LinkifiedText } from 'components/common/LinkifiedText';
import { WaitForData } from 'components/common/WaitForData';
import {
    infoIconColor,
    mutedButtonColor,
    mutedButtonHoverColor,
    warningIconColor
} from 'components/Theme/constants';
import { StatusString, SystemStatus } from 'models/Common/types';
import * as React from 'react';
import { useSystemStatus } from './useSystemStatus';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        bottom: 0,
        display: 'flex',
        justifyContent: 'center',
        left: 0,
        // The parent container extends the full width of the page.
        // We don't want it intercepting pointer events for visible items below it.
        pointerEvents: 'none',
        position: 'fixed',
        padding: theme.spacing(2),
        right: 0
    },
    closeButton: {
        alignItems: 'center',
        color: mutedButtonColor,
        '&:hover': {
            color: mutedButtonHoverColor
        },
        display: 'flex',
        height: theme.spacing(3)
    },
    statusPaper: {
        display: 'flex',
        padding: theme.spacing(2),
        pointerEvents: 'initial'
    },
    statusContentContainer: {
        alignItems: 'flex-start',
        display: 'flex',
        maxWidth: theme.spacing(131)
    },
    statusClose: {
        alignItems: 'flex-start',
        display: 'flex',
        flex: '0 0 auto'
    },
    statusIcon: {
        alignItems: 'center',
        display: 'flex',
        flex: '0 0 auto',
        lineHeight: `${theme.spacing(3)}px`
    },
    statusMessage: {
        flex: '1 1 auto',
        fontWeight: 'normal',
        lineHeight: `${theme.spacing(3)}px`,
        marginLeft: theme.spacing(2),
        marginRight: theme.spacing(2)
    }
}));

const InfoIcon = () => (
    <Info data-testid="info-icon" htmlColor={infoIconColor} />
);

const WarningIcon = () => (
    <Warning data-testid="warning-icon" htmlColor={warningIconColor} />
);

const statusIcons: Record<StatusString, React.ComponentType> = {
    normal: InfoIcon,
    degraded: WarningIcon,
    down: WarningIcon
};

const StatusIcon: React.FC<{ status: StatusString }> = ({ status }) => {
    const IconComponent = statusIcons[status] || statusIcons.normal;
    return <IconComponent />;
};

const RenderSystemStatusBanner: React.FC<{
    systemStatus: SystemStatus;
    onClose: () => void;
}> = ({ systemStatus: { message, status }, onClose }) => {
    const styles = useStyles();
    return (
        <div className={styles.container}>
            <Paper
                role="banner"
                className={styles.statusPaper}
                elevation={1}
                square={true}
            >
                <div className={styles.statusIcon}>
                    <StatusIcon status={status} />
                </div>
                <div className={styles.statusContentContainer}>
                    <Typography variant="h6" className={styles.statusMessage}>
                        <LinkifiedText text={message} />
                    </Typography>
                </div>
                <div className={styles.statusClose}>
                    <ButtonBase
                        aria-label="Close"
                        className={styles.closeButton}
                        onClick={onClose}
                    >
                        <Close fontSize="small" />
                    </ButtonBase>
                </div>
            </Paper>
        </div>
    );
};

/** Fetches and renders the system status returned by issuing a GET to
 * `env.STATUS_URL`. If the status includes a message, a dismissable toast
 * will be rendered. Otherwise, nothing will be rendered.
 */
export const SystemStatusBanner: React.FC<{}> = () => {
    const systemStatus = useSystemStatus();
    const [dismissed, setDismissed] = React.useState(false);
    const onClose = () => setDismissed(true);
    if (dismissed) {
        return null;
    }
    return (
        <WaitForData
            {...systemStatus}
            loadingComponent={Empty}
            errorComponent={Empty}
        >
            {systemStatus.value.message ? (
                <RenderSystemStatusBanner
                    systemStatus={systemStatus.value}
                    onClose={onClose}
                />
            ) : null}
        </WaitForData>
    );
};
