import * as React from 'react';
import * as moment from 'moment-timezone';
import { Bar } from 'react-chartjs-2';
import { Chart as ChartJS, registerables } from 'chart.js';
import ChartDataLabels from 'chartjs-plugin-datalabels';
import { makeStyles, Typography } from '@material-ui/core';

import { useNodeExecutionContext } from 'components/Executions/contextProvider/NodeExecutionDetails';
import { transformerWorkflowToPlainNodes } from 'components/WorkflowGraph/transformerWorkflowToDag';
import { isEndNode, isStartNode, isExpanded } from 'components/WorkflowGraph/utils';
import { tableHeaderColor } from 'components/Theme/constants';
import { NodeExecution } from 'models/Execution/types';
import { dNode } from 'models/Graph/types';
import { TaskNames } from './taskNames';
import { convertToPlainNodes, getBarOptions, TimeZone } from './helpers';
import { ChartHeader } from './chartHeader';
import { useChartDurationData } from './chartData';

// Register components to be usable by chart.js
ChartJS.register(...registerables, ChartDataLabels);

interface StyleProps {
  chartWidth: number;
  durationLength: number;
}

const useStyles = makeStyles(theme => ({
  chartHeader: (props: StyleProps) => ({
    marginTop: -10,
    marginLeft: -15,
    width: `${props.chartWidth + 20}px`,
    height: `${56 * props.durationLength + 20}px`
  }),
  taskNames: {
    display: 'flex',
    flexDirection: 'column',
    borderRight: `1px solid ${theme.palette.divider}`,
    overflowY: 'auto'
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
    paddingLeft: 30
  },
  taskDurations: {
    borderLeft: `1px solid ${theme.palette.divider}`,
    marginLeft: 4,
    flex: 1,
    overflow: 'hidden',
    display: 'flex',
    flexDirection: 'column'
  },
  taskDurationsLabelsView: {
    overflow: 'hidden',
    borderBottom: `4px solid ${theme.palette.divider}`
  },
  taskDurationsView: {
    flex: 1,
    overflowY: 'hidden'
  }
}));

const INTERVAL_LENGTH = 110;

interface ExProps {
  nodeExecutions: NodeExecution[];
  chartTimeInterval: number;
  chartTimezone: string;
}

export const ExecutionTimeline: React.FC<ExProps> = ({ nodeExecutions, chartTimeInterval, chartTimezone }) => {
  const [chartWidth, setChartWidth] = React.useState(0);
  const [labelInterval, setLabelInterval] = React.useState(INTERVAL_LENGTH);
  const durationsRef = React.useRef<HTMLDivElement>(null);
  const durationsLabelsRef = React.useRef<HTMLDivElement>(null);
  const taskNamesRef = React.createRef<HTMLDivElement>();

  const [originalNodes, setOriginalNodes] = React.useState<dNode[]>([]);

  const { compiledWorkflowClosure } = useNodeExecutionContext();

  // narusina - we need to propely merge executions with planeNodes instead of original Nodes
  React.useEffect(() => {
    const { nodes: originalNodes } = transformerWorkflowToPlainNodes(compiledWorkflowClosure!);
    setOriginalNodes(
      originalNodes.map(node => {
        const index = nodeExecutions.findIndex(exe => exe.id.nodeId === node.id && exe.scopedId === node.scopedId);
        return {
          ...node,
          execution: index >= 0 ? nodeExecutions[index] : undefined
        };
      })
    );
  }, [compiledWorkflowClosure, nodeExecutions]);

  const nodes = convertToPlainNodes(originalNodes);

  const { startedAt, totalDuration, durationLength, chartData } = useChartDurationData({ nodes: nodes });
  const styles = useStyles({ chartWidth: chartWidth, durationLength: durationLength });

  React.useEffect(() => {
    const calcWidth = Math.ceil(totalDuration / chartTimeInterval) * INTERVAL_LENGTH;
    if (!durationsRef.current) {
      setChartWidth(calcWidth);
      setLabelInterval(INTERVAL_LENGTH);
    } else if (calcWidth < durationsRef.current.clientWidth) {
      setLabelInterval(durationsRef.current.clientWidth / Math.ceil(totalDuration / chartTimeInterval));
      setChartWidth(durationsRef.current.clientWidth);
    } else {
      setChartWidth(calcWidth);
      setLabelInterval(INTERVAL_LENGTH);
    }
  }, [totalDuration, chartTimeInterval, durationsRef]);

  React.useEffect(() => {
    const durationsView = durationsRef?.current;
    const labelsView = durationsLabelsRef?.current;
    if (durationsView && labelsView) {
      const handleScroll = e => {
        durationsView.scrollTo({
          left: durationsView.scrollLeft + e.deltaY,
          behavior: 'smooth'
        });
        labelsView.scrollTo({
          left: labelsView.scrollLeft + e.deltaY,
          behavior: 'smooth'
        });
      };

      durationsView.addEventListener('wheel', handleScroll);

      return () => durationsView.removeEventListener('wheel', handleScroll);
    }

    return () => {};
  }, [durationsRef, durationsLabelsRef]);

  React.useEffect(() => {
    const el = taskNamesRef.current;
    if (el) {
      const handleScroll = e => {
        const canvasView = durationsRef?.current;
        if (canvasView) {
          canvasView.scrollTo({
            top: e.srcElement.scrollTop
            // behavior: 'smooth'
          });
        }
      };

      el.addEventListener('scroll', handleScroll);

      return () => el.removeEventListener('scroll', handleScroll);
    }

    return () => {};
  }, [taskNamesRef, durationsRef]);

  const toggleNode = (id: string, scopeId: string) => {
    const searchNode = (nodes: dNode[]) => {
      if (!nodes || nodes.length === 0) {
        return;
      }
      for (let i = 0; i < nodes.length; i++) {
        const node = nodes[i];
        if (isStartNode(node) || isEndNode(node)) {
          continue;
        }
        if (node.id === id && node.scopedId === scopeId) {
          nodes[i].expanded = !nodes[i].expanded;
          return;
        }
        if (node.nodes.length > 0 && isExpanded(node)) {
          searchNode(node.nodes);
        }
      }
    };
    searchNode(originalNodes);
    setOriginalNodes([...originalNodes]);
  };

  const labels = React.useMemo(() => {
    const len = Math.ceil(totalDuration / chartTimeInterval);
    const lbs = len > 0 ? new Array(len).fill(0) : [];
    return lbs.map((_, idx) => {
      const time = moment.utc(new Date(startedAt.getTime() + idx * chartTimeInterval * 1000));
      return chartTimezone === TimeZone.UTC ? time.format('hh:mm:ss A') : time.local().format('hh:mm:ss A');
    });
  }, [chartTimezone, startedAt, chartTimeInterval, totalDuration]);

  return (
    <>
      <div className={styles.taskNames}>
        <Typography className={styles.taskNamesHeader}>Task Name</Typography>
        <TaskNames nodes={nodes} ref={taskNamesRef} onToggle={toggleNode} />
      </div>
      <div className={styles.taskDurations}>
        <div className={styles.taskDurationsLabelsView} ref={durationsLabelsRef}>
          <ChartHeader labels={labels} chartWidth={chartWidth} labelInterval={labelInterval} />
        </div>
        <div className={styles.taskDurationsView} ref={durationsRef}>
          <div className={styles.chartHeader}>
            <Bar options={getBarOptions(chartTimeInterval)} data={chartData} />
          </div>
        </div>
      </div>
    </>
  );
};
