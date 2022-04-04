import * as React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import { Tab, Tabs } from '@material-ui/core';
import { NodeExecution } from 'models/Execution/types';
import { TaskTemplate } from 'models/Task/types';
import { useTabState } from 'components/hooks/useTabState';
import { PanelSection } from 'components/common/PanelSection';
import { DumpJSON } from 'components/common/DumpJSON';
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
  taskTemplate?: TaskTemplate | null;
}> = ({ nodeExecution, taskTemplate }) => {
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
      tabContent = <TaskExecutionsList nodeExecution={nodeExecution} />;
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
          <DumpJSON value={taskTemplate} />
        </PanelSection>
      ) : null;
      break;
    }
  }
  return (
    <>
      <Tabs {...tabState} className={styles.tabs}>
        <Tab value={tabIds.executions} label="Executions" />
        <Tab value={tabIds.inputs} label="Inputs" />
        <Tab value={tabIds.outputs} label="Outputs" />
        {!!taskTemplate && <Tab value={tabIds.task} label="Task" />}
      </Tabs>
      <div className={styles.content}>{tabContent}</div>
    </>
  );
};
