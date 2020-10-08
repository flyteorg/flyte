import { env } from 'common/env';
import { InterpreterOptions } from 'xstate';

export const detailsPanelWidth = 432;

export const labels = {
    moreOptionsButton: 'Display more options',
    moreOptionsMenu: 'More options menu'
};

export const defaultStateMachineConfig: Partial<InterpreterOptions> = {
    devTools: env.NODE_ENV === 'development'
};
