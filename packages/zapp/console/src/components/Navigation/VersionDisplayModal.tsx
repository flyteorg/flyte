import { Dialog, DialogContent } from '@material-ui/core';
import Link from '@material-ui/core/Link';
import { makeStyles, Theme } from '@material-ui/core/styles';
import * as React from 'react';
import { ClosableDialogTitle } from 'components/common/ClosableDialogTitle';
import { useAdminVersion } from 'components/hooks/useVersion';
import { env } from 'common/env';

const { version: platformVersion } = require('../../../package.json');

const useStyles = makeStyles((theme: Theme) => ({
  content: {
    paddingTop: theme.spacing(2),
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
  },
  dialog: {
    maxWidth: `calc(100% - ${theme.spacing(12)}px)`,
    maxHeight: `calc(100% - ${theme.spacing(12)}px)`,
    height: theme.spacing(34),
    width: theme.spacing(36),
  },
  versionWrapper: {
    minWidth: theme.spacing(20),
    display: 'flex',
    flexDirection: 'row',
    justifyContent: 'space-between',
    fontFamily: 'Apple SD Gothic Neo',
    fontWeight: 'normal',
    fontSize: '14px',
    lineHeight: '17px',
    color: '#636379',
    marginBottom: '6px',
  },
  title: {
    fontFamily: 'Open Sans',
    fontWeight: 'bold',
    fontSize: '16px',
    lineHeight: '22px',
    margin: '7px 0 26px 0',
    color: '#000',
  },
  link: {
    fontFamily: 'Apple SD Gothic Neo',
    fontWeight: 'normal',
    fontSize: '14px',
    lineHeight: '17px',
    color: '#636379',
    margin: '17px 0 2px 0',
  },
  version: {
    color: '#1982E3',
    fontSize: '14px',
  },
}));

interface VersionDisplayModalProps {
  onClose(): void;
}

export const VersionDisplayModal: React.FC<VersionDisplayModalProps> = ({ onClose }) => {
  const styles = useStyles();
  const { adminVersion } = useAdminVersion();

  const { DISABLE_GA } = env;

  return (
    <Dialog
      PaperProps={{ className: styles.dialog }}
      maxWidth={false}
      open={true}
      onClose={onClose}
    >
      <ClosableDialogTitle onClose={onClose}>{}</ClosableDialogTitle>
      <DialogContent className={styles.content}>
        <svg height={32} viewBox="100 162 200 175">
          <path
            fill="#40119C"
            d="M152.24 254.36S106 214.31 106 174.27c0 0 6.79-6.89 33.93-8.55l-.31 8.55c0 40.04 12.62 80.09 12.62 80.09zm46.24-80.09s-6.8-6.89-33.92-8.55l.31 8.55c0 40-12.63 80.09-12.63 80.09s46.24-40.05 46.24-80.09zm-16.69 131.07l-7.55 4c15 22.68 24.36 25.11 24.36 25.11 34.68-20 46.24-80.09 46.24-80.09s-28.37 30.96-63.05 50.98zm-29.43-51a11 11 0 00-.23 2.62c0 4.25 1.27 13.94 9.79 31l7.24-4.54c34.69-20 75.68-29.11 75.68-29.11s-25.69-8.9-53.09-8.9c-13.7 0-27.83 2.23-39.39 8.9zm121.69-50.8l7.24 4.54c8.53-17.1 9.79-26.78 9.79-31a11.34 11.34 0 00-.22-2.62c-11.57-6.68-25.69-8.9-39.4-8.9-27.4 0-53.09 8.89-53.09 8.89s41 9.08 75.68 29.11zm-29.44 51c.85-.22 10.08-3.52 24.37-25.11l-7.56-4c-34.68-20-63.05-51-63.05-51s11.56 60.08 46.24 80.1z"
          />
        </svg>
        <div className={styles.title}>Flyte Console</div>
        <div className={styles.versionWrapper}>
          <span>UI Version</span>
          <Link
            href={`https://github.com/flyteorg/flyteconsole/releases/tag/v${platformVersion}`}
            className={styles.version}
            target="_blank"
          >
            {platformVersion}
          </Link>
        </div>
        <div className={styles.versionWrapper}>
          <span>Admin Version</span>
          <Link
            href={`https://github.com/flyteorg/flyteadmin/releases/tag/v${adminVersion}`}
            className={styles.version}
            target="_blank"
          >
            {adminVersion}
          </Link>
        </div>
        <div className={styles.versionWrapper}>
          <span>Google Analytics</span>
          <Link
            href="https://github.com/flyteorg/flyteconsole#google-analytics"
            className={styles.version}
            target="_blank"
          >
            {DISABLE_GA === 'false' ? 'Active' : 'Inactive'}
          </Link>
        </div>
        <div className={styles.link}>Documentation Link:</div>
        <Link
          href="https://docs.flyte.org/en/latest/"
          style={{
            color: '#1982E3',
            fontSize: '13px',
          }}
          target="_blank"
        >
          https://docs.flyte.org/en/latest/
        </Link>
      </DialogContent>
    </Dialog>
  );
};
