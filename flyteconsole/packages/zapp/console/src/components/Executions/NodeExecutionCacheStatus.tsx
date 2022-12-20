import { NodeExecutionDetails } from 'components/Executions/types';
import { useNodeExecutionContext } from 'components/Executions/contextProvider/NodeExecutionDetails';
import { CatalogCacheStatus } from 'models/Execution/enums';
import { NodeExecution } from 'models/Execution/types';
import * as React from 'react';
import { isMapTaskType } from 'models/Task/utils';
import { useEffect, useState } from 'react';
import { CacheStatus } from './CacheStatus';

interface NodeExecutionCacheStatusProps {
  execution: NodeExecution;
  /** `normal` will render an icon with description message beside it
   *  `iconOnly` will render just the icon with the description as a tooltip
   */
  variant?: 'normal' | 'iconOnly';
}
/** For a given `NodeExecution.closure.taskNodeMetadata` object, will render
 * the cache status with a descriptive message. For `Core.CacheCatalogStatus.CACHE_HIT`,
 * it will also attempt to render a link to the source `WorkflowExecution` (normal
 * variant only).
 *
 * For Map Tasks, we will check the NodeExecutionDetail for the cache status instead. Since map tasks
 * cotains multiple tasks, the logic of the cache status is different.
 */
export const NodeExecutionCacheStatus: React.FC<NodeExecutionCacheStatusProps> = ({
  execution,
  variant = 'normal',
}) => {
  const taskNodeMetadata = execution.closure?.taskNodeMetadata;
  const { getNodeExecutionDetails } = useNodeExecutionContext();
  const [nodeDetails, setNodeDetails] = useState<NodeExecutionDetails | undefined>();

  useEffect(() => {
    let isCurrent = true;
    getNodeExecutionDetails(execution).then((res) => {
      if (isCurrent) {
        setNodeDetails(res);
      }
    });
    return () => {
      isCurrent = false;
    };
  });

  if (isMapTaskType(nodeDetails?.taskTemplate?.type)) {
    if (nodeDetails?.taskTemplate?.metadata?.cacheSerializable) {
      return <CacheStatus cacheStatus={CatalogCacheStatus.MAP_CACHE} variant={variant} />;
    }
  }

  // cachestatus can be 0
  if (taskNodeMetadata?.cacheStatus == null) {
    return null;
  }

  return (
    <CacheStatus
      cacheStatus={taskNodeMetadata.cacheStatus}
      sourceTaskExecutionId={taskNodeMetadata.catalogKey?.sourceTaskExecution}
      variant={variant}
    />
  );
};
