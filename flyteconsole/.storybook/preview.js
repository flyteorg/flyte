import React from 'react';
import { StorybookContainer } from './StorybookContainer';

//👇 Configures Storybook to log 'onXxx' actions (example: onArchiveTask and onPinTask ) in the UI
export const parameters = {
  actions: { argTypesRegex: '^on[A-Z].*' },
  controls: {
    matchers: {
      color: /(background|color)$/i,
      date: /Date$/,
    },
  },
  options: {
    storySort: {
      order: ['Primitives', 'Common'],
    },
  },
};

export const decorators = [
  (Story) => (
    <StorybookContainer>
      <Story />
    </StorybookContainer>
  ),
];
