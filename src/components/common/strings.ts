import { createLocalizedString } from 'basics/Locale';

const str = {
  annotationsHeader: 'Annotations',
  collapseButton: (showAll: boolean) => (showAll ? 'Collapse' : 'Expand'),
  domainSettingsTitle: 'Domain Settings',
  iamRoleHeader: 'IAM Role',
  inherited: 'Inherits from project level values',
  labelsHeader: 'Labels',
  maxParallelismHeader: 'Max parallelism',
  rawDataHeader: 'Raw output data config',
  securityContextHeader: 'Security Context',
  serviceAccountHeader: 'Service Account',
};

export { patternKey } from 'basics/Locale';
export default createLocalizedString(str);
