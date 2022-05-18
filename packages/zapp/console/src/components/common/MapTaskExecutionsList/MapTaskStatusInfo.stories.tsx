import * as React from 'react';
import { ComponentStory, ComponentMeta } from '@storybook/react';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { mockExecution as mockTaskExecution } from 'models/Execution/__mocks__/mockTaskExecutionsData';
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
  taskExecution: mockTaskExecution,
  taskLogs: [
    // logs without URI should be black and not clickable
    { uri: '#', name: 'Kubernetes Logs #0-0' },
    { uri: '#', name: 'Kubernetes Logs #0-1' },
    { name: 'Kubernetes Logs #0-2' },
    { name: 'Kubernetes Logs #0-3' },
    { uri: '#', name: 'Kubernetes Logs #0-4' },
  ],
  phase: TaskExecutionPhase.QUEUED,
  selectedPhase: TaskExecutionPhase.QUEUED,
  onTaskSelected: () => {},
};

export const AllSpace = Template.bind({});
AllSpace.args = {
  taskExecution: mockTaskExecution,
  taskLogs: [],
  phase: TaskExecutionPhase.SUCCEEDED,
  onTaskSelected: () => {},
};
