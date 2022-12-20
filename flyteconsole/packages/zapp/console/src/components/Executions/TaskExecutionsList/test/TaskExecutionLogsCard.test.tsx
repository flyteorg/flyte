import { render } from '@testing-library/react';
import * as React from 'react';
import { mockExecution as mockTaskExecution } from 'models/Execution/__mocks__/mockTaskExecutionsData';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { noLogsFoundString } from 'components/Executions/constants';
import { getPhaseConstants } from 'components/Executions/ExecutionStatusBadge';
import { TaskExecutionLogsCard } from '../TaskExecutionLogsCard';
import { formatRetryAttempt } from '../utils';

const headerText = formatRetryAttempt(0);
const taskLogs = [{ uri: '#', name: 'Kubernetes Logs #0-0' }];
const phase = TaskExecutionPhase.SUCCEEDED;

describe('TaskExecutionLogsCard', () => {
  it('should render card with logs provided', () => {
    const { queryByText } = render(
      <TaskExecutionLogsCard
        taskExecution={mockTaskExecution}
        headerText={headerText}
        logs={taskLogs}
        phase={phase}
      />,
    );
    const { text } = getPhaseConstants('task', phase);

    expect(queryByText(headerText)).toBeInTheDocument();
    expect(queryByText(text)).toBeInTheDocument();
    expect(queryByText(taskLogs[0].name)).toBeInTheDocument();
  });

  it('should render card with no logs found string', () => {
    const { queryByText } = render(
      <TaskExecutionLogsCard
        taskExecution={mockTaskExecution}
        headerText={headerText}
        logs={[]}
        phase={phase}
      />,
    );

    const { text } = getPhaseConstants('task', phase);

    expect(queryByText(headerText)).toBeInTheDocument();
    expect(queryByText(text)).toBeInTheDocument();
    expect(queryByText(noLogsFoundString)).toBeInTheDocument();
  });
});
