import { Domain, Project } from '../Project';

export const emptyProject = {
    id: 'emptyproject',
    name: 'emptyproject',
    domains: [],
    description: ''
};
export const mockDomainIds = ['development', 'production'];
export const mockProjectIds = Array.from(Array(10).keys()).map(
    idx => `project number ${idx}`
);
const makeDomain: (id: string) => Domain = id => ({
    id,
    name: id
});

const makeDomainList: (domainIds: string[]) => Domain[] = domainIds =>
    domainIds.map(id => makeDomain(id));

export const createMockProjects: () => Project[] = () =>
    mockProjectIds.map(id => ({
        id,
        name: id,
        domains: makeDomainList(mockDomainIds)
    }));

export const createMockProjectsMap: () => Map<string, Project> = () => {
    const projects = createMockProjects();
    return new Map(
        projects.map<[string, Project]>(p => [p.id, p])
    );
};
