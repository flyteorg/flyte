import { makeStyles, Theme } from '@material-ui/core/styles';
import { action } from '@storybook/addon-actions';
import { storiesOf } from '@storybook/react';
import * as React from 'react';
import { ExecutionFilters } from '../ExecutionFilters';
import { useWorkflowExecutionFiltersState } from '../filters/useExecutionFiltersState';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        borderLeft: `1px solid ${theme.palette.grey[400]}`,
        display: 'flex',
        height: '100vh',
        padding: `${theme.spacing(2)}px 0`,
        width: '100vw'
    }
}));

const changeAction = action('change');

const stories = storiesOf('Tables/ExecutionFilters', module);
stories.addDecorator(story => (
    <div className={useStyles().container}>{story()}</div>
));
stories.add('Basic', () => (
    <ExecutionFilters {...useWorkflowExecutionFiltersState()} />
));
