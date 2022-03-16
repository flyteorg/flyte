import { mapValues } from 'lodash';
import { Variable } from 'models/Common/types';
import { createMockLaunchPlan } from 'models/__mocks__/launchPlanData';
import { createMockTask, createMockTaskVersions } from 'models/__mocks__/taskData';
import { createMockWorkflow, createMockWorkflowVersions } from 'models/__mocks__/workflowData';

export function createMockObjects(variables: Record<string, Variable>) {
  const mockWorkflow = createMockWorkflow('MyWorkflow');
  const mockTask = createMockTask('MyTask');

  const mockWorkflowVersions = createMockWorkflowVersions(mockWorkflow.id.name, 10);

  const mockTaskVersions = createMockTaskVersions(mockTask.id.name, 10);

  const mockLaunchPlans = [mockWorkflow.id.name, 'OtherLaunchPlan'].map((name) => {
    const parameterMap = {
      parameters: mapValues(variables, (v) => ({ var: v })),
    };
    const launchPlan = createMockLaunchPlan(name, mockWorkflow.id.version);
    launchPlan.closure!.expectedInputs = parameterMap;
    return launchPlan;
  });
  return {
    mockWorkflow,
    mockLaunchPlans,
    mockTask,
    mockTaskVersions,
    mockWorkflowVersions,
  };
}
