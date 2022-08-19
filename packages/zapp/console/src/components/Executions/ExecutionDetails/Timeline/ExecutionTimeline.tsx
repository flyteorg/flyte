import * as React from 'react';
import { makeStyles, Typography } from '@material-ui/core';

import { useNodeExecutionContext } from 'components/Executions/contextProvider/NodeExecutionDetails';
import { transformerWorkflowToDag } from 'components/WorkflowGraph/transformerWorkflowToDag';
import { isEndNode, isStartNode, isExpanded } from 'components/WorkflowGraph/utils';
import { tableHeaderColor } from 'components/Theme/constants';
import { timestampToDate } from 'common/utils';
import { dNode } from 'models/Graph/types';
import { makeNodeExecutionDynamicWorkflowQuery } from 'components/Workflow/workflowQueries';
import { useQuery } from 'react-query';
import { createRef, useContext, useEffect, useRef, useState } from 'react';
import { NodeExecutionsByIdContext } from 'components/Executions/contexts';
import { checkForDynamicExecutions } from 'components/common/utils';
import { convertToPlainNodes } from './helpers';
import { ChartHeader } from './ChartHeader';
import { useScaleContext } from './scaleContext';
import { TaskNames } from './TaskNames';
import { getChartDurationData } from './TimelineChart/chartData';
import { TimelineChart } from './TimelineChart';

interface StyleProps {
  chartWidth: number;
  itemsShown: number;
}

const useStyles = makeStyles((theme) => ({
  chartHeader: (props: StyleProps) => ({
    marginTop: -10,
    marginLeft: -15,
    width: `${props.chartWidth + 20}px`,
    height: `${56 * props.itemsShown + 20}px`,
  }),
  taskNames: {
    display: 'flex',
    flexDirection: 'column',
    borderRight: `1px solid ${theme.palette.divider}`,
    overflowY: 'auto',
  },
  taskNamesHeader: {
    textTransform: 'uppercase',
    fontSize: 12,
    fontWeight: 'bold',
    lineHeight: '16px',
    color: tableHeaderColor,
    height: 45,
    flexBasis: 45,
    display: 'flex',
    alignItems: 'center',
    borderBottom: `4px solid ${theme.palette.divider}`,
    paddingLeft: 30,
  },
  taskDurations: {
    borderLeft: `1px solid ${theme.palette.divider}`,
    marginLeft: 4,
    flex: 1,
    overflow: 'hidden',
    display: 'flex',
    flexDirection: 'column',
  },
  taskDurationsLabelsView: {
    overflow: 'hidden',
    borderBottom: `4px solid ${theme.palette.divider}`,
  },
  taskDurationsView: {
    flex: 1,
    overflowY: 'hidden',
  },
}));

const INTERVAL_LENGTH = 110;

interface ExProps {
  chartTimezone: string;
}

export const ExecutionTimeline: React.FC<ExProps> = ({ chartTimezone }) => {
  const [chartWidth, setChartWidth] = useState(0);
  const [labelInterval, setLabelInterval] = useState(INTERVAL_LENGTH);
  const durationsRef = useRef<HTMLDivElement>(null);
  const durationsLabelsRef = useRef<HTMLDivElement>(null);
  const taskNamesRef = createRef<HTMLDivElement>();

  const [originalNodes, setOriginalNodes] = useState<dNode[]>([]);
  const [showNodes, setShowNodes] = useState<dNode[]>([]);
  const [startedAt, setStartedAt] = useState<Date>(new Date());

  const { compiledWorkflowClosure } = useNodeExecutionContext();
  const { chartInterval: chartTimeInterval } = useScaleContext();
  const { staticExecutionIdsMap } = compiledWorkflowClosure
    ? transformerWorkflowToDag(compiledWorkflowClosure)
    : [];

  const nodeExecutionsById = useContext(NodeExecutionsByIdContext);

  const dynamicParents = checkForDynamicExecutions(nodeExecutionsById, staticExecutionIdsMap);

  const { data: dynamicWorkflows } = useQuery(
    makeNodeExecutionDynamicWorkflowQuery(dynamicParents),
  );

  useEffect(() => {
    const nodes: dNode[] = compiledWorkflowClosure
      ? transformerWorkflowToDag(compiledWorkflowClosure, dynamicWorkflows).dag.nodes
      : [];
    // we remove start/end node info in the root dNode list during first assignment
    const initializeNodes = convertToPlainNodes(nodes);
    setOriginalNodes(initializeNodes);
  }, [dynamicWorkflows, compiledWorkflowClosure]);

  useEffect(() => {
    const initializeNodes = convertToPlainNodes(originalNodes);
    const updatedShownNodesMap = initializeNodes.map((node) => {
      const execution = nodeExecutionsById[node.scopedId];
      return {
        ...node,
        execution,
      };
    });
    setShowNodes(updatedShownNodesMap);

    // set startTime for all timeline offset and duration calculations.
    const firstStartedAt = updatedShownNodesMap[0]?.execution?.closure.startedAt;
    if (firstStartedAt) {
      setStartedAt(timestampToDate(firstStartedAt));
    }
  }, [originalNodes, nodeExecutionsById]);

  const { items: barItemsData, totalDurationSec } = getChartDurationData(showNodes, startedAt);
  const styles = useStyles({ chartWidth: chartWidth, itemsShown: showNodes.length });

  useEffect(() => {
    // Sync width of elements and intervals of ChartHeader (time labels) and TimelineChart
    const calcWidth = Math.ceil(totalDurationSec / chartTimeInterval) * INTERVAL_LENGTH;
    if (durationsRef.current && calcWidth < durationsRef.current.clientWidth) {
      setLabelInterval(
        durationsRef.current.clientWidth / Math.ceil(totalDurationSec / chartTimeInterval),
      );
      setChartWidth(durationsRef.current.clientWidth);
    } else {
      setChartWidth(calcWidth);
      setLabelInterval(INTERVAL_LENGTH);
    }
  }, [totalDurationSec, chartTimeInterval, durationsRef]);

  const onGraphScroll = () => {
    // cover horizontal scroll only
    const scrollLeft = durationsRef?.current?.scrollLeft ?? 0;
    const labelView = durationsLabelsRef?.current;
    if (labelView) {
      labelView.scrollTo({
        left: scrollLeft,
      });
    }
  };

  const onVerticalNodesScroll = () => {
    const scrollTop = taskNamesRef?.current?.scrollTop ?? 0;
    const graphView = durationsRef?.current;
    if (graphView) {
      graphView.scrollTo({
        top: scrollTop,
      });
    }
  };

  const toggleNode = (id: string, scopeId: string, level: number) => {
    const searchNode = (nodes: dNode[], nodeLevel: number) => {
      if (!nodes || nodes.length === 0) {
        return;
      }
      for (let i = 0; i < nodes.length; i++) {
        const node = nodes[i];
        if (isStartNode(node) || isEndNode(node)) {
          continue;
        }
        if (node.id === id && node.scopedId === scopeId && nodeLevel === level) {
          nodes[i].expanded = !nodes[i].expanded;
          return;
        }
        if (node.nodes.length > 0 && isExpanded(node)) {
          searchNode(node.nodes, nodeLevel + 1);
        }
      }
    };
    searchNode(originalNodes, 0);
    setOriginalNodes([...originalNodes]);
  };

  return (
    <>
      <div className={styles.taskNames}>
        <Typography className={styles.taskNamesHeader}>Task Name</Typography>
        <TaskNames
          nodes={showNodes}
          ref={taskNamesRef}
          onToggle={toggleNode}
          onScroll={onVerticalNodesScroll}
        />
      </div>
      <div className={styles.taskDurations}>
        <div className={styles.taskDurationsLabelsView} ref={durationsLabelsRef}>
          <ChartHeader
            startedAt={startedAt}
            chartWidth={chartWidth}
            labelInterval={labelInterval}
            chartTimezone={chartTimezone}
            totalDurationSec={totalDurationSec}
          />
        </div>
        <div className={styles.taskDurationsView} ref={durationsRef} onScroll={onGraphScroll}>
          <div className={styles.chartHeader}>
            <TimelineChart items={barItemsData} chartTimeIntervalSec={chartTimeInterval} />
          </div>
        </div>
      </div>
    </>
  );
};
