import { makeStyles, Theme } from '@material-ui/core/styles';
import { storiesOf } from '@storybook/react';
import { action } from '@storybook/addon-actions';
import * as React from 'react';

import { ExecutionFilters } from '../ExecutionFilters';
import {
  useWorkflowExecutionFiltersState,
  useNodeExecutionFiltersState,
} from '../filters/useExecutionFiltersState';

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    borderLeft: `1px solid ${theme.palette.grey[400]}`,
    display: 'flex',
    height: '100vh',
    padding: `${theme.spacing(2)}px 0`,
    width: '100vw',
  },
}));

const stories = storiesOf('Tables/ExecutionFilters', module);
stories.addDecorator((story) => <div className={useStyles().container}>{story()}</div>);
stories.add('Node executions', () => <ExecutionFilters {...useNodeExecutionFiltersState()} />);
stories.add('Workflow executions - all', () => (
  <ExecutionFilters
    {...useWorkflowExecutionFiltersState()}
    chartIds={['chart0']}
    clearCharts={action('clearCharts')}
    showArchived={false}
    onArchiveFilterChange={action('onArchiveFilterChange')}
  />
));
stories.add('Workflow executions - minimal', () => (
  <ExecutionFilters {...useWorkflowExecutionFiltersState()} />
));
