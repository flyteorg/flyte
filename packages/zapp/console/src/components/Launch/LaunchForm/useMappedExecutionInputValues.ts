import { log } from 'common/log';
import { APIContextValue, useAPIContext } from 'components/data/apiContext';
import { fetchWorkflowExecutionInputs } from 'components/Executions/useWorkflowExecution';
import { FetchableData } from 'components/hooks/types';
import { useFetchableData } from 'components/hooks/useFetchableData';
import { Variable } from 'models/Common/types';
import { Execution } from 'models/Execution/types';
import { LiteralValueMap } from './types';
import { createInputCacheKey, getInputDefintionForLiteralType } from './utils';

export interface UseMappedExecutionInputValuesArgs {
  execution: Execution;
  inputDefinitions: Record<string, Variable>;
}

export async function fetchAndMapExecutionInputValues(
  { execution, inputDefinitions }: UseMappedExecutionInputValuesArgs,
  apiContext: APIContextValue,
): Promise<LiteralValueMap> {
  const inputValues = await fetchWorkflowExecutionInputs(execution, apiContext);

  const { literals } = inputValues;
  const values: LiteralValueMap = Object.keys(literals).reduce((out, name) => {
    const input = inputDefinitions[name];
    if (!input) {
      log.error(`Unexpected missing input definition: ${name}`);
      return out;
    }
    const typeDefinition = getInputDefintionForLiteralType(input.type);

    const key = createInputCacheKey(name, typeDefinition);
    out.set(key, literals[name]);
    return out;
  }, new Map());
  return values;
}

/** Returns a fetchable that will result in a LiteralValueMap representing the
 * input values from an execution mapped to a set of input definitions from a Task or Workflow */
export function useMappedExecutionInputValues({
  execution,
  inputDefinitions,
}: UseMappedExecutionInputValuesArgs): FetchableData<LiteralValueMap> {
  const apiContext = useAPIContext();
  return useFetchableData<LiteralValueMap, UseMappedExecutionInputValuesArgs>(
    {
      debugName: 'MappedExecutionInputValues',
      defaultValue: {} as LiteralValueMap,
      doFetch: async ({ execution, inputDefinitions }) =>
        fetchAndMapExecutionInputValues({ execution, inputDefinitions }, apiContext),
    },
    { execution, inputDefinitions },
  );
}
