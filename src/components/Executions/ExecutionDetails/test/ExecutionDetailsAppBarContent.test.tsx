import {
    fireEvent,
    render,
    RenderResult,
    waitFor
} from '@testing-library/react';
import { labels as commonLabels } from 'components/common/constants';
import {
    ExecutionContext,
    ExecutionContextData
} from 'components/Executions/contexts';
import { Execution, Identifier, ResourceType } from 'models';
import { createMockExecution } from 'models/__mocks__/executionsData';
import { WorkflowExecutionPhase } from 'models/Execution/enums';
import * as React from 'react';
import { MemoryRouter } from 'react-router';
import { Routes } from 'routes';
import { delayedPromise, DelayedPromiseResult } from 'test/utils';
import { backLinkTitle, executionActionStrings } from '../constants';
import { ExecutionDetailsAppBarContent } from '../ExecutionDetailsAppBarContent';

jest.mock('components/Navigation/NavBarContent', () => ({
    NavBarContent: ({ children }: React.Props<any>) => children
}));

describe('ExecutionDetailsAppBarContent', () => {
    let execution: Execution;
    let executionContext: ExecutionContextData;
    let mockTerminateExecution: jest.Mock<Promise<void>>;
    let terminatePromise: DelayedPromiseResult<void>;
    let sourceId: Identifier;

    beforeEach(() => {
        execution = createMockExecution();
        sourceId = execution.closure.workflowId;
        mockTerminateExecution = jest.fn().mockImplementation(() => {
            terminatePromise = delayedPromise();
            return terminatePromise;
        });
        executionContext = {
            execution,
            terminateExecution: mockTerminateExecution
        };
    });

    const renderContent = () =>
        render(
            <MemoryRouter>
                <ExecutionContext.Provider value={executionContext}>
                    <ExecutionDetailsAppBarContent execution={execution} />
                </ExecutionContext.Provider>
            </MemoryRouter>
        );

    describe('for running executions', () => {
        beforeEach(() => {
            execution.closure.phase = WorkflowExecutionPhase.RUNNING;
        });

        it('renders an overflow menu', async () => {
            const { getByLabelText } = renderContent();
            await waitFor(() => getByLabelText(commonLabels.moreOptionsButton));
        });

        describe('in overflow menu', () => {
            let renderResult: RenderResult;
            let buttonEl: HTMLElement;
            let menuEl: HTMLElement;

            beforeEach(async () => {
                renderResult = renderContent();
                const { getByLabelText } = renderResult;
                buttonEl = await waitFor(() =>
                    getByLabelText(commonLabels.moreOptionsButton)
                );
                fireEvent.click(buttonEl);
                menuEl = await waitFor(() =>
                    getByLabelText(commonLabels.moreOptionsMenu)
                );
            });

            it('renders a clone option', () => {
                const { getByText } = renderResult;
                expect(
                    getByText(executionActionStrings.clone)
                ).toBeInTheDocument();
            });
        });
    });

    describe('for terminal executions', () => {
        beforeEach(() => {
            execution.closure.phase = WorkflowExecutionPhase.SUCCEEDED;
        });

        it('does not render an overflow menu', async () => {
            const { queryByLabelText } = renderContent();
            expect(queryByLabelText(commonLabels.moreOptionsButton)).toBeNull();
        });
    });

    it('renders a back link to the parent workflow', async () => {
        const { getByTitle } = renderContent();
        await waitFor(() =>
            expect(getByTitle(backLinkTitle)).toHaveAttribute(
                'href',
                Routes.WorkflowDetails.makeUrl(
                    sourceId.project,
                    sourceId.domain,
                    sourceId.name
                )
            )
        );
    });

    it('renders the workflow name in the app bar content', async () => {
        const { getByText } = renderContent();
        const { project, domain } = execution.id;
        await waitFor(() =>
            expect(
                getByText(`${project}/${domain}/${sourceId.name}/`)
            ).toBeInTheDocument()
        );
    });

    describe('for single task executions', () => {
        beforeEach(() => {
            execution.spec.launchPlan.resourceType = ResourceType.TASK;
            sourceId = execution.spec.launchPlan;
        });

        it('renders a back link to the parent task', async () => {
            const { getByTitle } = renderContent();
            await waitFor(() =>
                expect(getByTitle(backLinkTitle)).toHaveAttribute(
                    'href',
                    Routes.TaskDetails.makeUrl(
                        sourceId.project,
                        sourceId.domain,
                        sourceId.name
                    )
                )
            );
        });

        it('renders the task name in the app bar content', async () => {
            const { getByText } = renderContent();
            const { project, domain } = execution.id;
            await waitFor(() =>
                expect(
                    getByText(`${project}/${domain}/${sourceId.name}/`)
                ).toBeInTheDocument()
            );
        });
    });
});
