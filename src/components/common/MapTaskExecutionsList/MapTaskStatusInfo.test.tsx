import { fireEvent, render, waitFor } from '@testing-library/react';
import { noLogsFoundString } from 'components/Executions/constants';
import { getTaskExecutionPhaseConstants } from 'components/Executions/utils';
import { TaskExecutionPhase } from 'models/Execution/enums';
import * as React from 'react';

import { MapTaskStatusInfo } from './MapTaskStatusInfo';

const taskLogs = [
  { uri: '#', name: 'Kubernetes Logs #0-0' },
  { uri: '#', name: 'Kubernetes Logs #0-1' },
  { uri: '#', name: 'Kubernetes Logs #0-2' },
];

describe('MapTaskStatusInfo', () => {
  it('Phase and amount of links rendered correctly', async () => {
    const status = TaskExecutionPhase.RUNNING;
    const phaseData = getTaskExecutionPhaseConstants(status);

    const { queryByText, getByTitle } = render(
      <MapTaskStatusInfo taskLogs={taskLogs} status={status} expanded={false} />,
    );

    expect(queryByText(phaseData.text)).toBeInTheDocument();
    expect(queryByText(`x${taskLogs.length}`)).toBeInTheDocument();
    expect(queryByText('Logs')).not.toBeInTheDocument();

    // Expand item - see logs section
    const buttonEl = getByTitle('Expand row');
    fireEvent.click(buttonEl);
    await waitFor(() => {
      expect(queryByText(taskLogs[0].name)).toBeInTheDocument();
    });
  });

  it('Phase with no links show proper texts when opened', () => {
    const status = TaskExecutionPhase.ABORTED;
    const phaseData = getTaskExecutionPhaseConstants(status);

    const { queryByText } = render(
      <MapTaskStatusInfo taskLogs={[]} status={status} expanded={true} />,
    );

    expect(queryByText(phaseData.text)).toBeInTheDocument();
    expect(queryByText(`x0`)).toBeInTheDocument();
    expect(queryByText(noLogsFoundString)).toBeInTheDocument();
  });
});
