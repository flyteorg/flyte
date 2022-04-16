import { NamedEntityState } from 'models/enums';
import { WorkflowListStructureItem } from './types';

function isWorkflowStateArchive(workflow: WorkflowListStructureItem): boolean {
  const state = workflow?.state ?? null;
  return !!state && state === NamedEntityState.NAMED_ENTITY_ARCHIVED;
}

export function isWorkflowArchived(workflow: WorkflowListStructureItem): boolean {
  return isWorkflowStateArchive(workflow);
}
