import { timestampToDate } from 'common/utils';
import { CatalogCacheStatus, NodeExecutionPhase } from 'models/Execution/enums';
import { dNode } from 'models/Graph/types';
import { BarItemData } from './utils';

const WEEK_DURATION_SEC = 7 * 24 * 3600;

const EMPTY_BAR_ITEM: BarItemData = {
  phase: NodeExecutionPhase.UNDEFINED,
  startOffsetSec: 0,
  durationSec: 0,
  isFromCache: false,
};

export const getChartDurationData = (
  nodes: dNode[],
  startedAt: Date,
): { items: BarItemData[]; totalDurationSec: number } => {
  if (nodes.length === 0) return { items: [], totalDurationSec: 0 };

  let totalDurationSec = 0;
  const initialStartTime = startedAt.getTime();
  const result: BarItemData[] = nodes.map(({ execution }) => {
    if (!execution) {
      return EMPTY_BAR_ITEM;
    }

    let phase = execution.closure.phase;
    const isFromCache =
      execution.closure.taskNodeMetadata?.cacheStatus === CatalogCacheStatus.CACHE_HIT;

    // Offset values
    let startOffset = 0;
    const startedAt = execution.closure.startedAt;
    if (isFromCache) {
      if (execution.closure.createdAt) {
        startOffset = timestampToDate(execution.closure.createdAt).getTime() - initialStartTime;
      }
    } else if (startedAt) {
      startOffset = timestampToDate(startedAt).getTime() - initialStartTime;
    }

    // duration
    let durationSec = 0;
    if (isFromCache) {
      const updatedAt = execution.closure.updatedAt?.seconds?.toNumber() ?? 0;
      const createdAt = execution.closure.createdAt?.seconds?.toNumber() ?? 0;
      durationSec = updatedAt - createdAt;
      durationSec = durationSec === 0 ? 2 : durationSec;
    } else if (phase === NodeExecutionPhase.RUNNING) {
      if (startedAt) {
        const duration = Date.now() - timestampToDate(startedAt).getTime();
        durationSec = duration / 1000;
        if (durationSec > WEEK_DURATION_SEC) {
          // TODO: https://github.com/flyteorg/flyteconsole/issues/332
          // In some cases tasks which were needed to be ABORTED are stuck in running state,
          // In case if task is still running after a week - we assume it should have been aborted.
          // The proper fix should be covered by isue: flyteconsole#332
          phase = NodeExecutionPhase.ABORTED;
          const allegedDurationSec = Math.trunc(totalDurationSec - startOffset / 1000);
          durationSec = allegedDurationSec > 0 ? allegedDurationSec : 10;
        }
      }
    } else {
      durationSec = execution.closure.duration?.seconds?.toNumber() ?? 0;
    }

    const startOffsetSec = Math.trunc(startOffset / 1000);
    totalDurationSec = Math.max(totalDurationSec, startOffsetSec + durationSec);
    return { phase, startOffsetSec, durationSec, isFromCache };
  });

  // Do we want to get initialStartTime from different place, to avoid recalculating it.
  return { items: result, totalDurationSec };
};
