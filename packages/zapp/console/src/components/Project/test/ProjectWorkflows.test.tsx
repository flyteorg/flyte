import { fireEvent, render, waitFor } from '@testing-library/react';
import { APIContext } from 'components/data/apiContext';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { FetchableData } from 'components/hooks/types';
import { useUserProfile } from 'components/hooks/useUserProfile';
import { loadedFetchable } from 'components/hooks/__mocks__/fetchableData';
import { FilterOperationName } from 'models/AdminEntity/types';
import { listNamedEntities } from 'models/Common/api';
import { NamedEntity, UserProfile } from 'models/Common/types';
import { NamedEntityState } from 'models/enums';
import * as React from 'react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { MemoryRouter } from 'react-router';
import { createWorkflowName } from 'test/modelUtils';
import { createTestQueryClient } from 'test/utils';
import { ProjectWorkflows } from '../ProjectWorkflows';

const sampleUserProfile: UserProfile = {
  subject: 'subject',
} as UserProfile;

jest.mock('components/hooks/useUserProfile');
jest.mock('notistack', () => ({
  useSnackbar: () => ({ enqueueSnackbar: jest.fn() }),
}));

describe('ProjectWorkflows', () => {
  const project = 'TestProject';
  const domain = 'TestDomain';
  let workflowNames: NamedEntity[];
  let queryClient: QueryClient;
  let mockListNamedEntities: jest.Mock<ReturnType<typeof listNamedEntities>>;

  const mockUseUserProfile = useUserProfile as jest.Mock<FetchableData<UserProfile | null>>;

  beforeEach(() => {
    mockUseUserProfile.mockReturnValue(loadedFetchable(null, jest.fn()));
    queryClient = createTestQueryClient();
    workflowNames = ['MyWorkflow', 'MyOtherWorkflow'].map((name) =>
      createWorkflowName({ domain, name, project }),
    );
    mockListNamedEntities = jest.fn().mockResolvedValue({ entities: workflowNames });
  });

  const renderComponent = () =>
    render(
      <QueryClientProvider client={queryClient}>
        <APIContext.Provider
          value={mockAPIContextValue({ listNamedEntities: mockListNamedEntities })}
        >
          <ProjectWorkflows projectId={project} domainId={domain} />
        </APIContext.Provider>
      </QueryClientProvider>,
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
            value: NamedEntityState.NAMED_ENTITY_ACTIVE,
          },
        ],
      }),
    );
  });

  it('should display checkbox if user login', async () => {
    mockUseUserProfile.mockReturnValue(loadedFetchable(sampleUserProfile, jest.fn()));
    const { getAllByRole } = renderComponent();
    await waitFor(() => {});
    const checkboxes = getAllByRole(/checkbox/i) as HTMLInputElement[];
    expect(checkboxes).toHaveLength(1);
    expect(checkboxes[0]).toBeTruthy();
    expect(checkboxes[0]?.checked).toEqual(false);
  });

  /** user doesn't have its own workflow */
  it('clicking show archived should hide active workflows', async () => {
    mockUseUserProfile.mockReturnValue(loadedFetchable(sampleUserProfile, jest.fn()));
    const { getByText, queryByText, getAllByRole } = renderComponent();
    await waitFor(() => {});
    const checkboxes = getAllByRole(/checkbox/i) as HTMLInputElement[];
    expect(checkboxes[0]).toBeTruthy();
    expect(checkboxes[0]?.checked).toEqual(false);
    await waitFor(() => expect(getByText('MyWorkflow')));
    fireEvent.click(checkboxes[0]);
    // when user selects checkbox, table should have no workflows to display
    await waitFor(() => expect(queryByText('MyWorkflow')).not.toBeInTheDocument());
  });
});
