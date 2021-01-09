import { render, waitFor } from '@testing-library/react';
import { basicPythonWorkflow } from 'mocks/data/fixtures/basicPythonWorkflow';
import { oneFailedTaskWorkflow } from 'mocks/data/fixtures/oneFailedTaskWorkflow';
import { insertFixture } from 'mocks/data/insertFixture';
import { notFoundError, unexpectedError } from 'mocks/errors';
import { mockServer } from 'mocks/server';
import {
    DomainIdentifierScope,
    Execution,
    executionSortFields,
    SortDirection,
    sortQueryKeys
} from 'models';
import * as React from 'react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { MemoryRouter } from 'react-router';
import { createTestQueryClient, disableQueryLogger, enableQueryLogger } from 'test/utils';
import { failedToLoadExecutionsString } from '../constants';
import { ProjectExecutions } from '../ProjectExecutions';

jest.mock('components/Executions/Tables/WorkflowExecutionsTable');

const defaultQueryParams = {
    [sortQueryKeys.direction]: SortDirection[SortDirection.DESCENDING],
    [sortQueryKeys.key]: executionSortFields.createdAt
};

describe('ProjectExecutions', () => {
    let basicPythonFixture: ReturnType<typeof basicPythonWorkflow.generate>;
    let failedTaskFixture: ReturnType<typeof oneFailedTaskWorkflow.generate>;
    let executions: Execution[];
    let scope: DomainIdentifierScope;
    let queryClient: QueryClient;

    beforeEach(() => {
        queryClient = createTestQueryClient();
        basicPythonFixture = basicPythonWorkflow.generate();
        failedTaskFixture = oneFailedTaskWorkflow.generate();
        insertFixture(mockServer, basicPythonFixture);
        insertFixture(mockServer, failedTaskFixture);
        executions = [
            basicPythonFixture.workflowExecutions.top.data,
            failedTaskFixture.workflowExecutions.top.data
        ];
        const { domain, project } = executions[0].id;
        scope = { domain, project };
        mockServer.insertWorkflowExecutionList(
            scope,
            executions,
            defaultQueryParams
        );
    });

    const renderView = () =>
        render(
            <QueryClientProvider client={queryClient}>
                <ProjectExecutions
                    projectId={scope.project}
                    domainId={scope.domain}
                />
            </QueryClientProvider>,
            { wrapper: MemoryRouter }
        );

    it('displays executions after successful load', async () => {
        const { getByText } = renderView();
        await waitFor(() => expect(getByText(executions[0].id.name)));
    });

    it('shows name of corresponding workflow in row items', async () => {
        const { getByText } = renderView();
        await waitFor(() =>
            expect(getByText(executions[0].closure.workflowId.name))
        );
    });

    describe('when initial load fails', () => {
        const errorMessage = 'Something went wrong.';
        // Disable react-query logger output to avoid a console.error
        // when the request fails.
        beforeEach(() => {
            disableQueryLogger();
            mockServer.insertWorkflowExecutionList(
                scope,
                unexpectedError(errorMessage),
                defaultQueryParams
            );
        });
        afterEach(() => {
            enableQueryLogger();
        });

        it('shows error message', async () => {
            const { getByText } = renderView();
            await waitFor(() => expect(getByText(failedToLoadExecutionsString)));
        });
    });
});
