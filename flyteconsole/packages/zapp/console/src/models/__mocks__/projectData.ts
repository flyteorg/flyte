import { Domain, Project } from 'models/Project/types';

export const mockDomainIds = ['development', 'production'];
export const mockProjectIds = Array.from(Array(10).keys()).map((idx) => `project number ${idx}`);
const makeDomain: (id: string) => Domain = (id) => ({
  id,
  name: id,
});

const makeDomainList: (domainIds: string[]) => Domain[] = (domainIds) =>
  domainIds.map((id) => makeDomain(id));

export const createMockProjects: () => Project[] = () =>
  mockProjectIds.map((id) => ({
    id,
    name: id,
    domains: makeDomainList(mockDomainIds),
  }));
