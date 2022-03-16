import { TaskType } from 'models/Task/constants';
import { COLOR_SPECTRUM } from './colorSpectrum';

export const bodyFontFamily = 'Lato, helvetica, arial, sans-serif';
export const headerFontFamily = '"Open Sans", helvetica, arial, sans-serif';

export const whiteColor = COLOR_SPECTRUM.white.color;
export const primaryColor = COLOR_SPECTRUM.purple60.color;
export const primaryLightColor = COLOR_SPECTRUM.purple30.color;
export const primaryDarkColor = COLOR_SPECTRUM.purple70.color;
export const secondaryColor = COLOR_SPECTRUM.indigo100.color;
export const secondaryBackgroundColor = COLOR_SPECTRUM.gray5.color;

export const primaryTextColor = COLOR_SPECTRUM.gray100.color;
export const secondaryTextColor = COLOR_SPECTRUM.gray60.color;
export const interactiveTextColor = COLOR_SPECTRUM.purple60.color;
export const interactiveTextDisabledColor = COLOR_SPECTRUM.purple30.color;
export const interactiveTextBackgroundColor = COLOR_SPECTRUM.purple5.color;
export const positiveTextColor = COLOR_SPECTRUM.mint60.color;
export const negativeTextColor = COLOR_SPECTRUM.sunset60.color;
export const mutedPrimaryTextColor = '#4A4A4A';

export const tableHeaderColor = COLOR_SPECTRUM.gray40.color;
export const tablePlaceholderColor = COLOR_SPECTRUM.gray40.color;

export const selectedActionColor = COLOR_SPECTRUM.gray10.color;

export const separatorColor = COLOR_SPECTRUM.gray15.color;
export const skeletonColor = COLOR_SPECTRUM.gray15.color;
export const skeletonHighlightColor = COLOR_SPECTRUM.gray0.color;
export const listhoverColor = COLOR_SPECTRUM.gray5.color;
export const nestedListColor = COLOR_SPECTRUM.gray0.color;
export const buttonHoverColor = COLOR_SPECTRUM.gray0.color;
export const inputFocusBorderColor = COLOR_SPECTRUM.blue60.color;

export const warningIconColor = COLOR_SPECTRUM.sunset60.color;
export const infoIconColor = COLOR_SPECTRUM.blue40.color;

export const dangerousButtonBorderColor = COLOR_SPECTRUM.red20.color;
export const dangerousButtonColor = COLOR_SPECTRUM.red30.color;
export const dangerousButtonHoverColor = COLOR_SPECTRUM.red40.color;
export const mutedButtonColor = COLOR_SPECTRUM.gray30.color;
export const mutedButtonHoverColor = COLOR_SPECTRUM.gray60.color;

export const errorBackgroundColor = '#FBFBFC';

export const workflowLabelColor = COLOR_SPECTRUM.gray25.color;

export const statusColors = {
  FAILURE: COLOR_SPECTRUM.red20.color,
  RUNNING: COLOR_SPECTRUM.blue20.color,
  QUEUED: COLOR_SPECTRUM.amber20.color,
  SUCCESS: COLOR_SPECTRUM.mint20.color,
  SKIPPED: COLOR_SPECTRUM.sunset20.color,
  UNKNOWN: COLOR_SPECTRUM.gray20.color,
  WARNING: COLOR_SPECTRUM.yellow40.color,
};

export type TaskColorMap = Record<TaskType, string>;
export const taskColors: TaskColorMap = {
  [TaskType.BATCH_HIVE]: '#E1E8ED',
  [TaskType.DYNAMIC]: '#E1E8ED',
  [TaskType.HIVE]: '#E1E8ED',
  [TaskType.PYTHON]: '#7157D9',
  [TaskType.SPARK]: '#00B3A4',
  [TaskType.ARRAY]: '#E1E8ED',
  [TaskType.SIDECAR]: '#E1E8ED',
  [TaskType.UNKNOWN]: '#E1E8ED',
  [TaskType.WAITABLE]: '#E1E8ED',
  [TaskType.MPI]: '#00B3A4',
};

export const bodyFontSize = '0.875rem';
export const smallFontSize = '0.75rem';
