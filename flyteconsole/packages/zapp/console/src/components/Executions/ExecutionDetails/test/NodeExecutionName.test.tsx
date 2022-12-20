import * as React from 'react';
import { render, waitFor } from '@testing-library/react';
import { NodeExecutionDetailsContextProvider } from 'components/Executions/contextProvider/NodeExecutionDetails';
import { mockWorkflowId } from 'mocks/data/fixtures/types';
import { QueryClient, QueryClientProvider } from 'react-query';
import { createTestQueryClient } from 'test/utils';
import { insertFixture } from 'mocks/data/insertFixture';
import { mockServer } from 'mocks/server';
import { basicPythonWorkflow } from 'mocks/data/fixtures/basicPythonWorkflow';
import { NodeExecution } from 'models/Execution/types';
import { NodeExecutionName } from '../Timeline/NodeExecutionName';

jest.mock('components/Workflow/workflowQueries');
const { fetchWorkflow } = require('components/Workflow/workflowQueries');

const name = 'Test';
const templateName = 'TemplateTest';

describe('Executions > ExecutionDetails > NodeExecutionName', () => {
  let queryClient: QueryClient;
  let fixture: ReturnType<typeof basicPythonWorkflow.generate>;
  let execution: NodeExecution;

  beforeEach(() => {
    fixture = basicPythonWorkflow.generate();
    execution = fixture.workflowExecutions.top.nodeExecutions.pythonNode.data;
    queryClient = createTestQueryClient();
    insertFixture(mockServer, fixture);
    fetchWorkflow.mockImplementation(() => Promise.resolve(fixture.workflows.top));
  });

  const renderComponent = (props) =>
    render(
      <QueryClientProvider client={queryClient}>
        <NodeExecutionDetailsContextProvider workflowId={mockWorkflowId}>
          <NodeExecutionName {...props} />
        </NodeExecutionDetailsContextProvider>
      </QueryClientProvider>,
    );

  it('should only display title if execution is not provided', async () => {
    const { queryByText } = renderComponent({ name, templateName });
    await waitFor(() => queryByText(name));

    expect(queryByText(name)).toBeInTheDocument();
    expect(queryByText(templateName)).not.toBeInTheDocument();
  });

  it('should only display title if template name is not provided', async () => {
    const resultName = 'PythonTask';
    const { queryByText } = renderComponent({ name, execution });
    await waitFor(() => queryByText(resultName));

    expect(queryByText(resultName)).toBeInTheDocument();
    expect(queryByText(templateName)).not.toBeInTheDocument();
  });

  it('should display title and subtitle if template name is provided', async () => {
    const resultName = 'PythonTask';
    const { queryByText } = renderComponent({ name, templateName, execution });
    await waitFor(() => queryByText(resultName));

    expect(queryByText(resultName)).toBeInTheDocument();
    expect(queryByText(templateName)).toBeInTheDocument();
  });
});
