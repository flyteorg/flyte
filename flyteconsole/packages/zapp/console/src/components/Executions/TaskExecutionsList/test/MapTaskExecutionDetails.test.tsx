import { render } from '@testing-library/react';
import * as React from 'react';
import { MapTaskExecutionDetails } from '../MapTaskExecutionDetails';
import { MockMapTaskExecution } from '../TaskExecutions.mocks';

jest.mock('../TaskExecutionLogsCard.tsx', () => ({
  TaskExecutionLogsCard: jest.fn(({ children }) => <div data-testid="logs-card">{children}</div>),
}));

describe('MapTaskExecutionDetails', () => {
  it('should render list with 1 execution attempt', () => {
    const { queryAllByTestId } = render(
      <MapTaskExecutionDetails taskExecution={{ ...MockMapTaskExecution, taskIndex: 1 }} />,
    );
    const logsCards = queryAllByTestId('logs-card');
    expect(logsCards).toHaveLength(1);
    logsCards.forEach((card) => {
      expect(card).toBeInTheDocument();
    });
  });

  it('should render list with 2 execution attempts', () => {
    const { queryAllByTestId } = render(
      <MapTaskExecutionDetails taskExecution={{ ...MockMapTaskExecution, taskIndex: 3 }} />,
    );
    const logsCards = queryAllByTestId('logs-card');
    expect(logsCards).toHaveLength(2);
    logsCards.forEach((card) => {
      expect(card).toBeInTheDocument();
    });
  });
});
