import { getCacheKey } from 'components/Cache/utils';
import { Admin } from 'flyteidl';
import { cloneDeep } from 'lodash';
import { Identifier } from 'models/Common/types';
import { Task, TaskClosure } from 'models/Task/types';
import * as simpleClosure from './simpleTaskClosure.json';

const decodedClosure = Admin.TaskClosure.create(
  simpleClosure as unknown as Admin.ITaskClosure,
) as TaskClosure;

const taskId: (name: string, version: string) => Identifier = (name, version) => ({
  name,
  version,
  project: 'flyte',
  domain: 'development',
});

export const createMockTask: (name: string, version?: string) => Task = (
  name: string,
  version = 'abcdefg',
) => ({
  id: taskId(name, version),
  closure: createMockTaskClosure(),
});

export const createMockTaskClosure: () => TaskClosure = () => cloneDeep(decodedClosure);

export const createMockTaskVersions = (name: string, length: number) => {
  return Array.from({ length }, (_, idx) => {
    return createMockTask(name, getCacheKey({ idx }));
  });
};
