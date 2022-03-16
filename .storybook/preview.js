import React from 'react';
import { StorybookContainer } from './StorybookContainer';

export const parameters = {
  actions: { argTypesRegex: '^on[A-Z].*' },
  controls: {
    matchers: {
      color: /(background|color)$/i,
      date: /Date$/,
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
