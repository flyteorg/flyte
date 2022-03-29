import { CatalogCacheStatus, NodeExecutionPhase } from 'models/Execution/enums';

const dNodeBasicExecution = {
  id: 'other-root-n0',
  scopedId: 'other-root-n0',
  execution: {
    id: {
      nodeId: 'other-root-n0',
      executionId: { project: 'flytesnacks', domain: 'development', name: 'rnktdb3skr' },
    },
    closure: {
      phase: 3,
      startedAt: { seconds: { low: 1642627611, high: 0, unsigned: false }, nanos: 0 },
      duration: { seconds: { low: 55, high: 0, unsigned: false }, nanos: 0 },
      createdAt: { seconds: { low: 1642627611, high: 0, unsigned: false }, nanos: 0 },
      updatedAt: { seconds: { low: 1642627666, high: 0, unsigned: false }, nanos: 0 },
      outputUri:
        's3://flyte-demo/metadata/propeller/flytesnacks-development-rnktdb3skr/other-root-n0/data/0/outputs.pb',
    },
    metadata: { isParentNode: true, specNodeId: 'other-root-n0' },
    scopedId: 'other-root-n0',
  },
};

const getMockNodeExecution = (
  initialStartSec: number,
  phase: NodeExecutionPhase,
  startOffsetSec: number,
  durationSec: number,
  cacheStatus?: CatalogCacheStatus,
) => {
  const node = { ...dNodeBasicExecution } as any;
  node.execution.closure.phase = phase;
  if (cacheStatus) {
    node.execution.closure = {
      ...node.execution.closure,
      taskNodeMetadata: {
        cacheStatus: cacheStatus,
      },
    };
    if (cacheStatus === CatalogCacheStatus.CACHE_HIT) {
      node.execution.closure.createdAt.seconds.low = initialStartSec + startOffsetSec;
      node.execution.closure.updatedAt.seconds.low = initialStartSec + startOffsetSec + durationSec;
      return {
        ...node,
        execution: {
          ...node.execution,
          closure: {
            ...node.execution.closure,
            startedAt: undefined,
            duration: undefined,
          },
        },
      };
    }
  }
  node.execution.closure.startedAt.seconds.low = initialStartSec + startOffsetSec;
  node.execution.closure.duration.seconds.low = initialStartSec + startOffsetSec + durationSec;
  return node;
};

export const mockbarItems = [
  { phase: NodeExecutionPhase.FAILED, startOffsetSec: 0, durationSec: 15, isFromCache: false },
  { phase: NodeExecutionPhase.SUCCEEDED, startOffsetSec: 5, durationSec: 11, isFromCache: true },
  { phase: NodeExecutionPhase.RUNNING, startOffsetSec: 17, durationSec: 23, isFromCache: false },
  { phase: NodeExecutionPhase.QUEUED, startOffsetSec: 39, durationSec: 0, isFromCache: false },
];

export const getMockExecutionsForBarChart = (startTimeSec: number) => {
  const start = startTimeSec;
  return [
    getMockNodeExecution(start, NodeExecutionPhase.FAILED, 0, 15),
    getMockNodeExecution(start, NodeExecutionPhase.SUCCEEDED, 5, 11, CatalogCacheStatus.CACHE_HIT),
    getMockNodeExecution(start, NodeExecutionPhase.RUNNING, 17, 23, CatalogCacheStatus.CACHE_MISS),
    getMockNodeExecution(start, NodeExecutionPhase.QUEUED, 39, 0),
  ];
};
