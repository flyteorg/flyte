import { Core } from 'flyteidl';
import { getNamedEntity, NamedEntity } from 'models';
import { useFetchableData } from './useFetchableData';

export interface UseNamedEntityInput {
    resourceType: Core.ResourceType;
    project: string;
    domain: string;
    name: string;
}

/** Fetches a NamedEntity (Workflow, LaunchPlan, Task, etc) for a given
 * resourceType/project/domain/name. This is useful to determine any metadata
 * associated with the given entity name */
export function useNamedEntity(input: UseNamedEntityInput) {
    return useFetchableData<NamedEntity, UseNamedEntityInput>(
        {
            debugName: 'NamedEntity',
            defaultValue: {} as NamedEntity,
            doFetch: id => getNamedEntity(id),
            useCache: true
        },
        input
    );
}

/** Fetches the NamedEntity record for a LaunchPlan */
export function useLaunchPlanNamedEntity(
    input: Omit<UseNamedEntityInput, 'resourceType'>
) {
    return useNamedEntity({
        ...input,
        resourceType: Core.ResourceType.LAUNCH_PLAN
    });
}

/** Fetches the NamedEntity record for a Task */
export function useTaskNamedEntity(
    input: Omit<UseNamedEntityInput, 'resourceType'>
) {
    return useNamedEntity({ ...input, resourceType: Core.ResourceType.TASK });
}

/** Fetches the NamedEntity record for a Workflow */
export function useWorkflowNamedEntity(
    input: Omit<UseNamedEntityInput, 'resourceType'>
) {
    return useNamedEntity({
        ...input,
        resourceType: Core.ResourceType.WORKFLOW
    });
}
