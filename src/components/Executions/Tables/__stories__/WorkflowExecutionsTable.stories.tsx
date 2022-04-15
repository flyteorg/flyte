import * as React from 'react';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { action } from '@storybook/addon-actions';
import { storiesOf } from '@storybook/react';
import { ExecutionState } from 'models/Execution/enums';
import { createMockWorkflowExecutionsListResponse } from 'models/Execution/__mocks__/mockWorkflowExecutionsData';
import { WorkflowExecutionsTable, WorkflowExecutionsTableProps } from '../WorkflowExecutionsTable';

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    borderLeft: `1px solid ${theme.palette.grey[400]}`,
    display: 'flex',
    height: '100vh',
    padding: `${theme.spacing(2)}px 0`,
    width: '100vw',
  },
}));

const fetchAction = action('fetch');

const propsArchived: WorkflowExecutionsTableProps = {
  value: createMockWorkflowExecutionsListResponse(10, ExecutionState.EXECUTION_ARCHIVED).executions,
  lastError: null,
  isFetching: false,
  moreItemsAvailable: false,
  fetch: () => Promise.resolve(() => fetchAction() as unknown),
};

const props: WorkflowExecutionsTableProps = {
  value: createMockWorkflowExecutionsListResponse(10, ExecutionState.EXECUTION_ACTIVE).executions,
  lastError: null,
  isFetching: false,
  moreItemsAvailable: false,
  fetch: () => Promise.resolve(() => fetchAction() as unknown),
};

const stories = storiesOf('Tables/WorkflowExecutionsTable', module);
stories.addDecorator((story) => <div className={useStyles().container}>{story()}</div>);
stories.add('Basic', () => <WorkflowExecutionsTable {...props} />);
stories.add('Only archived items', () => <WorkflowExecutionsTable {...propsArchived} />);
stories.add('With more items available', () => (
  <WorkflowExecutionsTable {...props} moreItemsAvailable={true} />
));
stories.add('With no items', () => <WorkflowExecutionsTable {...props} value={[]} />);
