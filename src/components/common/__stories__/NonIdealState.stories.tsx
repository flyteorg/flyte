import Button from '@material-ui/core/Button';
import ErrorOutline from '@material-ui/icons/ErrorOutline';
import * as React from 'react';

import { action } from '@storybook/addon-actions';
import { storiesOf } from '@storybook/react';

import { NonIdealState } from '../NonIdealState';

const baseProps = {
  icon: ErrorOutline,
  title: 'Things are not as they should be.',
  description: 'If this were an actual problem, here is where we would tell you all the details of what happened.'
};

const actionButton = (
  <Button variant="contained" color="primary" onClick={action('action')}>
    Try again
  </Button>
);

const stories = storiesOf('Common/NonIdealState', module);
stories.add('small', () => (
  <NonIdealState {...baseProps} size="small">
    {actionButton}
  </NonIdealState>
));
stories.add('medium', () => (
  <NonIdealState {...baseProps} size="medium">
    {actionButton}
  </NonIdealState>
));
stories.add('large', () => (
  <NonIdealState {...baseProps} size="large">
    {actionButton}
  </NonIdealState>
));
