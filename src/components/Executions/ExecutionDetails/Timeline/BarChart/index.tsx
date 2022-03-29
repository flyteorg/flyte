import * as React from 'react';
import { Bar } from 'react-chartjs-2';
import { getBarOptions } from './barOptions';
import { BarItemData, generateChartData, getChartData } from './utils';

interface BarChartProps {
  items: BarItemData[];
  chartTimeIntervalSec: number;
}

export const BarChart = (props: BarChartProps) => {
  const phaseData = generateChartData(props.items);

  return (
    <Bar
      options={getBarOptions(props.chartTimeIntervalSec, phaseData.tooltipLabel) as any}
      data={getChartData(phaseData)}
    />
  );
};
