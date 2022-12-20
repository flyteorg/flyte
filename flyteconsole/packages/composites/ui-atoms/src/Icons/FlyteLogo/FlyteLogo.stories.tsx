import * as React from 'react';
import { ComponentStory, ComponentMeta } from '@storybook/react';
import { makeStyles } from '@material-ui/core/styles';
import { FlyteLogo } from '.';

const useStyles = makeStyles(() => ({
  container: {
    padding: '12px',
    width: '100%',
  },
}));

interface WrapperProps {
  backgroundColor: 'aliceblue' | 'gray' | 'darkblue';
  size: number;
  children: React.ReactNode;
}
const ComponentWrapper = (props: WrapperProps) => {
  const styles = useStyles();

  const height = props.size + 2 * 12; // + 2*padding size
  return (
    <div
      className={styles.container}
      style={{ backgroundColor: props.backgroundColor, height: `${height}px` }}
    >
      {props.children}
    </div>
  );
};

export default {
  title: 'UI Atoms/FlyteLogo',
  component: FlyteLogo,
} as ComponentMeta<typeof FlyteLogo>;

const Template: ComponentStory<typeof FlyteLogo> = (props) => <FlyteLogo {...props} />;
export const All = Template.bind({});
All.args = {
  size: 42,
  hideText: false,
  background: 'light',
};
All.decorators = [
  (Story, context) => {
    return (
      <div style={{ display: 'flex', flexDirection: 'column' }}>
        <ComponentWrapper backgroundColor="aliceblue" size={context.args.size}>
          {Story()}
        </ComponentWrapper>
        <ComponentWrapper backgroundColor="gray" size={context.args.size}>
          {Story()}
        </ComponentWrapper>
        <ComponentWrapper backgroundColor="darkblue" size={context.args.size}>
          {Story()}
        </ComponentWrapper>
      </div>
    );
  },
];

export const Default = Template.bind({});
Default.args = {
  size: 42,
};
Default.decorators = [
  (Story, context) => {
    return (
      <ComponentWrapper backgroundColor="darkblue" size={context.args.size}>
        {Story()}
      </ComponentWrapper>
    );
  },
];

export const NoText = Template.bind({});
NoText.args = {
  size: 42,
  hideText: true,
};
NoText.decorators = [
  (Story, context) => {
    return (
      <ComponentWrapper backgroundColor="darkblue" size={context.args.size}>
        {Story()}
      </ComponentWrapper>
    );
  },
];
