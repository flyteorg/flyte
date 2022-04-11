import { WorkflowId } from 'models/Workflow/types';
import { WorkflowExecutionPhase } from 'models/Execution/enums';
import { WorkflowExecutionIdentifier } from 'models/Execution/types';
import { NamedEntityIdentifier } from 'models/Common/types';
import { WorkflowExecutionState } from 'models/Workflow/enums';

export type WorkflowListItem = {
  id: WorkflowId;
  inputs?: string;
  outputs?: string;
  latestExecutionTime?: string;
  executionStatus?: WorkflowExecutionPhase[];
  executionIds?: WorkflowExecutionIdentifier[];
  description?: string;
  state: WorkflowExecutionState;
};

export type WorkflowListStructureItem = {
  id: NamedEntityIdentifier;
  description: string;
  state: WorkflowExecutionState;
};
