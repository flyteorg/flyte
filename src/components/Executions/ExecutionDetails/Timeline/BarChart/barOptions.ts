import { Chart as ChartJS, registerables, Tooltip } from 'chart.js';
import ChartDataLabels from 'chartjs-plugin-datalabels';

ChartJS.register(...registerables, ChartDataLabels);

// Create positioner to put tooltip at cursor position
Tooltip.positioners.cursor = function (_chartElements, coordinates) {
  return coordinates;
};

export const getBarOptions = (chartTimeIntervalSec: number, tooltipLabels: string[][]) => {
  return {
    animation: false as const,
    indexAxis: 'y' as const,
    elements: {
      bar: {
        borderWidth: 2,
      },
    },
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        display: false,
      },
      title: {
        display: false,
      },
      tooltip: {
        // Setting up tooltip: https://www.chartjs.org/docs/latest/configuration/tooltip.html
        position: 'cursor',
        filter: function (tooltipItem) {
          // no tooltip for offsets
          return tooltipItem.datasetIndex === 1;
        },
        callbacks: {
          label: function (context) {
            const index = context.dataIndex;
            return tooltipLabels[index] ?? '';
          },
        },
      },
    },
    scales: {
      x: {
        format: Intl.DateTimeFormat,
        position: 'top' as const,
        ticks: {
          display: false,
          autoSkip: false,
          stepSize: chartTimeIntervalSec,
        },
        stacked: true,
      },
      y: {
        stacked: true,
      },
    },
  };
};
