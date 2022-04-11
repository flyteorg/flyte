import { DomainIdentifierScope, ResourceType } from 'models/Common/types';
import { RequestConfig } from 'models/AdminEntity/types';
import { usePagination } from 'components/hooks/usePagination';
import { useAPIContext } from 'components/data/apiContext';
import { WorkflowListStructureItem } from './types';

export const useWorkflowInfoList = (scope: DomainIdentifierScope, config?: RequestConfig) => {
  const { listNamedEntities } = useAPIContext();

  return usePagination<WorkflowListStructureItem, DomainIdentifierScope>(
    { ...config, fetchArg: scope },
    async (scope, requestConfig) => {
      const { entities, ...rest } = await listNamedEntities(
        { ...scope, resourceType: ResourceType.WORKFLOW },
        requestConfig,
      );

      return {
        entities: entities.map(({ id, metadata: { description, state } }) => ({
          id,
          description,
          state,
        })),
        ...rest,
      };
    },
  );
};
