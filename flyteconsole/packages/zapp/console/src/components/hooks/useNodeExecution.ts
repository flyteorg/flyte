import { useAPIContext } from 'components/data/apiContext';
import { ExecutionData, NodeExecution, NodeExecutionIdentifier } from 'models/Execution/types';
import { FetchableData } from './types';
import { useFetchableData } from './useFetchableData';

/** A hook for fetching a NodeExecution */
export function useNodeExecution(id: NodeExecutionIdentifier): FetchableData<NodeExecution> {
  const { getNodeExecution } = useAPIContext();
  return useFetchableData<NodeExecution, NodeExecutionIdentifier>(
    {
      debugName: 'NodeExecution',
      defaultValue: {} as NodeExecution,
      doFetch: (id) => getNodeExecution(id),
    },
    id,
  );
}

/** Fetches the signed URLs for NodeExecution data (inputs/outputs) */
export function useNodeExecutionData(id: NodeExecutionIdentifier): FetchableData<ExecutionData> {
  const { getNodeExecutionData } = useAPIContext();
  return useFetchableData<ExecutionData, NodeExecutionIdentifier>(
    {
      debugName: 'NodeExecutionData',
      defaultValue: {} as ExecutionData,
      doFetch: (id) => getNodeExecutionData(id),
    },
    id,
  );
}
