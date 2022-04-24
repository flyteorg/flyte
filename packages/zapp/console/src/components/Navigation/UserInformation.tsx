import Link from '@material-ui/core/Link';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { WaitForData } from 'components/common/WaitForData';
import { useUserProfile } from 'components/hooks/useUserProfile';
import { getLoginUrl } from 'models/AdminEntity/utils';
import * as React from 'react';
import { VersionDisplayModal } from './VersionDisplayModal';

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    color: theme.palette.common.white,
  },
}));

const LoginLink: React.FC = () => (
  <Link href={getLoginUrl()} color="inherit">
    Login
  </Link>
);

/** Displays user info if logged in, or a login link otherwise. */
export const UserInformation: React.FC<{}> = () => {
  const profile = useUserProfile();
  const [showVersionInfo, setShowVersionInfo] = React.useState(false);

  let modalContent: JSX.Element | null = null;
  if (showVersionInfo) {
    const onClose = () => setShowVersionInfo(false);
    modalContent = <VersionDisplayModal onClose={onClose} />;
  }
  return (
    <WaitForData spinnerVariant="none" {...profile}>
      <div className={useStyles().container}>
        {!profile.value ? (
          <LoginLink />
        ) : !profile.value.preferredUsername || profile.value.preferredUsername === '' ? (
          profile.value.name
        ) : (
          profile.value.preferredUsername
        )}
      </div>
      <div
        onClick={() => setShowVersionInfo(true)}
        style={{
          cursor: 'pointer',
          display: 'flex',
          marginLeft: '5px',
        }}
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="20"
          height="20"
          fill="currentColor"
          className="bi bi-info-circle"
          viewBox="0 0 16 16"
        >
          <path d="M8 15A7 7 0 1 1 8 1a7 7 0 0 1 0 14zm0 1A8 8 0 1 0 8 0a8 8 0 0 0 0 16z" />
          <path d="m8.93 6.588-2.29.287-.082.38.45.083c.294.07.352.176.288.469l-.738 3.468c-.194.897.105 1.319.808 1.319.545 0 1.178-.252 1.465-.598l.088-.416c-.2.176-.492.246-.686.246-.275 0-.375-.193-.304-.533L8.93 6.588zM9 4.5a1 1 0 1 1-2 0 1 1 0 0 1 2 0z" />
        </svg>
      </div>
      {modalContent}
    </WaitForData>
  );
};
