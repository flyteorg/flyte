import { fireEvent, render } from '@testing-library/react';
import { KeyCodes } from 'common/constants';
import { createMockProjects } from 'models/__mocks__/projectData';
import * as React from 'react';
import { ProjectSelector, ProjectSelectorProps } from '../ProjectSelector';

describe('ProjectSelector', () => {
  let props: ProjectSelectorProps;
  const renderProjectSelector = () => render(<ProjectSelector {...props} />);

  beforeEach(() => {
    const projects = createMockProjects();
    props = {
      projects,
      onProjectSelected: jest.fn(),
      selectedProject: projects[0],
    };
  });
  describe('while collapsed', () => {
    it('should not render list items', () => {
      const { queryByText } = renderProjectSelector();
      expect(queryByText(props.projects[1].name)).toBeNull();
    });
  });

  describe('while expanded', () => {
    let rendered: ReturnType<typeof renderProjectSelector>;
    beforeEach(() => {
      rendered = renderProjectSelector();
      const expander = rendered.getByText(props.projects[0].name);
      fireEvent.click(expander);
    });

    it('should render list items', () => {
      expect(rendered.getByText(props.projects[1].name)).toBeTruthy();
    });

    it('should collapse when hitting escape', () => {
      fireEvent.keyDown(rendered.getByRole('search'), {
        keyCode: KeyCodes.ESCAPE,
      });
      expect(rendered.queryByText(props.projects[1].name)).toBeNull();
    });

    it('should collapse when an item is selected', () => {
      fireEvent.click(rendered.getByText(props.projects[1].name));
      expect(rendered.queryByText(props.projects[1].name)).toBeNull();
    });
  });
});
