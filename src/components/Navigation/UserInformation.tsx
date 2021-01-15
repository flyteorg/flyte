import Link from '@material-ui/core/Link';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { WaitForData } from 'components/common/WaitForData';
import { useUserProfile } from 'components/hooks/useUserProfile';
import { getLoginUrl } from 'models/AdminEntity/utils';
import * as React from 'react';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        color: theme.palette.common.white
    }
}));

const LoginLink: React.FC = () => (
    <Link href={getLoginUrl()} color="inherit">
        Login
    </Link>
);

/** Displays user info if logged in, or a login link otherwise. */
export const UserInformation: React.FC<{}> = () => {
    const profile = useUserProfile();
    return (
        <WaitForData spinnerVariant="none" {...profile}>
            <div className={useStyles().container}>
                {profile.value == null ? (
                    <LoginLink />
                ) : (
                    profile.value.preferredUsername
                )}
            </div>
        </WaitForData>
    );
};
