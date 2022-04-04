import * as React from 'react';
import { ComponentStory, ComponentMeta } from '@storybook/react';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { PanelViewDecorator } from 'components/common/__stories__/Decorators';
import { TaskExecutionsListContent } from './TaskExecutionsList';
import { MockPythonTaskExecution } from './TaskExecutions.mocks';

export default {
  title: 'Task/NodeExecutionTabs',
  component: TaskExecutionsListContent,
} as ComponentMeta<typeof TaskExecutionsListContent>;

// 👇 We create a “template” of how args map to rendering
const Template: ComponentStory<typeof TaskExecutionsListContent> = (args) => (
  <TaskExecutionsListContent {...args} />
);

export const PythonTaskExecution = Template.bind({});
PythonTaskExecution.args = { taskExecutions: [MockPythonTaskExecution] };

export const PythonTaskWithRetry = Template.bind({});
PythonTaskWithRetry.decorators = [(Story) => PanelViewDecorator(Story)];
PythonTaskWithRetry.args = {
  taskExecutions: [
    {
      ...MockPythonTaskExecution,
      closure: { ...MockPythonTaskExecution.closure, phase: TaskExecutionPhase.FAILED },
    },
    { ...MockPythonTaskExecution, id: { ...MockPythonTaskExecution.id, retryAttempt: 1 } },
  ],
};
