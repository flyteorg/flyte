import { Identifier } from 'models/Common/types';
import { Project } from './types';

export function getProjectDomain(project: Project, domainId: string) {
  const domain = project.domains.find((d) => d.id === domainId);
  if (!domain) {
    throw new Error(`Project ${project.name} has no domain with id ${domainId}`);
  }
  return domain;
}

export function makeProjectDomainAttributesPath(
  prefix: string,
  { project, domain }: Partial<Identifier>,
) {
  return [prefix, project, domain].join('/');
}
