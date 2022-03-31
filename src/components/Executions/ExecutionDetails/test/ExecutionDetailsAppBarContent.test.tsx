import { fireEvent, render, RenderResult, waitFor } from '@testing-library/react';
import { labels as commonLabels } from 'components/common/constants';
import { ExecutionContext, ExecutionContextData } from 'components/Executions/contexts';
import { Identifier, ResourceType } from 'models/Common/types';
import { WorkflowExecutionPhase } from 'models/Execution/enums';
import { Execution } from 'models/Execution/types';
import { createMockExecution } from 'models/__mocks__/executionsData';
import * as React from 'react';
import { MemoryRouter } from 'react-router';
import { Routes } from 'routes/routes';
import { QueryClient, QueryClientProvider } from 'react-query';
import { createTestQueryClient } from 'test/utils';
import { backLinkTitle, executionActionStrings } from '../constants';
import { ExecutionDetailsAppBarContent } from '../ExecutionDetailsAppBarContent';

jest.mock('components/Navigation/NavBarContent', () => ({
  NavBarContent: ({ children }: React.Props<any>) => children,
}));

describe('ExecutionDetailsAppBarContent', () => {
  let execution: Execution;
  let executionContext: ExecutionContextData;
  let sourceId: Identifier;
  let queryClient: QueryClient;

  beforeEach(() => {
    execution = createMockExecution();
    sourceId = execution.closure.workflowId;

    executionContext = {
      execution,
    };

    queryClient = createTestQueryClient();
  });

  const renderContent = () =>
    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>
          <ExecutionContext.Provider value={executionContext}>
            <ExecutionDetailsAppBarContent execution={execution} />
          </ExecutionContext.Provider>
        </MemoryRouter>
      </QueryClientProvider>,
    );

  describe('for running executions', () => {
    beforeEach(() => {
      execution.closure.phase = WorkflowExecutionPhase.RUNNING;
    });

    it('renders an overflow menu', async () => {
      const { getByLabelText } = renderContent();
      expect(getByLabelText(commonLabels.moreOptionsButton)).toBeTruthy();
    });

    describe('in overflow menu', () => {
      let renderResult: RenderResult;
      let buttonEl: HTMLElement;

      beforeEach(async () => {
        renderResult = renderContent();
        const { getByLabelText } = renderResult;
        buttonEl = await waitFor(() => getByLabelText(commonLabels.moreOptionsButton));
        fireEvent.click(buttonEl);
        await waitFor(() => getByLabelText(commonLabels.moreOptionsMenu));
      });

      it('renders a clone option', () => {
        const { getByText } = renderResult;
        expect(getByText(executionActionStrings.clone)).toBeInTheDocument();
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
        Routes.WorkflowDetails.makeUrl(sourceId.project, sourceId.domain, sourceId.name),
      ),
    );
  });

  it('renders the workflow name in the app bar content', async () => {
    const { getByText } = renderContent();
    const { project, domain } = execution.id;
    await waitFor(() =>
      expect(getByText(`${project}/${domain}/${sourceId.name}/`)).toBeInTheDocument(),
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
          Routes.TaskDetails.makeUrl(sourceId.project, sourceId.domain, sourceId.name),
        ),
      );
    });

    it('renders the task name in the app bar content', async () => {
      const { getByText } = renderContent();
      const { project, domain } = execution.id;
      await waitFor(() =>
        expect(getByText(`${project}/${domain}/${sourceId.name}/`)).toBeInTheDocument(),
      );
    });
  });
});
