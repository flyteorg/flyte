import { ResourceType } from 'models/Common/types';

interface VersionsDetailsSectionsFlags {
  details: boolean;
  graph: boolean;
}

export const versionsDetailsSections: { [k in ResourceType]: VersionsDetailsSectionsFlags } = {
  [ResourceType.DATASET]: {
    details: false,
    graph: false,
  },
  [ResourceType.LAUNCH_PLAN]: {
    details: false,
    graph: false,
  },
  [ResourceType.TASK]: {
    details: true,
    graph: false,
  },
  [ResourceType.UNSPECIFIED]: {
    details: false,
    graph: false,
  },
  [ResourceType.WORKFLOW]: {
    details: false,
    graph: true,
  },
};
