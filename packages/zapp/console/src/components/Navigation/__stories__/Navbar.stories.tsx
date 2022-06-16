import * as React from 'react';
import { ComponentStory, ComponentMeta } from '@storybook/react';
import { FeatureFlag, useFeatureFlagContext } from 'basics/FeatureFlags';
import { ExecutionContext } from 'components/Executions/contexts';
import { ExecutionDetailsAppBarContent } from 'components/Executions/ExecutionDetails/ExecutionDetailsAppBarContent';
import { mockExecution } from 'models/Execution/__mocks__/mockWorkflowExecutionsData';
import { createMockExecution } from 'models/__mocks__/executionsData';
import { NavBar } from '../NavBar';
import { FlyteNavigation, FlyteNavItem } from '../utils';

export default {
  title: 'Navigation/NavBar',
  component: NavBar,
} as ComponentMeta<typeof NavBar>;

// Helper to let us delay rendering of custom content until after the first
// mount of the app bar, so that the target element will exist.
const DelayedMount: React.FC<React.Props<{}>> = ({ children }) => {
  const [showContent, setShowContent] = React.useState(false);
  React.useEffect(() => setShowContent(true), []);
  return showContent ? <>{children}</> : null;
};

const removeCustomToggle = {
  // Remove useCustomContent toggle from UI
  useCustomContent: {
    table: {
      disable: true,
    },
  },
};
const removeNavData = {
  // Remove useCustomContent toggle from UI
  navigationData: {
    table: {
      disable: true,
    },
  },
};

/* ****************************************
 * Stories
 **************************************** */

const Template: ComponentStory<typeof NavBar> = (props) => <NavBar {...props} />;
export const Default = Template.bind({});
Default.argTypes = { ...removeCustomToggle, ...removeNavData };

export const ExecutionDetails = Template.bind({});
ExecutionDetails.argTypes = removeNavData;
ExecutionDetails.args = {
  useCustomContent: true,
};
const execution = createMockExecution();
const executionContext = { execution };
ExecutionDetails.decorators = [
  (Story, context) => {
    return (
      <>
        {Story()}
        <DelayedMount>
          {/* allow to switch between two views */}
          {context.args.useCustomContent ? (
            <ExecutionContext.Provider value={executionContext}>
              <ExecutionDetailsAppBarContent execution={mockExecution} />
            </ExecutionContext.Provider>
          ) : null}
        </DelayedMount>
      </>
    );
  },
];

export const OnlyMine = Template.bind({});
OnlyMine.argTypes = { ...removeCustomToggle, ...removeNavData };
OnlyMine.decorators = [
  (Story) => {
    const { setFeatureFlag } = useFeatureFlagContext();
    React.useEffect(() => {
      console.log('Set');
      setFeatureFlag(FeatureFlag.OnlyMine, true);
      return () => {
        console.log('Clean');
        setFeatureFlag(FeatureFlag.OnlyMine, false);
      };
    }, []);

    return Story();
  },
];

/* ****************************************
 * Show Component Coloring
 **************************************** */
const ColoredNavBar = ({
  useCustomContent,
  color,
  background,
  items,
}: {
  useCustomContent: boolean;
  color: string;
  background: string;
  items: FlyteNavItem[];
}) => {
  const navigationData: FlyteNavigation = {
    color,
    background,
    items,
  };
  return (
    <>
      <NavBar useCustomContent={useCustomContent} navigationData={navigationData} />
      <DelayedMount>
        {/* allow to switch between two views */}
        {useCustomContent ? (
          <ExecutionContext.Provider value={executionContext}>
            <ExecutionDetailsAppBarContent execution={mockExecution} />
          </ExecutionContext.Provider>
        ) : null}
      </DelayedMount>
    </>
  );
};
const FullTemplate: ComponentStory<typeof ColoredNavBar> = (props) => <ColoredNavBar {...props} />;
export const WithColor = FullTemplate.bind({});
WithColor.argTypes = removeNavData as any;
WithColor.args = {
  useCustomContent: false,
  color: 'white',
  background: 'black',
  items: [
    { title: 'Dashboard', url: '/projects/flytesnacks/executions?domain=development&duration=all' },
  ],
};
