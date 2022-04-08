import { ComponentMeta, ComponentStory } from '@storybook/react';
import * as React from 'react';
import { NodeExecutionPhase } from 'models/Execution/enums';
import { TimelineChart } from '.';

const barItems = [
  { phase: NodeExecutionPhase.FAILED, startOffsetSec: 0, durationSec: 5, isFromCache: false },
  { phase: NodeExecutionPhase.SUCCEEDED, startOffsetSec: 10, durationSec: 2, isFromCache: true },
  { phase: NodeExecutionPhase.SUCCEEDED, startOffsetSec: 0, durationSec: 1, isFromCache: true },
  { phase: NodeExecutionPhase.RUNNING, startOffsetSec: 0, durationSec: 10, isFromCache: false },
  { phase: NodeExecutionPhase.UNDEFINED, startOffsetSec: 15, durationSec: 25, isFromCache: false },
  { phase: NodeExecutionPhase.SUCCEEDED, startOffsetSec: 7, durationSec: 20, isFromCache: false },
];

export default {
  title: 'Workflow/Timeline',
  component: TimelineChart,
} as ComponentMeta<typeof TimelineChart>;

const Template: ComponentStory<typeof TimelineChart> = (args) => <TimelineChart {...args} />;
export const BarSection = Template.bind({});
BarSection.args = {
  items: barItems,
  chartTimeIntervalSec: 1,
};
