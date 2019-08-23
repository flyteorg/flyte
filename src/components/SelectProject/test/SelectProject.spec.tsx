import * as React from 'react';
import * as renderer from 'react-test-renderer';

import { useProjects } from 'components/hooks';
import {
    loadedFetchable,
    loadingFetchable
} from 'components/hooks/__mocks__/fetchableData';
import { Project } from 'models';
import { createMockProjects } from 'models/__mocks__/projectData';

import { SelectProject } from '../SelectProject';

// Skipping full render of the ProjectList component because it has its own snapshot
jest.mock('../ProjectList');
jest.mock('components/hooks');

// TODO: Enzyme doesn't yet have support for hooks, and the timeline isn't looking like any time soon. The current best option is to switch to using react-testing-library instead.
describe.skip('SelectProject', () => {
    const useProjectsMock = useProjects as jest.Mock<
        ReturnType<typeof useProjects>
    >;
    const renderSnapshot = () => renderer.create(<SelectProject />).toJSON();

    describe('while loading', () => {
        beforeEach(() => {
            useProjectsMock.mockReturnValue(loadingFetchable([]));
        });
        it('should match the known snapshot', () => {
            expect(renderSnapshot()).toMatchSnapshot();
        });
    });

    describe('After successfully loading', () => {
        let projects: Project[];
        beforeEach(() => {
            projects = createMockProjects();
            useProjectsMock.mockReturnValue(loadedFetchable(projects));
        });

        it('should match the known snapshot', () => {
            expect(renderSnapshot()).toMatchSnapshot();
        });
    });
});
