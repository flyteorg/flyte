import { getCacheKey } from 'components/Cache/utils';
import { Admin } from 'flyteidl';
import { cloneDeep } from 'lodash';
import { Identifier } from 'models/Common/types';
import { Workflow, WorkflowClosure } from 'models/Workflow/types';
import * as simpleClosure from './simpleWorkflowClosure.json';

const decodedClosure = Admin.WorkflowClosure.create(
  simpleClosure as unknown as Admin.IWorkflowClosure,
) as WorkflowClosure;

const workflowId: (name: string, version: string) => Identifier = (name, version) => ({
  name,
  version,
  project: 'flyte',
  domain: 'development',
});

export const createMockWorkflow: (name: string, version?: string) => Workflow = (
  name: string,
  version = 'abcdefg',
) => ({
  id: workflowId(name, version),
});

export const createMockWorkflowClosure: () => WorkflowClosure = () => cloneDeep(decodedClosure);

export const createMockWorkflowVersions = (name: string, length: number) => {
  return Array.from({ length }, (_, idx) => {
    return createMockWorkflow(name, getCacheKey({ idx }));
  });
};
