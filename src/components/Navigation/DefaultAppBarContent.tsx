import { makeStyles, Theme } from '@material-ui/core/styles';
import classnames from 'classnames';
import { FlyteLogo } from 'components/common/FlyteLogo';
import { useCommonStyles } from 'components/common/styles';
import * as React from 'react';
import { Link } from 'react-router-dom';
import { Routes } from 'routes/routes';
import { UserInformation } from './UserInformation';

const useStyles = makeStyles((theme: Theme) => ({
  spacer: {
    flexGrow: 1,
  },
  rightNavBarItem: {
    marginLeft: theme.spacing(2),
  },
}));

/** Renders the default content for the app bar, which is the logo and help links */
export const DefaultAppBarContent: React.FC = () => {
  const commonStyles = useCommonStyles();
  const styles = useStyles();
  return (
    <>
      <Link className={classnames(commonStyles.linkUnstyled)} to={Routes.SelectProject.path}>
        <FlyteLogo size={32} variant="dark" />
      </Link>
      <div className={styles.spacer} />
      <UserInformation />
    </>
  );
};
