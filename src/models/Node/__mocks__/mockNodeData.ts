import { mockTasks } from 'models/Task/__mocks__/mockTaskData';
import { CompiledNode } from '../types';

export const mockNodes: CompiledNode[] = mockTasks.map<CompiledNode>(({ template }) => {
  const { id } = template;
  return { id: id.name, taskNode: { referenceId: id } };
});
