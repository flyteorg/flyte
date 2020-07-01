import { render } from '@testing-library/react';
import { dashedValueString } from 'common/constants';
import { Execution } from 'models';
import { createMockExecution } from 'models/__mocks__/executionsData';
import * as React from 'react';
import { ExecutionMetadataLabels } from '../constants';
import { ExecutionMetadata } from '../ExecutionMetadata';

const clusterTestId = `metadata-${ExecutionMetadataLabels.cluster}`;
const startTimeTestId = `metadata-${ExecutionMetadataLabels.time}`;
const durationTestId = `metadata-${ExecutionMetadataLabels.duration}`;

describe('ExecutionMetadata', () => {
    let execution: Execution;
    beforeEach(() => {
        execution = createMockExecution();
    });

    const renderMetadata = () =>
        render(<ExecutionMetadata execution={execution} />);

    it('shows cluster name if available', () => {
        const { getByTestId } = renderMetadata();

        expect(
            execution.spec.metadata.systemMetadata?.executionCluster
        ).toBeDefined();
        expect(getByTestId(clusterTestId)).toHaveTextContent(
            execution.spec.metadata.systemMetadata!.executionCluster!
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
        expect(getByTestId(startTimeTestId)).toHaveTextContent(
            dashedValueString
        );
    });

    it('shows empty string for duration if not available', () => {
        delete execution.closure.duration;
        const { getByTestId } = renderMetadata();
        expect(getByTestId(durationTestId)).toHaveTextContent(
            dashedValueString
        );
    });
});
