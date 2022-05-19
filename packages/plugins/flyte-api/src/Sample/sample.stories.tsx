import * as React from 'react';
import { ComponentStory, ComponentMeta } from '@storybook/react';
import { makeStyles, Theme } from '@material-ui/core/styles';

import { SampleComponent } from '.';

export default {
  title: 'Flyte-API/Sample',
  component: SampleComponent,
} as ComponentMeta<typeof SampleComponent>;

const useStyles = makeStyles((_theme: Theme) => ({
  updatedOne: {
    backgroundColor: 'lightblue',
    color: 'black',
  },
  updatedTwo: {
    backgroundColor: 'black',
    color: 'yellow',
  },
}));

const Template: ComponentStory<typeof SampleComponent> = () => <SampleComponent />;
export const Primary = Template.bind({});

export const Secondary: ComponentStory<typeof SampleComponent> = () => {
  const styles = useStyles();
  return <SampleComponent className={styles.updatedOne} />;
};

export const Tertiary: ComponentStory<typeof SampleComponent> = () => {
  const styles = useStyles();
  return <SampleComponent className={styles.updatedTwo} />;
};
