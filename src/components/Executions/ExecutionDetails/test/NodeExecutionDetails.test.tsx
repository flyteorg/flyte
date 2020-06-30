import { render, waitFor } from '@testing-library/react';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { APIContext } from 'components/data/apiContext';
import {
    DetailedNodeExecution,
    NodeExecutionDisplayType
} from 'components/Executions/types';
import { createMockNodeExecutions } from 'models/Execution/__mocks__/mockNodeExecutionsData';
import { listTaskExecutions } from 'models/Execution/api';
import * as React from 'react';
import { NodeExecutionDetails } from '../NodeExecutionDetails';

describe('NodeExecutionDetails', () => {
    let execution: DetailedNodeExecution;
    let mockListTaskExecutions: jest.Mock<ReturnType<
        typeof listTaskExecutions
    >>;
    beforeEach(() => {
        const { executions } = createMockNodeExecutions(1);
        execution = {
            ...executions[0],
            displayType: NodeExecutionDisplayType.PythonTask,
            displayId: 'com.flyte.testTask',
            cacheKey: 'abcdefg'
        };
        mockListTaskExecutions = jest.fn().mockResolvedValue({ entities: [] });
    });

    const renderComponent = () =>
        render(
            <APIContext.Provider
                value={mockAPIContextValue({
                    listTaskExecutions: mockListTaskExecutions
                })}
            >
                <NodeExecutionDetails execution={execution} />
            </APIContext.Provider>
        );

    it('renders displayId', async () => {
        const { queryByText } = renderComponent();
        await waitFor(() => {});
        expect(queryByText(execution.displayId)).toBeInTheDocument();
    });
});
