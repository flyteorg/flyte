import * as React from 'react';
import { useFlyteApi } from '@flyteconsole/flyte-api';
import { Link, makeStyles, Theme } from '@material-ui/core';
import { WaitForData } from 'components/common/WaitForData';
import { useUserProfile } from 'components/hooks/useUserProfile';
import t from './strings';

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    color: theme.palette.common.white,
  },
}));

const LoginLink = (props: { loginUrl: string }) => {
  return (
    <Link href={props.loginUrl} color="inherit">
      {t('login')}
    </Link>
  );
};

/** Displays user info if logged in, or a login link otherwise. */
export const UserInformation: React.FC<{}> = () => {
  const style = useStyles();
  const profile = useUserProfile();
  const apiContext = useFlyteApi();

  return (
    <WaitForData spinnerVariant="none" {...profile}>
      <div className={style.container}>
        {!profile.value ? (
          <LoginLink loginUrl={apiContext.getLoginUrl()} />
        ) : !profile.value.preferredUsername || profile.value.preferredUsername === '' ? (
          profile.value.name
        ) : (
          profile.value.preferredUsername
        )}
      </div>
    </WaitForData>
  );
};
