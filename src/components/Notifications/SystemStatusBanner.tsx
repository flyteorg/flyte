import { ButtonBase, SvgIcon, Typography } from '@material-ui/core';
import Paper from '@material-ui/core/Paper';
import { makeStyles, Theme } from '@material-ui/core/styles';
import Close from '@material-ui/icons/Close';
import Info from '@material-ui/icons/Info';
import Warning from '@material-ui/icons/Warning';
import { WaitForData } from 'components/common';
import {
    infoIconColor,
    mutedButtonColor,
    mutedButtonHoverColor,
    warningIconColor
} from 'components/Theme';
import { StatusString, SystemStatus } from 'models/Common/types';
import * as React from 'react';
import { useSystemStatus } from './useSystemStatus';

const useStyles = makeStyles((theme: Theme) => ({
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
        bottom: theme.spacing(2),
        display: 'flex',
        left: '50%',
        padding: theme.spacing(2),
        position: 'fixed',
        transform: 'translateX(-50%)'
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

interface StatusConstantValues {
    color: string;
    IconComponent: typeof SvgIcon;
}

const statusConstants: Record<StatusString, StatusConstantValues> = {
    normal: {
        color: infoIconColor,
        IconComponent: Info
    },
    degraded: {
        color: warningIconColor,
        IconComponent: Warning
    },
    down: {
        color: warningIconColor,
        IconComponent: Warning
    }
};

const StatusIcon: React.FC<{ status: StatusString }> = ({ status }) => {
    const { color, IconComponent } =
        statusConstants[status] || statusConstants.normal;
    return <IconComponent htmlColor={color} />;
};

const RenderSystemStatusBanner: React.FC<{
    systemStatus: SystemStatus;
    onClose: () => void;
}> = ({ systemStatus: { message, status }, onClose }) => {
    const styles = useStyles();
    return (
        <Paper className={styles.statusPaper} elevation={1} square={true}>
            <div className={styles.statusIcon}>
                <StatusIcon status={status} />
            </div>
            <div className={styles.statusContentContainer}>
                <Typography variant="h6" className={styles.statusMessage}>
                    {message}
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
        <WaitForData {...systemStatus}>
            {systemStatus.value.message ? (
                <RenderSystemStatusBanner
                    systemStatus={systemStatus.value}
                    onClose={onClose}
                />
            ) : null}
        </WaitForData>
    );
};
