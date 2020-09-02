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
import { Execution } from 'models';
import { createMockExecution } from 'models/__mocks__/executionsData';
import { WorkflowExecutionPhase } from 'models/Execution/enums';
import * as React from 'react';
import { MemoryRouter } from 'react-router';
import { delayedPromise, DelayedPromiseResult } from 'test/utils';
import { executionActionStrings } from '../constants';
import { ExecutionDetailsAppBarContent } from '../ExecutionDetailsAppBarContent';

jest.mock('components/Navigation/NavBarContent', () => ({
    NavBarContent: ({ children }: React.Props<any>) => children
}));

describe('ExecutionDetailsAppBarContent', () => {
    let execution: Execution;
    let executionContext: ExecutionContextData;
    let mockTerminateExecution: jest.Mock<Promise<void>>;
    let terminatePromise: DelayedPromiseResult<void>;

    beforeEach(() => {
        execution = createMockExecution();
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
});
