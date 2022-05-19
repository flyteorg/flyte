import { render, waitFor, screen } from '@testing-library/react';
import * as React from 'react';
import { createMockTask } from 'models/__mocks__/taskData';
import { Task } from 'models/Task/types';
import { Identifier } from 'models/Common/types';
import { versionDetailsUrlGenerator } from 'components/Entities/generators';
import { TaskVersionDetailsLink } from '../VersionDetails/VersionDetailsLink';

describe('TaskVersionDetailsLink', () => {
  let mockTask: Task;

  const createMocks = () => {
    mockTask = createMockTask('MyTask');
  };

  const renderLink = (id: Identifier) => {
    return render(<TaskVersionDetailsLink id={id} />);
  };

  beforeEach(() => {
    createMocks();
  });

  it('renders and checks text', async () => {
    const id: Identifier = mockTask.id;
    renderLink(id);
    await waitFor(() => {
      expect(screen.getByText('Task Details')).toBeInTheDocument();
    });
  });

  it('renders and checks containing icon', () => {
    const id: Identifier = mockTask.id;
    const { container } = renderLink(id);
    expect(container.querySelector('svg')).not.toBeNull();
  });

  it('renders and checks url', () => {
    const id: Identifier = mockTask.id;
    const { container } = renderLink(id);
    expect(container.firstElementChild).toHaveAttribute('href', versionDetailsUrlGenerator(id));
  });
});
