import { ComponentMeta, ComponentStory } from '@storybook/react';
import * as React from 'react';
import { NodeExecutionPhase } from 'models/Execution/enums';
import { BarItemData } from './utils';
import { BarChart } from '.';

const phaseEnumTyping = {
  options: Object.values(NodeExecutionPhase),
  mapping: Object.values(NodeExecutionPhase),
  control: {
    type: 'select',
    labels: Object.keys(NodeExecutionPhase),
  },
};

interface SingleItemProps extends BarItemData {
  chartTimeIntervalSec: number;
}

/**
 * This is a fake storybook only component, to allow ease experimentation whith single bar item
 */
const SingleBarItem = (props: SingleItemProps) => {
  const items = [props];
  return <BarChart items={items} chartTimeIntervalSec={props.chartTimeIntervalSec} />;
};

export default {
  title: 'Workflow/Timeline',
  component: SingleBarItem,
  // 👇 Creates specific argTypes
  argTypes: {
    phase: phaseEnumTyping,
  },
} as ComponentMeta<typeof SingleBarItem>;

const TemplateSingleItem: ComponentStory<typeof SingleBarItem> = (args) => (
  <SingleBarItem {...args} />
);

export const BarChartSingleItem = TemplateSingleItem.bind({});
// const phaseDataSingle = generateChartData([barItems[0]]);
BarChartSingleItem.args = {
  phase: NodeExecutionPhase.ABORTED,
  startOffsetSec: 15,
  durationSec: 30,
  isFromCache: false,
  chartTimeIntervalSec: 5,
};
