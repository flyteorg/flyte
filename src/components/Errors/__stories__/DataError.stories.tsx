import * as React from 'react';

import { action } from '@storybook/addon-actions';
import { storiesOf } from '@storybook/react';

import { DataError } from '../DataError';

const retryAction = action('retry');

const stories = storiesOf('Errors/DataError', module);
stories.add('Title Only', () => (
  <DataError errorTitle="Something went wrong" retry={retryAction} />
));

const error = new Error('A special thing we depended on did not go well.');
stories.add('Title and error', () => (
  <DataError errorTitle="Something went wrong" error={error} retry={retryAction} />
));
