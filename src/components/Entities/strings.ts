import { createLocalizedString } from 'basics/Locale';
import { startCase } from 'lodash';

const str = {
  workflowVersionsTitle: 'Recent Workflow Versions',
  viewAll: 'View All',
  schedulesHeader: 'Schedules',
  collapseButton: (showAll: boolean) => (showAll ? 'Collapse' : 'Expand'),
  launchStrings: (typeString: string) => `Launch ${startCase(typeString)}`,
  noDescription: (typeString: string) => `This ${typeString} has no description.`,
  noSchedules: (typeString: string) => `This ${typeString} has no schedules.`,
};

export { patternKey } from 'basics/Locale';
export default createLocalizedString(str);
