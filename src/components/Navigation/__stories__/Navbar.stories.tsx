import { storiesOf } from '@storybook/react';
import { ExecutionDetailsAppBarContent } from 'components/Executions/ExecutionDetails/ExecutionDetailsAppBarContent';
import { mockExecution } from 'models/Execution/__mocks__/mockWorkflowExecutionsData';
import * as React from 'react';
import { NavBar } from '../NavBar';

// Helper to let us delay rendering of custom content until after the first
// mount of the app bar, so that the target element will exist.
const DelayedMount: React.FC<React.Props<{}>> = ({ children }) => {
    const [showContent, setShowContent] = React.useState(false);
    React.useEffect(() => setShowContent(true), []);
    return showContent ? <>{children}</> : null;
};

const stories = storiesOf('Navigation/NavBar', module);
stories.add('default', () => <NavBar />);
stories.add('execution details', () => (
    <>
        <NavBar useCustomContent={true} />
        <DelayedMount>
            <ExecutionDetailsAppBarContent execution={mockExecution} />
        </DelayedMount>
    </>
));
