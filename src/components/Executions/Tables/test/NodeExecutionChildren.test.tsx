import { render, waitFor } from '@testing-library/react';
import { noExecutionsFoundString } from 'common/constants';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { APIContext } from 'components/data/apiContext';
import {
    DetailedNodeExecution,
    NodeExecutionDisplayType
} from 'components/Executions/types';
import {
    listTaskExecutions,
    NodeExecution,
    SortDirection,
    taskSortFields
} from 'models';
import { mockNodeExecutionResponse } from 'models/Execution/__mocks__/mockNodeExecutionsData';
import * as React from 'react';
import { NodeExecutionChildren } from '../NodeExecutionChildren';

describe('NodeExecutionChildren', () => {
    let nodeExecution: DetailedNodeExecution;
    let mockListTaskExecutions: jest.Mock<ReturnType<
        typeof listTaskExecutions
    >>;

    const renderChildren = () =>
        render(
            <APIContext.Provider
                value={mockAPIContextValue({
                    listTaskExecutions: mockListTaskExecutions
                })}
            >
                <NodeExecutionChildren execution={nodeExecution} />
            </APIContext.Provider>
        );
    beforeEach(() => {
        nodeExecution = {
            ...(mockNodeExecutionResponse as NodeExecution),
            displayId: 'testNode',
            displayType: NodeExecutionDisplayType.PythonTask,
            cacheKey: 'abcdefg'
        };
        mockListTaskExecutions = jest.fn().mockResolvedValue({ entities: [] });
    });

    it('Renders message when no task executions exist', async () => {
        const { queryByText } = renderChildren();
        await waitFor(() => {});
        expect(mockListTaskExecutions).toHaveBeenCalled();
        expect(queryByText(noExecutionsFoundString)).toBeInTheDocument();
    });

    it('Requests items in correct order', async () => {
        renderChildren();
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
