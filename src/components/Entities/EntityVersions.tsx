import * as React from 'react';
import { Routes } from 'routes/routes';
import { history } from 'routes/history';
import Typography from '@material-ui/core/Typography';
import { IconButton, makeStyles, Theme } from '@material-ui/core';
import ExpandLess from '@material-ui/icons/ExpandLess';
import ExpandMore from '@material-ui/icons/ExpandMore';
import { LocalCacheItem, useLocalCache } from 'basics/LocalCache';
import { WaitForData } from 'components/common/WaitForData';
import { WorkflowVersionsTable } from 'components/Executions/Tables/WorkflowVersionsTable';
import { isLoadingState } from 'components/hooks/fetchMachine';
import { useWorkflowVersions } from 'components/hooks/useWorkflowVersions';
import { interactiveTextColor } from 'components/Theme/constants';
import { SortDirection } from 'models/AdminEntity/types';
import { ResourceIdentifier } from 'models/Common/types';
import { executionSortFields } from 'models/Execution/constants';
import { executionFilterGenerator } from './generators';
import { WorkflowVersionsTablePageSize } from './constants';
import t from './strings';

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

  const versions = useWorkflowVersions(
    { domain, project },
    {
      sort,
      filter: baseFilters,
      limit: showAll ? 100 : WorkflowVersionsTablePageSize,
    },
  );

  const preventDefault = (e) => e.preventDefault();
  const handleViewAll = React.useCallback(() => {
    history.push(
      Routes.WorkflowVersionDetails.makeUrl(
        project,
        domain,
        name,
        versions.value[0].id.version ?? '',
      ),
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
          <Typography className={styles.header} variant="h6">
            {t('workflowVersionsTitle')}
          </Typography>
          <Typography className={styles.viewAll} variant="body1" onClick={handleViewAll}>
            {t('viewAll')}
          </Typography>
        </div>
      )}
      <WaitForData {...versions}>
        {showTable || showAll ? (
          <WorkflowVersionsTable
            {...versions}
            isFetching={isLoadingState(versions.state)}
            versionView={showAll}
          />
        ) : (
          <div className={styles.divider} />
        )}
      </WaitForData>
    </>
  );
};
