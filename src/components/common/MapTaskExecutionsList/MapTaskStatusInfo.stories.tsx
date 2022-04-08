import * as React from 'react';
import { ComponentStory, ComponentMeta } from '@storybook/react';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { MapTaskStatusInfo } from './MapTaskStatusInfo';
import { PanelViewDecorator } from '../__stories__/Decorators';

export default {
  title: 'Task/NodeExecutionTabs/MapTaskStatusInfo',
  component: MapTaskStatusInfo,
  parameters: { actions: { argTypesRegex: 'toggleExpanded' } },
} as ComponentMeta<typeof MapTaskStatusInfo>;

const Template: ComponentStory<typeof MapTaskStatusInfo> = (args) => (
  <MapTaskStatusInfo {...args} />
);

export const Default = Template.bind({});
Default.decorators = [(Story) => PanelViewDecorator(Story)];
Default.args = {
  taskLogs: [
    { uri: '#', name: 'Kubernetes Logs #0-0' },
    { uri: '#', name: 'Kubernetes Logs #0-1' },
    { uri: '#', name: 'Kubernetes Logs #0-2' },
    { uri: '#', name: 'Kubernetes Logs #0-3' },
    { uri: '#', name: 'Kubernetes Logs #0-4' },
  ],
  status: TaskExecutionPhase.QUEUED,
  expanded: true,
};

export const AllSpace = Template.bind({});
AllSpace.args = {
  taskLogs: [],
  status: TaskExecutionPhase.SUCCEEDED,
};
