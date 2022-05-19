import * as React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import { Tab, Tabs } from '@material-ui/core';
import { NodeExecution } from 'models/Execution/types';
import { TaskTemplate } from 'models/Task/types';
import { useTabState } from 'components/hooks/useTabState';
import { PanelSection } from 'components/common/PanelSection';
import { DumpJSON } from 'components/common/DumpJSON';
import { isMapTaskType } from 'models/Task/utils';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { TaskVersionDetailsLink } from 'components/Entities/VersionDetails/VersionDetailsLink';
import { Identifier } from 'models/Common/types';
import { TaskExecutionsList } from '../../TaskExecutionsList/TaskExecutionsList';
import { NodeExecutionInputs } from './NodeExecutionInputs';
import { NodeExecutionOutputs } from './NodeExecutionOutputs';

const useStyles = makeStyles((theme) => {
  return {
    content: {
      overflowY: 'auto',
    },
    tabs: {
      borderBottom: `1px solid ${theme.palette.divider}`,
      '& .MuiTabs-flexContainer': {
        justifyContent: 'space-around',
      },
    },
    tabItem: {
      margin: theme.spacing(0, 1),
    },
  };
});

const tabIds = {
  executions: 'executions',
  inputs: 'inputs',
  outputs: 'outputs',
  task: 'task',
};

const defaultTab = tabIds.executions;

export const NodeExecutionTabs: React.FC<{
  nodeExecution: NodeExecution;
  shouldShowTaskDetails: boolean;
  phase?: TaskExecutionPhase;
  taskTemplate?: TaskTemplate | null;
}> = ({ nodeExecution, shouldShowTaskDetails, taskTemplate, phase }) => {
  const styles = useStyles();
  const tabState = useTabState(tabIds, defaultTab);

  if (tabState.value === tabIds.task && !taskTemplate) {
    // Reset tab value, if task tab is selected, while no taskTemplate is avaible
    // can happen when user switches between nodeExecutions without closing the drawer
    tabState.onChange(() => {
      /* */
    }, defaultTab);
  }

  let tabContent: JSX.Element | null = null;
  switch (tabState.value) {
    case tabIds.executions: {
      tabContent = <TaskExecutionsList nodeExecution={nodeExecution} phase={phase} />;
      break;
    }
    case tabIds.inputs: {
      tabContent = <NodeExecutionInputs execution={nodeExecution} />;
      break;
    }
    case tabIds.outputs: {
      tabContent = <NodeExecutionOutputs execution={nodeExecution} />;
      break;
    }
    case tabIds.task: {
      tabContent = taskTemplate ? (
        <PanelSection>
          <TaskVersionDetailsLink id={taskTemplate.id as Identifier} />
          <DumpJSON value={taskTemplate} />
        </PanelSection>
      ) : null;
      break;
    }
  }

  const executionLabel = isMapTaskType(taskTemplate?.type)
    ? shouldShowTaskDetails
      ? 'Execution'
      : 'Map Execution'
    : 'Executions';

  return (
    <>
      <Tabs {...tabState} className={styles.tabs}>
        <Tab className={styles.tabItem} value={tabIds.executions} label={executionLabel} />
        <Tab className={styles.tabItem} value={tabIds.inputs} label="Inputs" />
        <Tab className={styles.tabItem} value={tabIds.outputs} label="Outputs" />
        {!!taskTemplate && <Tab className={styles.tabItem} value={tabIds.task} label="Task" />}
      </Tabs>
      <div className={styles.content}>{tabContent}</div>
    </>
  );
};
