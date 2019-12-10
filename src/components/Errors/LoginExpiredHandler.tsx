import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle
} from '@material-ui/core';
import { useAPIContext } from 'components/data/apiContext';
import { getLoginUrl } from 'models';
import * as React from 'react';

const title = 'Your login session has expired';
const description = 'Please re-authenticate to continue.';

/** Detects expired login state via context and displays a Dialog directing
 * the user to re-authenticate.
 */
export const LoginExpiredHandler: React.FC = () => {
    const { loginStatus } = useAPIContext();
    return (
        <Dialog
            open={loginStatus.expired}
            aria-describedby="login-expired-description"
        >
            <DialogTitle id="login-expired-title">{title}</DialogTitle>
            <DialogContent>
                <DialogContentText id="login-expired-description">
                    {description}
                </DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button href={getLoginUrl()} color="primary">
                    Login
                </Button>
            </DialogActions>
        </Dialog>
    );
};
