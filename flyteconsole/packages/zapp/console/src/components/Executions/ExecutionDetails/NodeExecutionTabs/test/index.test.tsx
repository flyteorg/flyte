import { render } from '@testing-library/react';
import { useTabState } from 'components/hooks/useTabState';
import { extractTaskTemplates } from 'components/hooks/utils';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { createMockNodeExecutions } from 'models/Execution/__mocks__/mockNodeExecutionsData';
import { TaskType } from 'models/Task/constants';
import { createMockWorkflow } from 'models/__mocks__/workflowData';
import * as React from 'react';
import { mockExecution as mockTaskExecution } from 'models/Execution/__mocks__/mockTaskExecutionsData';
import { NodeExecutionTabs } from '../index';

const getMockNodeExecution = () => createMockNodeExecutions(1).executions[0];
const nodeExecution = getMockNodeExecution();
const workflow = createMockWorkflow('SampleWorkflow');
const taskTemplate = { ...extractTaskTemplates(workflow)[0], type: TaskType.ARRAY };
const phase = TaskExecutionPhase.SUCCEEDED;

jest.mock('components/hooks/useTabState');

describe('NodeExecutionTabs', () => {
  const mockUseTabState = useTabState as jest.Mock<any>;
  mockUseTabState.mockReturnValue({ onChange: jest.fn(), value: 'executions' });
  describe('with map tasks', () => {
    it('should display proper tab name when it was provided and shouldShow is TRUE', () => {
      const { queryByText, queryAllByRole } = render(
        <NodeExecutionTabs
          nodeExecution={nodeExecution}
          selectedTaskExecution={{ ...mockTaskExecution, taskIndex: 0 }}
          phase={phase}
          taskTemplate={taskTemplate}
          onTaskSelected={jest.fn()}
        />,
      );
      expect(queryAllByRole('tab')).toHaveLength(4);
      expect(queryByText('Executions')).toBeInTheDocument();
    });

    it('should display proper tab name when it was provided and shouldShow is FALSE', () => {
      const { queryByText, queryAllByRole } = render(
        <NodeExecutionTabs
          nodeExecution={nodeExecution}
          selectedTaskExecution={null}
          phase={phase}
          taskTemplate={taskTemplate}
          onTaskSelected={jest.fn()}
        />,
      );

      expect(queryAllByRole('tab')).toHaveLength(4);
      expect(queryByText('Map Execution')).toBeInTheDocument();
    });
  });

  describe('without map tasks', () => {
    it('should display proper tab name when mapTask was not provided', () => {
      const { queryAllByRole, queryByText } = render(
        <NodeExecutionTabs
          nodeExecution={nodeExecution}
          selectedTaskExecution={null}
          onTaskSelected={jest.fn()}
        />,
      );

      expect(queryAllByRole('tab')).toHaveLength(3);
      expect(queryByText('Executions')).toBeInTheDocument();
    });
  });
});
