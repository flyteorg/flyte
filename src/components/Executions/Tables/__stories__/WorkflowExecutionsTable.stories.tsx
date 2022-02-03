import { Collapse } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { action } from '@storybook/addon-actions';
import { storiesOf } from '@storybook/react';
import { ExecutionState } from 'models/Execution/enums';
import { createMockWorkflowExecutionsListResponse } from 'models/Execution/__mocks__/mockWorkflowExecutionsData';
import { SnackbarProvider } from 'notistack';
import * as React from 'react';
import {
    WorkflowExecutionsTable,
    WorkflowExecutionsTableProps
} from '../WorkflowExecutionsTable';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        borderLeft: `1px solid ${theme.palette.grey[400]}`,
        display: 'flex',
        height: '100vh',
        padding: `${theme.spacing(2)}px 0`,
        width: '100vw'
    }
}));

const fetchAction = action('fetch');

const propsArchived: WorkflowExecutionsTableProps = {
    value: createMockWorkflowExecutionsListResponse(
        10,
        ExecutionState.EXECUTION_ARCHIVED
    ).executions,
    lastError: null,
    isFetching: false,
    moreItemsAvailable: false,
    fetch: () => Promise.resolve(() => fetchAction() as unknown)
};

const props: WorkflowExecutionsTableProps = {
    value: createMockWorkflowExecutionsListResponse(
        10,
        ExecutionState.EXECUTION_ACTIVE
    ).executions,
    lastError: null,
    isFetching: false,
    moreItemsAvailable: false,
    fetch: () => Promise.resolve(() => fetchAction() as unknown)
};

// wrapper - to ensure that error/success notification shown as expected in storybook
const Wrapper = props => {
    return (
        <SnackbarProvider
            maxSnack={2}
            anchorOrigin={{ vertical: 'top', horizontal: 'right' }}
            TransitionComponent={Collapse}
        >
            {props.children}
        </SnackbarProvider>
    );
};

const stories = storiesOf('Tables/WorkflowExecutionsTable', module);
stories.addDecorator(story => (
    <div className={useStyles().container}>{story()}</div>
));
stories.add('Basic', () => (
    <Wrapper>
        <WorkflowExecutionsTable {...props} />
    </Wrapper>
));
stories.add('Only archived items', () => (
    <Wrapper>
        <WorkflowExecutionsTable {...propsArchived} />
    </Wrapper>
));
stories.add('With more items available', () => (
    <Wrapper>
        <WorkflowExecutionsTable {...props} moreItemsAvailable={true} />
    </Wrapper>
));
stories.add('With no items', () => (
    <Wrapper>
        <WorkflowExecutionsTable {...props} value={[]} />
    </Wrapper>
));
