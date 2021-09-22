import { env } from 'common/env';
import { InterpreterOptions } from 'xstate';

export const detailsPanelWidth = 432;
export const loadingSpinnerDelayMs = 1000;

export const labels = {
    moreOptionsButton: 'Display more options',
    moreOptionsMenu: 'More options menu'
};

export const defaultStateMachineConfig: Partial<InterpreterOptions> = {
    devTools: env.NODE_ENV === 'development'
};

export const barChartColors = {
    default: '#e5e5e5',
    success: '#78dfb1',
    failure: '#f2a4ad'
};
