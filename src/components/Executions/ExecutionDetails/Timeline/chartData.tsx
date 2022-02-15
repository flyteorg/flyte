import { durationToMilliseconds, timestampToDate } from 'common/utils';
import { getNodeExecutionPhaseConstants } from 'components/Executions/utils';
import { NodeExecutionPhase } from 'models/Execution/enums';
import { dNode } from 'models/Graph/types';
import * as React from 'react';

interface DataProps {
  nodes: dNode[];
}

export const useChartDurationData = (props: DataProps) => {
  const colorData = React.useMemo(() => {
    const definedExecutions = props.nodes.map(
      ({ execution }) =>
        getNodeExecutionPhaseConstants(execution?.closure.phase ?? NodeExecutionPhase.UNDEFINED).badgeColor
    );
    return definedExecutions;
  }, [props.nodes]);

  const startedAt = React.useMemo(() => {
    if (props.nodes.length === 0 || !props.nodes[0].execution?.closure.startedAt) {
      return new Date();
    }
    return timestampToDate(props.nodes[0].execution?.closure.startedAt);
  }, [props.nodes]);

  const stackedData = React.useMemo(() => {
    let undefinedStart = 0;
    for (const node of props.nodes) {
      const exec = node.execution;
      if (exec?.closure.startedAt) {
        const startedTime = timestampToDate(exec?.closure.startedAt).getTime();
        const absoluteDuration =
          startedTime -
          startedAt.getTime() +
          (exec?.closure.duration ? durationToMilliseconds(exec?.closure.duration) : Date.now() - startedTime);
        if (absoluteDuration > undefinedStart) {
          undefinedStart = absoluteDuration;
        }
      }
    }
    undefinedStart = undefinedStart / 1000;

    const definedExecutions = props.nodes.map(({ execution }) =>
      execution?.closure.startedAt
        ? (timestampToDate(execution?.closure.startedAt).getTime() - startedAt.getTime()) / 1000
        : 0
    );

    return definedExecutions;
  }, [props.nodes, startedAt]);

  // Divide by 1000 to calculate all duration data be second based.
  const durationData = React.useMemo(() => {
    const definedExecutions = props.nodes.map(node => {
      const exec = node.execution;
      if (!exec) return 0;
      if (exec.closure.phase === NodeExecutionPhase.RUNNING) {
        if (!exec.closure.startedAt) {
          return 0;
        }
        return (Date.now() - timestampToDate(exec.closure.startedAt).getTime()) / 1000;
      }
      if (!exec.closure.duration) {
        return 0;
      }
      return durationToMilliseconds(exec.closure.duration) / 1000;
    });
    return definedExecutions;
  }, [props.nodes]);

  const totalDuration = React.useMemo(() => {
    const durations = durationData.map((duration, idx) => duration + stackedData[idx]);
    return Math.max(...durations);
  }, [durationData, stackedData]);

  const stackedColorData = React.useMemo(() => {
    return durationData.map(duration => {
      return duration === 0 ? '#4AE3AE40' : 'rgba(0, 0, 0, 0)';
    });
  }, [durationData]);

  const chartData = React.useMemo(() => {
    return {
      labels: durationData.map(() => ''),
      datasets: [
        {
          data: stackedData,
          backgroundColor: stackedColorData,
          barPercentage: 1,
          borderWidth: 0,
          datalabels: {
            labels: {
              title: null
            }
          }
        },
        {
          data: durationData.map(duration => {
            return duration === -1 ? 10 : duration === 0 ? 0.5 : duration;
          }),
          backgroundColor: colorData,
          barPercentage: 1,
          borderWidth: 0,
          datalabels: {
            color: '#292936' as const,
            align: 'start' as const,
            formatter: function(value, context) {
              if (durationData[context.dataIndex] === -1) {
                return '';
              }
              return Math.round(value) + 's';
            }
          }
        }
      ]
    };
  }, [durationData, stackedData, colorData, stackedColorData]);

  return {
    startedAt,
    totalDuration,
    durationLength: durationData.length,
    chartData
  };
};
