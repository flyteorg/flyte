import { createLocalizedString } from '.';

const str = {
  regularString: 'Regular string',
  value_: 'Default value',
  value_on: 'Value ON',
  value_off: 'Value OFF',
  activeUsers: (active: number, all: number) => `${active} out of ${all} users`,
};

export { patternKey } from '.';
export default createLocalizedString(str);
