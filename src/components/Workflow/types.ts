import { WorkflowId } from 'models/Workflow/types';
import { WorkflowExecutionPhase } from 'models/Execution/enums';
import { WorkflowExecutionIdentifier } from 'models/Execution/types';
import { NamedEntityIdentifier } from 'models/Common/types';

export type WorkflowListItem = {
  id: WorkflowId;
  inputs?: string;
  outputs?: string;
  latestExecutionTime?: string;
  executionStatus?: WorkflowExecutionPhase[];
  executionIds?: WorkflowExecutionIdentifier[];
  description?: string;
};

export type WorkflowListStructureItem = {
  id: NamedEntityIdentifier;
  description: string;
};
