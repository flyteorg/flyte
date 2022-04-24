import { makeStyles, Theme } from '@material-ui/core/styles';
import { contentMarginGridUnits } from 'common/layout';
import { WaitForData } from 'components/common/WaitForData';
import { EntityDescription } from 'components/Entities/EntityDescription';
import { useProject } from 'components/hooks/useProjects';
import { useChartState } from 'components/hooks/useChartState';
import { ResourceIdentifier } from 'models/Common/types';
import * as React from 'react';
import { entitySections } from './constants';
import { EntityDetailsHeader } from './EntityDetailsHeader';
import { EntityExecutions } from './EntityExecutions';
import { EntitySchedules } from './EntitySchedules';
import { EntityVersions } from './EntityVersions';
import { EntityExecutionsBarChart } from './EntityExecutionsBarChart';

const useStyles = makeStyles((theme: Theme) => ({
  metadataContainer: {
    display: 'flex',
    marginBottom: theme.spacing(5),
    marginTop: theme.spacing(2),
    width: '100%',
  },
  descriptionContainer: {
    flex: '2 1 auto',
    marginRight: theme.spacing(2),
  },
  executionsContainer: {
    display: 'flex',
    flex: '1 1 auto',
    flexDirection: 'column',
    margin: `0 -${theme.spacing(contentMarginGridUnits)}px`,
    flexBasis: theme.spacing(80),
  },
  versionsContainer: {
    display: 'flex',
    flexDirection: 'column',
  },
  schedulesContainer: {
    flex: '1 2 auto',
    marginRight: theme.spacing(30),
  },
}));

interface EntityDetailsProps {
  id: ResourceIdentifier;
}

/**
 * A view which optionally renders description, schedules, executions, and a
 * launch button/form for a given entity. Note: not all components are suitable
 * for use with all entities (not all entities have schedules, for example).
 * @param id
 */
export const EntityDetails: React.FC<EntityDetailsProps> = ({ id }) => {
  const sections = entitySections[id.resourceType];
  const project = useProject(id.project);
  const styles = useStyles();
  const { chartIds, onToggle, clearCharts } = useChartState();

  return (
    <WaitForData {...project}>
      <EntityDetailsHeader project={project.value} id={id} launchable={!!sections.launch} />

      <div className={styles.metadataContainer}>
        {sections.description ? (
          <div className={styles.descriptionContainer}>
            <EntityDescription id={id} />
          </div>
        ) : null}
        {sections.schedules ? (
          <div className={styles.schedulesContainer}>
            <EntitySchedules id={id} />
          </div>
        ) : null}
      </div>

      {sections.versions ? (
        <div className={styles.versionsContainer}>
          <EntityVersions id={id} />
        </div>
      ) : null}

      <EntityExecutionsBarChart onToggle={onToggle} chartIds={chartIds} id={id} />

      {sections.executions ? (
        <div className={styles.executionsContainer}>
          <EntityExecutions chartIds={chartIds} id={id} clearCharts={clearCharts} />
        </div>
      ) : null}
    </WaitForData>
  );
};
