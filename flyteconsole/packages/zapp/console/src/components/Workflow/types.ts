import { WorkflowId } from 'models/Workflow/types';
import { WorkflowExecutionPhase } from 'models/Execution/enums';
import { WorkflowExecutionIdentifier } from 'models/Execution/types';
import { NamedEntityIdentifier } from 'models/Common/types';
import { NamedEntityState } from 'models/enums';

export type WorkflowListItem = {
  id: WorkflowId;
  inputs?: string;
  outputs?: string;
  latestExecutionTime?: string;
  executionStatus?: WorkflowExecutionPhase[];
  executionIds?: WorkflowExecutionIdentifier[];
  description?: string;
  state: NamedEntityState;
};

export type WorkflowListStructureItem = {
  id: NamedEntityIdentifier;
  description: string;
  state: NamedEntityState;
};
