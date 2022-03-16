import { createLocalizedString } from 'basics/Locale';
import { dateFromNow } from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { Protobuf } from 'flyteidl';

const str = {
  tableLabel_name: 'execution id',
  tableLabel_phase: 'status',
  tableLabel_startedAt: 'start time',
  tableLabel_duration: 'duration',
  tableLabel_actions: '',
  cancelAction: 'Cancel',
  inputOutputTooltip: 'View Inputs &amp; Outputs',
  launchPlanTooltip: 'View Launch Plan',
  archiveAction: (isArchived: boolean) => (isArchived ? 'Unarchive' : 'Archive'),
  archiveSuccess: (isArchived: boolean) =>
    `Item was successfully ${isArchived ? 'archived' : 'unarchived'}`,
  archiveError: (isArchived: boolean) =>
    `Error: Something went wrong, we can not ${isArchived ? 'archive' : 'unarchive'} item`,
  lastRunStartedAt: (startedAt?: Protobuf.ITimestamp) => {
    return startedAt ? `Last run ${dateFromNow(timestampToDate(startedAt))}` : '';
  },
};

export { patternKey } from 'basics/Locale';
export default createLocalizedString(str);
