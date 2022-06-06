import * as React from 'react';
import { ComponentStory, ComponentMeta } from '@storybook/react';

import { SampleComponent } from '.';

export default {
  title: 'Primitives/NavBar',
  component: SampleComponent,
} as ComponentMeta<typeof SampleComponent>;

export const Primary: ComponentStory<typeof SampleComponent> = () => {
  return <SampleComponent />;
};

const Template: ComponentStory<typeof SampleComponent> = () => <SampleComponent />;
export const Secondary = Template.bind({});
