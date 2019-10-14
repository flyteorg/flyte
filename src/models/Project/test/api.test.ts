import { Admin } from 'flyteidl';
import { emptyProject } from 'models/__mocks__/projectData';
import { getAdminEntity } from '../../AdminEntity/AdminEntity';
import { listProjects } from '../api';
jest.mock('../../AdminEntity/AdminEntity.ts');

const mockGetAdminEntity = getAdminEntity as jest.Mock;

describe('Project.api', () => {
    let projects: Admin.Project[];
    beforeEach(() => {
        projects = [
            { ...emptyProject, id: 'projectb', name: 'B Project' },
            { ...emptyProject, id: 'projecta', name: 'aproject' }
        ];
        mockGetAdminEntity.mockImplementation(({ transform }) =>
            Promise.resolve(transform({ projects }))
        );
    });
    describe('listProjects', () => {
        it('sorts projects by case-insensitive name', async () => {
            const projectsResult = await listProjects();
            expect(projectsResult).toEqual([projects[1], projects[0]]);
        });
    });
});
