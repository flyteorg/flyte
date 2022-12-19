import { makeStyles, Theme } from '@material-ui/core/styles';
import classnames from 'classnames';
import { AppInfo, VersionInfo } from '@flyteconsole/components';
import { FlyteLogo } from '@flyteconsole/ui-atoms';
import { useCommonStyles } from 'components/common/styles';
import * as React from 'react';
import { Link } from 'react-router-dom';
import { Routes } from 'routes/routes';
import { FeatureFlag, useFeatureFlag } from 'basics/FeatureFlags';
import { useAdminVersion } from 'components/hooks/useVersion';
import { env } from 'common/env';
import { NavigationDropdown } from './NavigationDropdown';
import { UserInformation } from './UserInformation';
import { OnlyMine } from './OnlyMine';
import { FlyteNavItem } from './utils';
import t, { patternKey } from './strings';

const { version: platformVersion } = require('../../../package.json');

const useStyles = makeStyles((theme: Theme) => ({
  spacer: {
    flexGrow: 1,
  },
  rightNavBarItem: {
    marginLeft: theme.spacing(2),
  },
}));

interface DefaultAppBarProps {
  items: FlyteNavItem[];
  console?: string;
}

/** Renders the default content for the app bar, which is the logo and help links */
export const DefaultAppBarContent = (props: DefaultAppBarProps) => {
  const commonStyles = useCommonStyles();
  const styles = useStyles();

  const isFlagEnabled = useFeatureFlag(FeatureFlag.OnlyMine);
  const { adminVersion } = useAdminVersion();
  const isGAEnabled = env.ENABLE_GA === 'true' && env.GA_TRACKING_ID !== '';

  const versions: VersionInfo[] = [
    {
      name: t('versionConsoleUi'),
      version: platformVersion,
      url: `https://github.com/flyteorg/flyteconsole/releases/tag/v${platformVersion}`,
    },
    {
      name: t('versionAdmin'),
      version: adminVersion,
      url: `https://github.com/flyteorg/flyteadmin/releases/tag/v${adminVersion}`,
    },
    {
      name: t('versionGoogleAnalytics'),
      version: t(patternKey('gaDisable', isGAEnabled.toString())),
      url: 'https://github.com/flyteorg/flyteconsole#google-analytics',
    },
  ];

  return (
    <>
      <Link className={classnames(commonStyles.linkUnstyled)} to={Routes.SelectProject.path}>
        <FlyteLogo size={32} />
      </Link>
      {props.items?.length > 0 ? (
        <NavigationDropdown items={props.items} console={props.console} />
      ) : (
        false
      )}
      <div className={styles.spacer} />
      {isFlagEnabled && <OnlyMine />}
      <UserInformation />
      <AppInfo versions={versions} documentationUrl="https://docs.flyte.org/en/latest/" />
    </>
  );
};
