import { createLocalizedString } from 'basics/Locale';

const str = {
  allExecutionsTitle: 'All Executions in the Project',
  last100ExecutionsTitle: 'Last 100 Executions in the Project',
  tasksTotal: (n: number) => `${n} Tasks`,
  workflowsTotal: (n: number) => `${n} Workflows`,
};

export { patternKey } from 'basics/Locale';
export default createLocalizedString(str);
