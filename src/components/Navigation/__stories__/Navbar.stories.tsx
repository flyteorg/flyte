import { storiesOf } from '@storybook/react';
import { ExecutionContext } from 'components/Executions/contexts';
import { ExecutionDetailsAppBarContent } from 'components/Executions/ExecutionDetails/ExecutionDetailsAppBarContent';
import { mockExecution } from 'models/Execution/__mocks__/mockWorkflowExecutionsData';
import { createMockExecution } from 'models/__mocks__/executionsData';
import * as React from 'react';
import { FeatureFlag, useFeatureFlagContext } from 'basics/FeatureFlags';
import { NavBar } from '../NavBar';

// Helper to let us delay rendering of custom content until after the first
// mount of the app bar, so that the target element will exist.
const DelayedMount: React.FC<React.Props<{}>> = ({ children }) => {
  const [showContent, setShowContent] = React.useState(false);
  React.useEffect(() => setShowContent(true), []);
  return showContent ? <>{children}</> : null;
};

const execution = createMockExecution();
const executionContext = { execution };

const stories = storiesOf('Navigation/NavBar', module);
stories.add('default', () => <NavBar />);
stories.add('execution details', () => (
  <>
    <NavBar useCustomContent={true} />
    <DelayedMount>
      <ExecutionContext.Provider value={executionContext}>
        <ExecutionDetailsAppBarContent execution={mockExecution} />
      </ExecutionContext.Provider>
    </DelayedMount>
  </>
));
stories.add('only mine', () => {
  const { setFeatureFlag } = useFeatureFlagContext();
  React.useEffect(() => {
    setFeatureFlag(FeatureFlag.OnlyMine, true);
    return () => {
      setFeatureFlag(FeatureFlag.OnlyMine, false);
    };
  }, []);

  return <NavBar />;
});
