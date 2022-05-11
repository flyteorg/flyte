import * as React from 'react';
import classnames from 'classnames';
import { Link, makeStyles, Theme } from '@material-ui/core';
import { FlyteLogo } from '@flyteconsole/ui-atoms';
import t from './strings';

const headerFontFamily = '"Open Sans", helvetica, arial, sans-serif';

/* eslint-disable no-dupe-keys */
const useStyles = makeStyles((theme: Theme) => ({
  title: {
    fontFamily: headerFontFamily,
    fontWeight: 'bold',
    fontSize: '16px',
    lineHeight: '22px',
    margin: theme.spacing(1, 0),
    color: '#000',
  },
  versionsContainer: {
    width: '100%',
    margin: theme.spacing(3),
    padding: theme.spacing(0, 1),
  },
  versionWrapper: {
    display: 'flex',
    justifyContent: 'space-between',
    marginBottom: theme.spacing(1),
  },
  versionName: {
    fontFamily: 'Apple SD Gothic Neo',
    fontWeight: 'normal',
    fontSize: '14px',
    lineHeight: '17px',
    color: '#636379',
  },
  versionLink: {
    color: '#1982E3',
    fontSize: '14px',
    marginLeft: theme.spacing(1),
  },
}));

export type VersionInfo = {
  name: string;
  version: string;
  url: string;
};

export interface VersionDisplayProps {
  versions: VersionInfo[];
  documentationUrl: string;
}

export const VersionDisplay = (props: VersionDisplayProps): JSX.Element => {
  const styles = useStyles();

  const VersionItem = (info: VersionInfo) => {
    return (
      <div
        className={classnames(styles.versionName, styles.versionWrapper)}
        key={`${info.name} + ${info.version}`}
      >
        <span>{info.name}</span>
        <Link href={info.url} className={styles.versionLink} target="_blank">
          {info.version}
        </Link>
      </div>
    );
  };

  const versionsList = props.versions.map((info) => VersionItem(info));

  return (
    <>
      <div style={{ height: `32px` }}>
        <FlyteLogo size={32} background="light" hideText={true} />
      </div>
      <div className={styles.title}>{t('modalTitle')}</div>

      <div className={styles.versionsContainer}>{versionsList}</div>

      <div className={styles.versionName}>{t('docsLink')}</div>
      <Link href={props.documentationUrl} className={styles.versionLink} target="_blank">
        {props.documentationUrl}
      </Link>
    </>
  );
};
