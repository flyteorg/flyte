import { createLocalizedString } from '@flyteconsole/locale';

const str = {
  noLogsFoundString: 'No logs found',
  aborted: 'Aborted',
  aborting: 'Aborting',
  failed: 'Failed',
  failing: 'Failing',
  waiting: 'Waiting',
  queued: 'Queued',
  skipped: 'Skipped',
  recovered: 'Recovered',
  initializing: 'Initializing',
  running: 'Running',
  succeeded: 'Succeeded',
  succeeding: 'Succeeding',
  timedOut: 'Timed Out',
  paused: 'Paused',
  unknown: 'Unknown',
  cacheDisabledMessage: 'Caching was disabled for this execution.',
  cacheHitMessage: 'Output for this execution was read from cache.',
  cacheLookupFailureMessage: 'Failed to lookup cache information.',
  cacheMissMessage: 'No cached output was found for this execution.',
  cachePopulatedMessage: 'The result of this execution was written to cache.',
  cachePutFailure: 'Failed to write output for this execution to cache.',
  mapCacheMessage: "Check the detail panel for each task's cache status.",
  unknownCacheStatusString: 'Cache status is unknown',
  viewSourceExecutionString: 'View source execution',
  fromCache: 'From cache',
  readFromCache: 'Read from cache',
};

export { patternKey } from '@flyteconsole/locale';
export default createLocalizedString(str);
