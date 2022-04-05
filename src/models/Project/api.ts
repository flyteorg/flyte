import { Admin } from 'flyteidl';
import { sortBy } from 'lodash';
import { endpointPrefixes } from 'models/Common/constants';

import { getAdminEntity } from 'models/AdminEntity/AdminEntity';
import { IdentifierScope } from 'models/Common/types';
import { Project } from './types';
import { makeProjectDomainAttributesPath } from './utils';

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

export const getProjectDomainAttributes = (scope: IdentifierScope) =>
  getAdminEntity<
    Admin.ProjectDomainAttributesGetResponse,
    Admin.ProjectDomainAttributesGetResponse
  >(
    {
      path: makeProjectDomainAttributesPath(endpointPrefixes.projectDomainAtributes, scope),
      messageType: Admin.ProjectDomainAttributesGetResponse,
    },
    {
      params: {
        resource_type: 'WORKFLOW_EXECUTION_CONFIG',
      },
    },
  );
