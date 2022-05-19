import * as React from 'react';
import { history } from 'routes/history';
import Typography from '@material-ui/core/Typography';
import { IconButton, makeStyles, Theme } from '@material-ui/core';
import ExpandLess from '@material-ui/icons/ExpandLess';
import ExpandMore from '@material-ui/icons/ExpandMore';
import { LocalCacheItem, useLocalCache } from 'basics/LocalCache';
import { WaitForData } from 'components/common/WaitForData';
import { EntityVersionsTable } from 'components/Executions/Tables/EntityVersionsTable';
import { isLoadingState } from 'components/hooks/fetchMachine';
import { useEntityVersions } from 'components/hooks/Entity/useEntityVersions';
import { interactiveTextColor } from 'components/Theme/constants';
import { SortDirection } from 'models/AdminEntity/types';
import { Identifier, ResourceIdentifier } from 'models/Common/types';
import { executionSortFields } from 'models/Execution/constants';
import { executionFilterGenerator, versionDetailsUrlGenerator } from './generators';
import { WorkflowVersionsTablePageSize, entityStrings } from './constants';
import t, { patternKey } from './strings';

const useStyles = makeStyles((theme: Theme) => ({
  headerContainer: {
    display: 'flex',
  },
  collapseButton: {
    marginTop: theme.spacing(-0.5),
  },
  header: {
    flexGrow: 1,
    marginBottom: theme.spacing(1),
    marginRight: theme.spacing(1),
  },
  viewAll: {
    color: interactiveTextColor,
    cursor: 'pointer',
  },
  divider: {
    borderBottom: `1px solid ${theme.palette.divider}`,
    marginBottom: theme.spacing(1),
  },
}));

export interface EntityVersionsProps {
  id: ResourceIdentifier;
  showAll?: boolean;
}

/**
 * The tab/page content for viewing a workflow's versions.
 * @param id
 * @param showAll - shows all available entity versions
 */
export const EntityVersions: React.FC<EntityVersionsProps> = ({ id, showAll = false }) => {
  const { domain, project, resourceType, name } = id;
  const [showTable, setShowTable] = useLocalCache(LocalCacheItem.ShowWorkflowVersions);
  const styles = useStyles();
  const sort = {
    key: executionSortFields.createdAt,
    direction: SortDirection.DESCENDING,
  };

  const baseFilters = React.useMemo(
    () => executionFilterGenerator[resourceType](id),
    [id, resourceType],
  );

  // we are getting all the versions for this id
  // so we don't want to specify which version
  const versions = useEntityVersions(
    { ...id, version: '' },
    {
      sort,
      filter: baseFilters,
      limit: showAll ? 100 : WorkflowVersionsTablePageSize,
    },
  );

  const preventDefault = (e) => e.preventDefault();
  const handleViewAll = React.useCallback(() => {
    history.push(
      versionDetailsUrlGenerator({
        ...id,
        version: versions.value[0].id.version ?? '',
      } as Identifier),
    );
  }, [project, domain, name, versions]);

  return (
    <>
      {!showAll && (
        <div className={styles.headerContainer}>
          <IconButton
            className={styles.collapseButton}
            edge="start"
            disableRipple={true}
            disableTouchRipple={true}
            onClick={() => setShowTable(!showTable)}
            onMouseDown={preventDefault}
            size="small"
            aria-label=""
            title={t('collapseButton', showTable)}
          >
            {showTable ? <ExpandLess /> : <ExpandMore />}
          </IconButton>
          <Typography className={styles.header} variant="h3">
            {t(patternKey('versionsTitle', entityStrings[id.resourceType]))}
          </Typography>
          <Typography className={styles.viewAll} variant="body1" onClick={handleViewAll}>
            {t('viewAll')}
          </Typography>
        </div>
      )}
      <WaitForData {...versions}>
        {showTable || showAll ? (
          <EntityVersionsTable
            {...versions}
            isFetching={isLoadingState(versions.state)}
            versionView={showAll}
            resourceType={resourceType}
          />
        ) : (
          <div className={styles.divider} />
        )}
      </WaitForData>
    </>
  );
};
