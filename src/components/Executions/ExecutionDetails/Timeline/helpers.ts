import { isEndNode, isExpanded, isStartNode } from 'components/WorkflowGraph/utils';
import { dNode } from 'models/Graph/types';

export const TimeZone = {
  Local: 'local',
  UTC: 'utc'
};

export function convertToPlainNodes(nodes: dNode[], level = 0): dNode[] {
  const result: dNode[] = [];
  if (!nodes || nodes.length === 0) {
    return result;
  }
  nodes.forEach(node => {
    if (isStartNode(node) || isEndNode(node)) {
      return;
    }
    result.push({ ...node, level });
    if (node.nodes.length > 0 && isExpanded(node)) {
      result.push(...convertToPlainNodes(node.nodes, level + 1));
    }
  });
  return result;
}

export const getBarOptions = (chartTimeInterval: number) => {
  return {
    animation: false as const,
    indexAxis: 'y' as const,
    elements: {
      bar: {
        borderWidth: 2
      }
    },
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        display: false
      },
      title: {
        display: false
      },
      tooltip: {
        filter: function(tooltipItem) {
          return tooltipItem.datasetIndex === 1;
        }
      }
    },
    scales: {
      x: {
        format: Intl.DateTimeFormat,
        position: 'top' as const,
        ticks: {
          display: false,
          autoSkip: false,
          stepSize: chartTimeInterval
        },
        stacked: true
      },
      y: {
        stacked: true
      }
    }
  };
};
