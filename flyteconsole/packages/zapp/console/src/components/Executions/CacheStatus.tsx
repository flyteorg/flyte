import { SvgIconProps, Tooltip, Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import CachedOutlined from '@material-ui/icons/CachedOutlined';
import ErrorOutlined from '@material-ui/icons/ErrorOutlined';
import InfoOutlined from '@material-ui/icons/InfoOutlined';
import classnames from 'classnames';
import { assertNever } from 'common/utils';
import { PublishedWithChangesOutlined } from 'components/common/PublishedWithChanges';
import { useCommonStyles } from 'components/common/styles';
import { CatalogCacheStatus } from 'models/Execution/enums';
import { TaskExecutionIdentifier } from 'models/Execution/types';
import { MapCacheIcon } from '@flyteconsole/ui-atoms';
import * as React from 'react';
import { Link as RouterLink } from 'react-router-dom';
import { Routes } from 'routes/routes';
import {
  cacheStatusMessages,
  unknownCacheStatusString,
  viewSourceExecutionString,
} from './constants';

const useStyles = makeStyles((theme: Theme) => ({
  cacheStatus: {
    alignItems: 'center',
    display: 'flex',
    marginTop: theme.spacing(1),
  },
  sourceExecutionLink: {
    fontWeight: 'normal',
  },
}));

/** Renders the appropriate icon for a given CatalogCacheStatus */
const NodeExecutionCacheStatusIcon: React.FC<
  SvgIconProps & {
    status: CatalogCacheStatus;
  }
> = React.forwardRef(({ status, ...props }, ref) => {
  switch (status) {
    case CatalogCacheStatus.CACHE_DISABLED:
    case CatalogCacheStatus.CACHE_MISS: {
      return <InfoOutlined {...props} ref={ref} data-testid="cache-icon" />;
    }
    case CatalogCacheStatus.CACHE_HIT: {
      return <CachedOutlined {...props} ref={ref} data-testid="cache-icon" />;
    }
    case CatalogCacheStatus.CACHE_POPULATED: {
      return <PublishedWithChangesOutlined {...props} ref={ref} data-testid="cache-icon" />;
    }
    case CatalogCacheStatus.CACHE_LOOKUP_FAILURE:
    case CatalogCacheStatus.CACHE_PUT_FAILURE: {
      return <ErrorOutlined {...props} ref={ref} data-testid="cache-icon" />;
    }
    case CatalogCacheStatus.MAP_CACHE: {
      return <MapCacheIcon {...props} ref={ref} data-testid="cache-icon" />;
    }
    default: {
      assertNever(status as never);
      return null;
    }
  }
});

export interface CacheStatusProps {
  cacheStatus: CatalogCacheStatus | null | undefined;
  /** `normal` will render an icon with description message beside it
   *  `iconOnly` will render just the icon with the description as a tooltip
   */
  variant?: 'normal' | 'iconOnly';
  sourceTaskExecutionId?: TaskExecutionIdentifier;
  iconStyles?: React.CSSProperties;
}

export const CacheStatus: React.FC<CacheStatusProps> = ({
  cacheStatus,
  sourceTaskExecutionId,
  variant = 'normal',
  iconStyles,
}) => {
  const commonStyles = useCommonStyles();
  const styles = useStyles();

  if (cacheStatus == null) {
    return null;
  }

  const message = cacheStatusMessages[cacheStatus] || unknownCacheStatusString;

  return variant === 'iconOnly' ? (
    <Tooltip title={message}>
      <NodeExecutionCacheStatusIcon
        className={classnames(commonStyles.iconRight, commonStyles.iconSecondary)}
        style={iconStyles}
        status={cacheStatus}
      />
    </Tooltip>
  ) : (
    <>
      <Typography className={styles.cacheStatus} variant="subtitle1" color="textSecondary">
        <NodeExecutionCacheStatusIcon
          status={cacheStatus}
          className={classnames(commonStyles.iconSecondary, commonStyles.iconLeft)}
        />
        {message}
      </Typography>
      {sourceTaskExecutionId && (
        <RouterLink
          className={classnames(commonStyles.primaryLink, styles.sourceExecutionLink)}
          to={Routes.ExecutionDetails.makeUrl(sourceTaskExecutionId.nodeExecutionId.executionId)}
        >
          {viewSourceExecutionString}
        </RouterLink>
      )}
    </>
  );
};
