// Canvas will be created on first use and cached for better performance
let textMeasureCanvas: HTMLCanvasElement;

/** Uses a canvas to measure the needed width to render a string using a given
 * font definition.
 */
export function measureText(fontDefinition: string, text: string) {
  if (!textMeasureCanvas) {
    textMeasureCanvas = document.createElement('canvas');
  }
  const context = textMeasureCanvas.getContext('2d');
  if (!context) {
    throw new Error('Unable to create canvas context for text measurement');
  }

  context.font = fontDefinition;
  return context.measureText(text);
}

/**
 * Note:
 *  Dynamic nodes are deteremined at runtime and thus do not come
 *  down as part of the workflow closure. We can detect and place
 *  dynamic nodes by finding orphan execution id's and then mapping
 *  those executions into the dag by using the executions 'uniqueParentId'
 *  to render that node as a subworkflow
 */
export const checkForDynamicExecutions = (allExecutions, staticExecutions) => {
  const parentsToFetch = {};
  for (const executionId in allExecutions) {
    if (!staticExecutions[executionId]) {
      const dynamicExecution = allExecutions[executionId];
      const dynamicExecutionId = dynamicExecution.metadata.specNodeId || dynamicExecution.id;
      const uniqueParentId = dynamicExecution.fromUniqueParentId;
      if (uniqueParentId) {
        if (parentsToFetch[uniqueParentId]) {
          parentsToFetch[uniqueParentId].push(dynamicExecutionId);
        } else {
          parentsToFetch[uniqueParentId] = [dynamicExecutionId];
        }
      }
    }
  }
  const result = {};
  for (const parentId in parentsToFetch) {
    result[parentId] = allExecutions[parentId];
  }
  return result;
};
