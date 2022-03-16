import { useEffect, useRef, useState } from 'react';
import { IconButton, Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import Tab from '@material-ui/core/Tab';
import Tabs from '@material-ui/core/Tabs';
import Close from '@material-ui/icons/Close';
import classnames from 'classnames';
import { useCommonStyles } from 'components/common/styles';
import { InfoIcon } from 'components/common/Icons/InfoIcon';
import { bodyFontFamily, smallFontSize } from 'components/Theme/constants';
import { ExecutionStatusBadge } from 'components/Executions/ExecutionStatusBadge';
import { LocationState } from 'components/hooks/useLocationState';
import { useTabState } from 'components/hooks/useTabState';
import { LocationDescriptor } from 'history';
import { PaginatedEntityResponse } from 'models/AdminEntity/types';
import { Workflow } from 'models/Workflow/types';
import { NodeExecution, NodeExecutionIdentifier, TaskExecution } from 'models/Execution/types';
import { TaskTemplate } from 'models/Task/types';
import * as React from 'react';
import Skeleton from 'react-loading-skeleton';
import { useQuery, useQueryClient } from 'react-query';
import { Link as RouterLink } from 'react-router-dom';
import { Routes } from 'routes/routes';
import { RemoteLiteralMapViewer } from 'components/Literals/RemoteLiteralMapViewer';
import { fetchWorkflow } from 'components/Workflow/workflowQueries';
import { dNode } from 'models/Graph/types';
import { transformWorkflowToKeyedDag, getNodeNameFromDag } from 'components/WorkflowGraph/utils';
import { NodeExecutionCacheStatus } from '../NodeExecutionCacheStatus';
import { makeListTaskExecutionsQuery, makeNodeExecutionQuery } from '../nodeExecutionQueries';
import { TaskExecutionsList } from '../TaskExecutionsList/TaskExecutionsList';
import { NodeExecutionDetails } from '../types';
import { useNodeExecutionContext } from '../contextProvider/NodeExecutionDetails';
import { NodeExecutionInputs } from './NodeExecutionInputs';
import { NodeExecutionOutputs } from './NodeExecutionOutputs';
import { NodeExecutionTaskDetails } from './NodeExecutionTaskDetails';
import { getTaskExecutionDetailReasons } from './utils';
import { ExpandableMonospaceText } from '../../common/ExpandableMonospaceText';
import { fetchWorkflowExecution } from '../useWorkflowExecution';

const useStyles = makeStyles((theme: Theme) => {
  const paddingVertical = `${theme.spacing(2)}px`;
  const paddingHorizontal = `${theme.spacing(3)}px`;
  return {
    notRunStatus: {
      alignItems: 'center',
      backgroundColor: 'gray',
      borderRadius: '4px',
      color: theme.palette.text.primary,
      display: 'flex',
      flex: '0 0 auto',
      height: theme.spacing(3),
      fontSize: smallFontSize,
      justifyContent: 'center',
      textTransform: 'uppercase',
      width: theme.spacing(11),
      fontFamily: bodyFontFamily,
      fontWeight: 'bold',
    },
    closeButton: {
      marginLeft: theme.spacing(1),
    },
    container: {
      display: 'flex',
      flexDirection: 'column',
      height: '100%',
      paddingTop: theme.spacing(2),
      width: '100%',
    },
    content: {
      overflowY: 'auto',
    },
    displayId: {
      marginBottom: theme.spacing(1),
    },
    header: {
      borderBottom: `${theme.spacing(1)}px solid ${theme.palette.divider}`,
    },
    headerContent: {
      padding: `0 ${paddingHorizontal} ${paddingVertical} ${paddingHorizontal}`,
    },
    nodeTypeContainer: {
      alignItems: 'flex-end',
      borderTop: `1px solid ${theme.palette.divider}`,
      display: 'flex',
      flexDirection: 'row',
      fontWeight: 'bold',
      justifyContent: 'space-between',
      marginTop: theme.spacing(2),
      paddingTop: theme.spacing(2),
    },
    nodeTypeContent: {
      minWidth: theme.spacing(9),
    },
    nodeTypeLink: {
      fontWeight: 'normal',
    },
    tabs: {
      borderBottom: `1px solid ${theme.palette.divider}`,
    },
    title: {
      alignItems: 'flex-start',
      display: 'flex',
      justifyContent: 'space-between',
    },
    statusContainer: {
      display: 'flex',
      flexDirection: 'column',
    },
    statusHeaderContainer: {
      display: 'flex',
      alignItems: 'center',
    },
    reasonsIcon: {
      marginLeft: theme.spacing(1),
      cursor: 'pointer',
    },
    statusBody: {
      marginTop: theme.spacing(2),
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

interface NodeExecutionDetailsProps {
  nodeExecutionId: NodeExecutionIdentifier;
  onClose?: () => void;
}

const NodeExecutionLinkContent: React.FC<{
  execution: NodeExecution;
}> = ({ execution }) => {
  const commonStyles = useCommonStyles();
  const styles = useStyles();
  const { workflowNodeMetadata } = execution.closure;
  if (!workflowNodeMetadata) {
    return null;
  }
  const linkTarget: LocationDescriptor<LocationState> = {
    pathname: Routes.ExecutionDetails.makeUrl(workflowNodeMetadata.executionId),
    state: {
      backLink: Routes.ExecutionDetails.makeUrl(execution.id.executionId),
    },
  };
  return (
    <RouterLink
      className={classnames(commonStyles.primaryLink, styles.nodeTypeLink)}
      to={linkTarget}
    >
      View Sub-Workflow
    </RouterLink>
  );
};

const ExecutionTypeDetails: React.FC<{
  details?: NodeExecutionDetails;
  execution: NodeExecution;
}> = ({ details, execution }) => {
  const styles = useStyles();
  const commonStyles = useCommonStyles();
  return (
    <div className={classnames(commonStyles.textSmall, styles.nodeTypeContainer)}>
      <div className={styles.nodeTypeContent}>
        <div className={classnames(commonStyles.microHeader, commonStyles.textMuted)}>Type</div>
        <div>{details ? details.displayType : <Skeleton />}</div>
      </div>
      {<NodeExecutionLinkContent execution={execution} />}
    </div>
  );
};

const NodeExecutionTabs: React.FC<{
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
      tabContent = taskTemplate ? <NodeExecutionTaskDetails taskTemplate={taskTemplate} /> : null;
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

const WorkflowTabs: React.FC<{
  dagData: dNode;
  nodeId: string;
}> = ({ dagData, nodeId }) => {
  const styles = useStyles();
  const tabState = useTabState(tabIds, tabIds.inputs);
  const commonStyles = useCommonStyles();
  let tabContent: JSX.Element | null = null;
  const id = nodeId.slice(nodeId.lastIndexOf('-') + 1);
  const taskTemplate = dagData[id].value.template;

  switch (tabState.value) {
    case tabIds.inputs: {
      tabContent = taskTemplate ? (
        <div className={commonStyles.detailsPanelCard}>
          <div className={commonStyles.detailsPanelCardContent}>
            <RemoteLiteralMapViewer blob={taskTemplate.interface.inputs} map={null} />
          </div>
        </div>
      ) : null;
      break;
    }
    case tabIds.task: {
      tabContent = taskTemplate ? <NodeExecutionTaskDetails taskTemplate={taskTemplate} /> : null;
      break;
    }
  }
  return (
    <>
      <Tabs {...tabState} className={styles.tabs}>
        <Tab value={tabIds.inputs} label="Inputs" />
        {!!taskTemplate && <Tab value={tabIds.task} label="Task" />}
      </Tabs>
      <div className={styles.content}>{tabContent}</div>
    </>
  );
};

/** DetailsPanel content which renders execution information about a given NodeExecution
 */
export const NodeExecutionDetailsPanelContent: React.FC<NodeExecutionDetailsProps> = ({
  nodeExecutionId,
  onClose,
}) => {
  const isMounted = useRef(false);
  useEffect(() => {
    isMounted.current = true;
    return () => {
      isMounted.current = false;
    };
  }, []);

  const queryClient = useQueryClient();
  const detailsContext = useNodeExecutionContext();

  const [isReasonsVisible, setReasonsVisible] = React.useState(false);
  const [dag, setDag] = React.useState<any>(null);
  const [details, setDetails] = React.useState<NodeExecutionDetails | undefined>();

  const nodeExecutionQuery = useQuery<NodeExecution, Error>({
    ...makeNodeExecutionQuery(nodeExecutionId),
    // The selected NodeExecution has been fetched at this point, we don't want to
    // issue an additional fetch.
    staleTime: Infinity,
  });

  React.useEffect(() => {
    let isCurrent = true;
    detailsContext.getNodeExecutionDetails(nodeExecution).then((res) => {
      if (isCurrent) {
        setDetails(res);
      }
    });

    return () => {
      isCurrent = false;
    };
  });

  React.useEffect(() => {
    setReasonsVisible(false);
  }, [nodeExecutionId]);

  const nodeExecution = nodeExecutionQuery.data;

  const getWorkflowDag = async () => {
    const workflowExecution = await fetchWorkflowExecution(
      queryClient,
      nodeExecutionId.executionId,
    );
    const workflowData: Workflow = await fetchWorkflow(
      queryClient,
      workflowExecution.closure.workflowId,
    );
    if (workflowData) {
      const keyedDag = transformWorkflowToKeyedDag(workflowData);
      if (isMounted.current) setDag(keyedDag);
    }
  };

  if (!nodeExecution) {
    getWorkflowDag();
  } else {
    if (dag) setDag(null);
  }
  const listTaskExecutionsQuery = useQuery<PaginatedEntityResponse<TaskExecution>, Error>({
    ...makeListTaskExecutionsQuery(nodeExecutionId),
    staleTime: Infinity,
  });

  const reasons = getTaskExecutionDetailReasons(listTaskExecutionsQuery.data);

  const commonStyles = useCommonStyles();
  const styles = useStyles();
  const displayName = details?.displayName ?? <Skeleton />;

  const isRunningPhase = React.useMemo(() => {
    return nodeExecution?.closure.phase === 1 || nodeExecution?.closure.phase === 2;
  }, [nodeExecution]);

  const handleReasonsVisibility = React.useCallback(() => {
    setReasonsVisible((prevVisibility) => !prevVisibility);
  }, []);

  const statusContent = nodeExecution ? (
    <div className={styles.statusContainer}>
      <div className={styles.statusHeaderContainer}>
        <ExecutionStatusBadge phase={nodeExecution.closure.phase} type="node" />
        {isRunningPhase && (
          <InfoIcon className={styles.reasonsIcon} onClick={handleReasonsVisibility} />
        )}
      </div>
      {isRunningPhase && isReasonsVisible && (
        <div className={styles.statusBody}>
          <ExpandableMonospaceText initialExpansionState={false} text={reasons.join('\n')} />
        </div>
      )}
    </div>
  ) : (
    <div className={styles.notRunStatus}>NOT RUN</div>
  );

  const detailsContent = nodeExecution ? (
    <>
      <NodeExecutionCacheStatus taskNodeMetadata={nodeExecution.closure.taskNodeMetadata} />
      <ExecutionTypeDetails details={details} execution={nodeExecution} />
    </>
  ) : null;

  const tabsContent = nodeExecution ? (
    <NodeExecutionTabs nodeExecution={nodeExecution} taskTemplate={details?.taskTemplate} />
  ) : null;
  return (
    <section className={styles.container}>
      <header className={styles.header}>
        <div className={styles.headerContent}>
          <Typography className={classnames(commonStyles.textWrapped, styles.title)} variant="h3">
            {nodeExecutionId.nodeId}
            <IconButton className={styles.closeButton} onClick={onClose} size="small">
              <Close />
            </IconButton>
          </Typography>
          <Typography
            className={classnames(commonStyles.textWrapped, styles.displayId)}
            variant="subtitle1"
            color="textSecondary"
          >
            {dag ? getNodeNameFromDag(dag, nodeExecutionId.nodeId) : displayName}
          </Typography>
          {statusContent}
          {!dag && detailsContent}
        </div>
      </header>
      {dag ? <WorkflowTabs nodeId={nodeExecutionId.nodeId} dagData={dag} /> : tabsContent}
    </section>
  );
};
