import { render } from '@testing-library/react';
import * as React from 'react';
import { TaskExecutionsListItem } from '../TaskExecutionsListItem';
import { MockMapTaskExecution } from '../TaskExecutions.mocks';

jest.mock('../TaskExecutionLogsCard.tsx', () => ({
  TaskExecutionLogsCard: jest.fn(({ children }) => <div data-testid="logs-card">{children}</div>),
}));

describe('TaskExecutionsListItem', () => {
  it('should render execution logs card', () => {
    const { queryByTestId } = render(
      <TaskExecutionsListItem taskExecution={MockMapTaskExecution} />,
    );
    expect(queryByTestId('logs-card')).toBeInTheDocument();
  });
});
