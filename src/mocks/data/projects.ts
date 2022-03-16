import { Project } from 'models/Project/types';
import { testDomain, testProject } from './constants';

export function emptyProject(id: string, name?: string) {
  return {
    id,
    name: name ?? id,
    domains: [],
    description: '',
  };
}

const flyteTest: Project = {
  id: testProject,
  name: testProject,
  description: 'An umbrella project with a single domain to contain all of the test data.',
  domains: [
    {
      id: testDomain,
      name: testDomain,
    },
  ],
};

export const projects = {
  flyteTest,
};
