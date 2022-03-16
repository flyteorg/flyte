import { Admin } from 'flyteidl';
import { sortBy } from 'lodash';
import { endpointPrefixes } from 'models/Common/constants';

import { getAdminEntity } from 'models/AdminEntity/AdminEntity';
import { Project } from './types';

/** Fetches the list of available `Project`s */
export const listProjects = () =>
  getAdminEntity<Admin.Projects, Project[]>({
    path: endpointPrefixes.project,
    messageType: Admin.Projects,
    // We want the returned list to be sorted ascending by name, but the
    // admin endpoint doesn't support sorting for projects.
    transform: ({ projects }: Admin.Projects) =>
      sortBy(projects, (project) => `${project.name}`.toLowerCase()) as Project[],
  });
