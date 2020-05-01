import { render, waitFor } from '@testing-library/react';
import { noExecutionsFoundString } from 'common/constants';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { APIContext } from 'components/data/apiContext';
import {
    listTaskExecutions,
    NodeExecution,
    SortDirection,
    taskSortFields
} from 'models';
import { mockNodeExecutionResponse } from 'models/Execution/__mocks__/mockNodeExecutionsData';
import * as React from 'react';
import { TaskExecutionsList } from '../TaskExecutionsList';

describe('TaskExecutionsList', () => {
    let nodeExecution: NodeExecution;
    let mockListTaskExecutions: jest.Mock<ReturnType<
        typeof listTaskExecutions
    >>;

    const renderList = () =>
        render(
            <APIContext.Provider
                value={mockAPIContextValue({
                    listTaskExecutions: mockListTaskExecutions
                })}
            >
                <TaskExecutionsList nodeExecution={nodeExecution} />
            </APIContext.Provider>
        );
    beforeEach(() => {
        nodeExecution = { ...mockNodeExecutionResponse } as NodeExecution;
        mockListTaskExecutions = jest.fn().mockResolvedValue({ entities: [] });
    });

    it('Renders message when no task executions exist', async () => {
        const { queryByText } = renderList();
        await waitFor(() => {});
        expect(mockListTaskExecutions).toHaveBeenCalled();
        expect(queryByText(noExecutionsFoundString)).toBeInTheDocument();
    });

    it('Requests items in correct order', async () => {
        renderList();
        await waitFor(() => {});
        expect(mockListTaskExecutions).toHaveBeenCalledWith(
            expect.anything(),
            expect.objectContaining({
                sort: {
                    key: taskSortFields.createdAt,
                    direction: SortDirection.ASCENDING
                }
            })
        );
    });
});
