import * as React from 'react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { IconButton, Typography, Tab, Tabs } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import Close from '@material-ui/icons/Close';
import ArrowBackIos from '@material-ui/icons/ArrowBackIos';
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
import Skeleton from 'react-loading-skeleton';
import { useQuery, useQueryClient } from 'react-query';
import { Link as RouterLink } from 'react-router-dom';
import { Routes } from 'routes/routes';
import { NoDataIsAvailable } from 'components/Literals/LiteralMapViewer';
import { fetchWorkflow } from 'components/Workflow/workflowQueries';
import { PanelSection } from 'components/common/PanelSection';
import { DumpJSON } from 'components/common/DumpJSON';
import { dNode } from 'models/Graph/types';
import { NodeExecutionPhase, TaskExecutionPhase } from 'models/Execution/enums';
import { transformWorkflowToKeyedDag, getNodeNameFromDag } from 'components/WorkflowGraph/utils';
import { TaskVersionDetailsLink } from 'components/Entities/VersionDetails/VersionDetailsLink';
import { Identifier } from 'models/Common/types';
import { NodeExecutionCacheStatus } from '../NodeExecutionCacheStatus';
import { makeListTaskExecutionsQuery, makeNodeExecutionQuery } from '../nodeExecutionQueries';
import { NodeExecutionDetails } from '../types';
import { useNodeExecutionContext } from '../contextProvider/NodeExecutionDetails';
import { getTaskExecutionDetailReasons } from './utils';
import { ExpandableMonospaceText } from '../../common/ExpandableMonospaceText';
import { fetchWorkflowExecution } from '../useWorkflowExecution';
import { NodeExecutionTabs } from './NodeExecutionTabs';

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

interface NodeExecutionDetailsProps {
  nodeExecutionId: NodeExecutionIdentifier;
  phase?: TaskExecutionPhase;
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
      <NodeExecutionLinkContent execution={execution} />
    </div>
  );
};

// TODO FC#393: Check if it could be replaced with tabsContent or simplified further.
// Check if we need to request task info instead of relying on dag
// Also check strange setDag pattern
const WorkflowTabs: React.FC<{
  dagData: dNode;
  nodeId: string;
}> = ({ dagData, nodeId }) => {
  const styles = useStyles();
  const tabState = useTabState(tabIds, tabIds.inputs);

  let tabContent: JSX.Element | null = null;
  const id = nodeId.slice(nodeId.lastIndexOf('-') + 1);
  const taskTemplate = dagData[id]?.value.template;
  switch (tabState.value) {
    case tabIds.inputs: {
      tabContent = taskTemplate ? (
        <PanelSection>
          <NoDataIsAvailable />
        </PanelSection>
      ) : null;
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
  phase,
  onClose,
}) => {
  const commonStyles = useCommonStyles();
  const styles = useStyles();
  const queryClient = useQueryClient();
  const detailsContext = useNodeExecutionContext();

  const [isReasonsVisible, setReasonsVisible] = useState<boolean>(false);
  const [dag, setDag] = useState<any>(null);
  const [details, setDetails] = useState<NodeExecutionDetails | undefined>();
  const [shouldShowTaskDetails, setShouldShowTaskDetails] = useState<boolean>(false); // TODO to be reused in https://github.com/flyteorg/flyteconsole/issues/312

  const isMounted = useRef(false);
  useEffect(() => {
    isMounted.current = true;
    return () => {
      isMounted.current = false;
    };
  }, []);

  const nodeExecutionQuery = useQuery<NodeExecution, Error>({
    ...makeNodeExecutionQuery(nodeExecutionId),
    // The selected NodeExecution has been fetched at this point, we don't want to
    // issue an additional fetch.
    staleTime: Infinity,
  });

  useEffect(() => {
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

  useEffect(() => {
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

  const onBackClick = () => {
    setShouldShowTaskDetails(false);
  };

  const headerTitle = useMemo(() => {
    // TODO to be reused in https://github.com/flyteorg/flyteconsole/issues/312
    // // eslint-disable-next-line no-useless-escape
    // const regex = /\-([\w\s-]+)\-/; // extract string between first and last dash

    // const mapTaskHeader = `${mapTask?.[0].externalId?.match(regex)?.[1]} of ${
    //   nodeExecutionId.nodeId
    // }`;
    // const header = shouldShowTaskDetails ? mapTaskHeader : nodeExecutionId.nodeId;
    const header = nodeExecutionId.nodeId;

    return (
      <Typography className={classnames(commonStyles.textWrapped, styles.title)} variant="h3">
        <div>
          {shouldShowTaskDetails && (
            <IconButton onClick={onBackClick} size="small">
              <ArrowBackIos />
            </IconButton>
          )}
          {header}
        </div>
        <IconButton className={styles.closeButton} onClick={onClose} size="small">
          <Close />
        </IconButton>
      </Typography>
    );
  }, [nodeExecutionId, shouldShowTaskDetails]);

  const isRunningPhase = useMemo(() => {
    return (
      nodeExecution?.closure.phase === NodeExecutionPhase.QUEUED ||
      nodeExecution?.closure.phase === NodeExecutionPhase.RUNNING
    );
  }, [nodeExecution]);

  const handleReasonsVisibility = () => {
    setReasonsVisible(!isReasonsVisible);
  };

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

  let detailsContent: JSX.Element | null = null;
  if (nodeExecution) {
    detailsContent = (
      <>
        <NodeExecutionCacheStatus taskNodeMetadata={nodeExecution.closure.taskNodeMetadata} />
        <ExecutionTypeDetails details={details} execution={nodeExecution} />
      </>
    );
  }

  const tabsContent: JSX.Element | null = nodeExecution ? (
    <NodeExecutionTabs
      nodeExecution={nodeExecution}
      shouldShowTaskDetails={shouldShowTaskDetails}
      phase={phase}
      taskTemplate={details?.taskTemplate}
    />
  ) : null;

  const displayName = details?.displayName ?? <Skeleton />;

  return (
    <section className={styles.container}>
      <header className={styles.header}>
        <div className={styles.headerContent}>
          {headerTitle}
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
