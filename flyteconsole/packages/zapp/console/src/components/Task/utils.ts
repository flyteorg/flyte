import { NamedEntity } from 'models/Common/types';
import { NamedEntityState } from 'models/enums';

function isTaskStateArchive(task: NamedEntity): boolean {
  const state = task?.metadata?.state ?? null;
  return !!state && state === NamedEntityState.NAMED_ENTITY_ARCHIVED;
}

export function isTaskArchived(task: NamedEntity): boolean {
  return isTaskStateArchive(task);
}
