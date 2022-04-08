import * as React from 'react';
import { ComponentStory, ComponentMeta } from '@storybook/react';
import { Core } from 'flyteidl';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { PanelViewDecorator } from 'components/common/__stories__/Decorators';
import { TaskType } from 'models/Task/constants';
import { TaskExecutionsListContent } from './TaskExecutionsList';
import {
  getMockMapTaskLogItem,
  MockMapTaskExecution,
  MockPythonTaskExecution,
  MockTaskExceutionLog,
} from './TaskExecutions.mocks';

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

export const MapTaskOldView = Template.bind({});
MapTaskOldView.args = {
  taskExecutions: [
    {
      ...MockPythonTaskExecution,
      closure: {
        ...MockPythonTaskExecution.closure,
        taskType: TaskType.ARRAY,
        logs: new Array(5).fill(MockTaskExceutionLog),
      },
    },
  ],
};

export const MapTaskBase = Template.bind({});
MapTaskBase.args = {
  taskExecutions: [MockMapTaskExecution],
};

export const MapTaskEverything = Template.bind({});
MapTaskEverything.decorators = [(Story) => PanelViewDecorator(Story)];
MapTaskEverything.args = {
  taskExecutions: [
    {
      ...MockMapTaskExecution,
      closure: {
        ...MockMapTaskExecution.closure,
        logs: [MockTaskExceutionLog],
        error: {
          code: '666',
          message: 'Fake error occured, if you know what I mean',
          errorUri: '#',
          kind: Core.ExecutionError.ErrorKind.USER,
        },
        metadata: {
          ...MockMapTaskExecution.closure.metadata,
          externalResources: [
            getMockMapTaskLogItem(TaskExecutionPhase.ABORTED, true, 1),
            getMockMapTaskLogItem(TaskExecutionPhase.ABORTED, true, 5),
            getMockMapTaskLogItem(TaskExecutionPhase.FAILED, true, 2),
            getMockMapTaskLogItem(TaskExecutionPhase.FAILED, true, 2, 1),
            getMockMapTaskLogItem(TaskExecutionPhase.FAILED, true, 10),
            getMockMapTaskLogItem(TaskExecutionPhase.FAILED, true, 10, 1),
            getMockMapTaskLogItem(TaskExecutionPhase.INITIALIZING, false, 3),
            getMockMapTaskLogItem(TaskExecutionPhase.INITIALIZING, false, 12),
            getMockMapTaskLogItem(TaskExecutionPhase.WAITING_FOR_RESOURCES, false, 4),
            getMockMapTaskLogItem(TaskExecutionPhase.QUEUED, true, 6),
            getMockMapTaskLogItem(TaskExecutionPhase.QUEUED, true, 7),
            getMockMapTaskLogItem(TaskExecutionPhase.RUNNING, true, 8),
            getMockMapTaskLogItem(TaskExecutionPhase.UNDEFINED, false, 9),
            getMockMapTaskLogItem(TaskExecutionPhase.SUCCEEDED, true),
            getMockMapTaskLogItem(TaskExecutionPhase.SUCCEEDED, true, 2, 2),
            getMockMapTaskLogItem(TaskExecutionPhase.SUCCEEDED, true, 11),
          ],
        },
      },
    },
    { ...MockMapTaskExecution, id: { ...MockMapTaskExecution.id, retryAttempt: 1 } },
  ],
};
