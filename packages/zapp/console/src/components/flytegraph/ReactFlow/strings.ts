import { createLocalizedString } from '@flyteconsole/locale';

const str = {
  pausedTasksButton: 'Paused Tasks',
  legendButton: (isVisible: boolean) => `${isVisible ? 'Hide' : 'Show'} Legend`,
  resumeTooltip: 'Resume',
};

export { patternKey } from '@flyteconsole/locale';
export default createLocalizedString(str);
