import ThemeProvider from '@material-ui/styles/ThemeProvider';
import { render } from '@testing-library/react';
import { muiTheme } from 'components/Theme/muiTheme';
import * as React from 'react';
import { mockExecution as mockTaskExecution } from 'models/Execution/__mocks__/mockTaskExecutionsData';

import { TaskNameList } from '../TaskNameList';

const taskLogs = [
  { uri: '#', name: 'Kubernetes Logs #0-0' },
  { uri: '#', name: 'Kubernetes Logs #0-1' },
  { uri: '#', name: 'Kubernetes Logs #0-2' },
];

const taskLogsWithoutUri = [
  { name: 'Kubernetes Logs #0-0' },
  { name: 'Kubernetes Logs #0-1' },
  { name: 'Kubernetes Logs #0-2' },
];

describe('TaskNameList', () => {
  it('should render log names in color if they have URI', async () => {
    const { queryAllByTestId } = render(
      <ThemeProvider theme={muiTheme}>
        <TaskNameList
          logs={taskLogs}
          taskExecution={mockTaskExecution}
          onTaskSelected={jest.fn()}
        />
      </ThemeProvider>,
    );

    const logs = queryAllByTestId('map-task-log');
    expect(logs).toHaveLength(3);
    logs.forEach((log) => {
      expect(log).toBeInTheDocument();
      expect(log).toHaveStyle({ color: '#8B37FF' });
    });
  });

  it('should render log names in black if they have URI', () => {
    const { queryAllByTestId } = render(
      <ThemeProvider theme={muiTheme}>
        <TaskNameList
          logs={taskLogsWithoutUri}
          taskExecution={mockTaskExecution}
          onTaskSelected={jest.fn()}
        />
      </ThemeProvider>,
    );

    const logs = queryAllByTestId('map-task-log');
    expect(logs).toHaveLength(3);
    logs.forEach((log) => {
      expect(log).toBeInTheDocument();
      expect(log).toHaveStyle({ color: '#292936' });
    });
  });
});
