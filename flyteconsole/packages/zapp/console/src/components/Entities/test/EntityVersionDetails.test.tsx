import { render, waitFor, screen } from '@testing-library/react';
import { ThemeProvider } from '@material-ui/styles';
import { muiTheme } from 'components/Theme/muiTheme';
import { ResourceIdentifier } from 'models/Common/types';
import * as React from 'react';
import { createMockTask } from 'models/__mocks__/taskData';
import { Task } from 'models/Task/types';
import { getTask } from 'models/Task/api';
import { APIContext } from 'components/data/apiContext';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { EntityVersionDetails } from '../VersionDetails/EntityVersionDetails';

describe('EntityVersionDetails', () => {
  let mockTask: Task;
  let mockGetTask: jest.Mock<ReturnType<typeof getTask>>;

  const createMocks = () => {
    mockTask = createMockTask('MyTask');
    mockGetTask = jest.fn().mockImplementation(() => Promise.resolve(mockTask));
  };

  const renderDetails = (id: ResourceIdentifier) => {
    return render(
      <ThemeProvider theme={muiTheme}>
        <APIContext.Provider
          value={mockAPIContextValue({
            getTask: mockGetTask,
          })}
        >
          <EntityVersionDetails id={id} />
        </APIContext.Provider>
      </ThemeProvider>,
    );
  };

  describe('Task Version Details', () => {
    beforeEach(() => {
      createMocks();
    });

    it('renders and checks text', async () => {
      const id: ResourceIdentifier = mockTask.id as ResourceIdentifier;
      renderDetails(id);

      // check text for Task Details
      await waitFor(() => {
        expect(screen.getByText('Task Details')).toBeInTheDocument();
      });

      // check text for image
      await waitFor(() => {
        expect(
          screen.getByText(mockTask.closure.compiledTask.template?.container?.image || ''),
        ).toBeInTheDocument();
      });

      // check for env vars
      if (mockTask.closure.compiledTask.template?.container?.env) {
        const envVars = mockTask.closure.compiledTask.template?.container?.env;
        for (let i = 0; i < envVars.length; i++) {
          await waitFor(() => {
            expect(screen.getByText(envVars[i].key || '')).toBeInTheDocument();
            expect(screen.getByText(envVars[i].value || '')).toBeInTheDocument();
          });
        }
      }
    });
  });
});
