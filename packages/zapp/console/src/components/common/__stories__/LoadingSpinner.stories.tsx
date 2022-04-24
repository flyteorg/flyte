import { storiesOf } from '@storybook/react';
import * as React from 'react';

import { LoadingSpinner } from '../LoadingSpinner';

const stories = storiesOf('Common', module);
stories.add('LoadingSpinner', () => (
  <div>
    <LoadingSpinner size="small" />
    <LoadingSpinner size="medium" />
    <LoadingSpinner size="large" />
  </div>
));
