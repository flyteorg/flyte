import * as React from 'react';
import { ComponentStory, ComponentMeta } from '@storybook/react';
import { makeStyles, Theme } from '@material-ui/core/styles';

import { NavBar } from '.';

export default {
  title: 'Primitives/NavBar',
  component: NavBar,
} as ComponentMeta<typeof NavBar>;

const useStyles = makeStyles((theme: Theme) => ({
  updatedOne: {
    backgroundColor: 'lightblue',
    color: 'black',
  },
  updatedTwo: {
    backgroundColor: 'black',
    color: 'yellow',
  },
}));

const Template: ComponentStory<typeof NavBar> = () => <NavBar />;
export const Primary = Template.bind({});

export const Secondary: ComponentStory<typeof NavBar> = () => {
  const styles = useStyles();
  return <NavBar className={styles.updatedOne} />;
};

export const Tertiary: ComponentStory<typeof NavBar> = () => {
  const styles = useStyles();
  return <NavBar className={styles.updatedTwo} />;
};
