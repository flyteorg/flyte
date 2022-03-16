import { render, waitFor } from '@testing-library/react';
import { APIContext } from 'components/data/apiContext';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { Admin } from 'flyteidl';
import { FilterOperationName } from 'models/AdminEntity/types';
import { listNamedEntities } from 'models/Common/api';
import { NamedEntity } from 'models/Common/types';
import * as React from 'react';
import { MemoryRouter } from 'react-router';
import { createWorkflowName } from 'test/modelUtils';
import { ProjectWorkflows } from '../ProjectWorkflows';

describe('ProjectWorkflows', () => {
  const project = 'TestProject';
  const domain = 'TestDomain';
  let workflowNames: NamedEntity[];
  let mockListNamedEntities: jest.Mock<ReturnType<typeof listNamedEntities>>;

  beforeEach(() => {
    workflowNames = ['MyWorkflow', 'MyOtherWorkflow'].map((name) =>
      createWorkflowName({ domain, name, project }),
    );
    mockListNamedEntities = jest.fn().mockResolvedValue({ entities: workflowNames });
  });

  const renderComponent = () =>
    render(
      <APIContext.Provider
        value={mockAPIContextValue({
          listNamedEntities: mockListNamedEntities,
        })}
      >
        <ProjectWorkflows projectId={project} domainId={domain} />
      </APIContext.Provider>,
      { wrapper: MemoryRouter },
    );

  it('does not show archived workflows', async () => {
    renderComponent();
    await waitFor(() => {});

    expect(mockListNamedEntities).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({
        filter: [
          {
            key: 'state',
            operation: FilterOperationName.EQ,
            value: Admin.NamedEntityState.NAMED_ENTITY_ACTIVE,
          },
        ],
      }),
    );
  });
});
