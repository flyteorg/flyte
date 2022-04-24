import { render } from '@testing-library/react';
import { dashedValueString } from 'common/constants';
import { Protobuf } from 'flyteidl';
import { Execution, WorkflowExecutionIdentifier } from 'models/Execution/types';
import { createMockExecution } from 'models/__mocks__/executionsData';
import * as React from 'react';
import { MemoryRouter } from 'react-router';
import { Routes } from 'routes/routes';
import { ExecutionMetadataLabels } from '../constants';
import { ExecutionMetadata } from '../ExecutionMetadata';

const clusterTestId = `metadata-${ExecutionMetadataLabels.cluster}`;
const startTimeTestId = `metadata-${ExecutionMetadataLabels.time}`;
const durationTestId = `metadata-${ExecutionMetadataLabels.duration}`;
const interruptibleTestId = `metadata-${ExecutionMetadataLabels.interruptible}`;

jest.mock('models/Launch/api', () => ({
  getLaunchPlan: jest.fn(() => Promise.resolve({ spec: {} })),
}));

describe('ExecutionMetadata', () => {
  let execution: Execution;
  beforeEach(() => {
    execution = createMockExecution();
  });

  const renderMetadata = () =>
    render(
      <MemoryRouter>
        <ExecutionMetadata execution={execution} />
      </MemoryRouter>,
    );

  it('shows cluster name if available', () => {
    const { getByTestId } = renderMetadata();

    expect(execution.spec.metadata.systemMetadata?.executionCluster).toBeDefined();
    expect(getByTestId(clusterTestId)).toHaveTextContent(
      execution.spec.metadata.systemMetadata!.executionCluster!,
    );
  });

  it('shows empty string for cluster if no metadata', () => {
    delete execution.spec.metadata.systemMetadata;
    const { getByTestId } = renderMetadata();
    expect(getByTestId(clusterTestId)).toHaveTextContent(dashedValueString);
  });

  it('shows empty string for cluster if no cluster name', () => {
    delete execution.spec.metadata.systemMetadata?.executionCluster;
    const { getByTestId } = renderMetadata();
    expect(getByTestId(clusterTestId)).toHaveTextContent(dashedValueString);
  });

  it('shows empty string for start time if not available', () => {
    delete execution.closure.startedAt;
    const { getByTestId } = renderMetadata();
    expect(getByTestId(startTimeTestId)).toHaveTextContent(dashedValueString);
  });

  it('shows empty string for duration if not available', () => {
    delete execution.closure.duration;
    const { getByTestId } = renderMetadata();
    expect(getByTestId(durationTestId)).toHaveTextContent(dashedValueString);
  });

  it('shows reference execution if it exists', () => {
    const referenceExecution: WorkflowExecutionIdentifier = {
      project: 'project',
      domain: 'domain',
      name: '123abc',
    };
    execution.spec.metadata.referenceExecution = referenceExecution;
    const { getByText } = renderMetadata();
    expect(getByText(referenceExecution.name)).toHaveAttribute(
      'href',
      Routes.ExecutionDetails.makeUrl(referenceExecution),
    );
  });

  it('shows true if execution was marked as interruptible', () => {
    execution.spec.interruptible = Protobuf.BoolValue.create({ value: true });
    const { getByTestId } = renderMetadata();
    expect(getByTestId(interruptibleTestId)).toHaveTextContent('true');
  });

  it('shows false if execution was not marked as interruptible', () => {
    execution.spec.interruptible = Protobuf.BoolValue.create({ value: false });
    const { getByTestId } = renderMetadata();
    expect(getByTestId(interruptibleTestId)).toHaveTextContent('false');
  });

  it('shows dashes if no interruptible value is found in execution spec', () => {
    delete execution.spec.interruptible;
    const { getByTestId } = renderMetadata();
    expect(getByTestId(interruptibleTestId)).toHaveTextContent(dashedValueString);
  });
});
