import { emptyProject } from 'mocks/data/projects';
import { mockServer } from 'mocks/server';
import { listProjects } from '../api';
import { Project } from '../types';

describe('Project.api', () => {
  let projects: Project[];
  beforeEach(() => {
    projects = [emptyProject('projectb', 'B Project'), emptyProject('projecta', 'aproject')];
    mockServer.insertProjects(projects);
  });
  describe('listProjects', () => {
    it('sorts projects by case-insensitive name', async () => {
      const projectsResult = await listProjects();
      expect(projectsResult).toEqual([projects[1], projects[0]]);
    });
  });
});
