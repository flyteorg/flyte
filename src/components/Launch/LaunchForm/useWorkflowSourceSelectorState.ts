import { APIContextValue, useAPIContext } from 'components/data/apiContext';
import { FilterOperationName, SortDirection } from 'models/AdminEntity/types';
import { NamedEntityIdentifier } from 'models/Common/types';
import { LaunchPlan } from 'models/Launch/types';
import { workflowSortFields } from 'models/Workflow/constants';
import { Workflow, WorkflowId } from 'models/Workflow/types';
import { useMemo, useState } from 'react';
import { SearchableSelectorOption } from './SearchableSelector';
import { WorkflowSourceSelectorState } from './types';
import { useVersionSelectorOptions } from './useVersionSelectorOptions';
import {
  launchPlansToSearchableSelectorOptions,
  versionsToSearchableSelectorOptions,
} from './utils';

function useLaunchPlanSelectorOptions(launchPlans: LaunchPlan[]) {
  return useMemo(() => launchPlansToSearchableSelectorOptions(launchPlans), [launchPlans]);
}

function generateFetchSearchResults(
  { listWorkflows }: APIContextValue,
  workflowId: NamedEntityIdentifier,
) {
  return async (query: string) => {
    const { project, domain, name } = workflowId;
    const { entities: workflows } = await listWorkflows(
      { project, domain, name },
      {
        filter: [
          {
            key: 'version',
            operation: FilterOperationName.CONTAINS,
            value: query,
          },
        ],
        sort: {
          key: workflowSortFields.createdAt,
          direction: SortDirection.DESCENDING,
        },
      },
    );
    return versionsToSearchableSelectorOptions(workflows);
  };
}

interface UseWorkflowSourceSelectorStateArgs {
  /** The currently selected launch plan */
  launchPlan?: LaunchPlan;
  /** List of options to show for the launch plan selector. */
  launchPlanOptions: LaunchPlan[];
  /** The parent workflow for which we are selecting a version. */
  sourceId: NamedEntityIdentifier;
  /** The currently selected Workflow version. */
  workflowVersion?: WorkflowId;
  /** The list of options to show for the Workflow selector. */
  workflowVersionOptions: Workflow[];
  /** Callback fired when a workflow has been selected. */
  selectWorkflowVersion(workflow: WorkflowId): void;
  /** Callback fired when a launch plan has been selected. */
  selectLaunchPlan(launchPlan: LaunchPlan): void;
}

/** Generates state for the workflow/launch plan selectors render when using a workflow
 * as a source in the Launch form.
 */
export function useWorkflowSourceSelectorState({
  launchPlan,
  launchPlanOptions,
  sourceId,
  workflowVersion,
  workflowVersionOptions,
  selectLaunchPlan,
  selectWorkflowVersion,
}: UseWorkflowSourceSelectorStateArgs): WorkflowSourceSelectorState {
  const apiContext = useAPIContext();
  const workflowSelectorOptions = useVersionSelectorOptions(workflowVersionOptions);
  const [workflowVersionSearchOptions, setWorkflowVersionSearchOptions] = useState<
    SearchableSelectorOption<WorkflowId>[]
  >([]);
  const launchPlanSelectorOptions = useLaunchPlanSelectorOptions(launchPlanOptions);

  const selectedWorkflow = useMemo(() => {
    if (!workflowVersion) {
      return undefined;
    }
    // Search both the default and search results to match our selected workflow
    // with the correct SearchableSelector item.
    return [...workflowSelectorOptions, ...workflowVersionSearchOptions].find(
      (option) => option.id === workflowVersion.version,
    );
  }, [workflowVersion, workflowVersionOptions]);

  const selectedLaunchPlan = useMemo(() => {
    if (!launchPlan) {
      return undefined;
    }
    return launchPlanSelectorOptions.find((option) => option.id === launchPlan.id.name);
  }, [launchPlan, launchPlanOptions]);

  const onSelectLaunchPlan = useMemo(
    () =>
      ({ data }: SearchableSelectorOption<LaunchPlan>) =>
        selectLaunchPlan(data),
    [selectLaunchPlan],
  );
  const onSelectWorkflowVersion = useMemo(
    () =>
      ({ data }: SearchableSelectorOption<WorkflowId>) =>
        selectWorkflowVersion(data),
    [selectWorkflowVersion],
  );

  const fetchSearchResults = useMemo(() => {
    const doFetch = generateFetchSearchResults(apiContext, sourceId);
    return async (query: string) => {
      const results = await doFetch(query);
      setWorkflowVersionSearchOptions(results);
      return results;
    };
  }, [apiContext, sourceId, setWorkflowVersionSearchOptions]);

  return {
    fetchSearchResults,
    launchPlanSelectorOptions,
    onSelectLaunchPlan,
    onSelectWorkflowVersion,
    selectedLaunchPlan,
    selectedWorkflow,
    workflowSelectorOptions,
  };
}
