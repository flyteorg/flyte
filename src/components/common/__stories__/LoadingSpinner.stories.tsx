import { storiesOf } from '@storybook/react';
import { basicStoryContainer } from '__stories__/decorators';
import * as React from 'react';

import { LoadingSpinner } from '../LoadingSpinner';

const stories = storiesOf('Common', module);
stories.addDecorator(basicStoryContainer);
stories.add('LoadingSpinner', () => (
    <div>
        <LoadingSpinner size="small" />
        <LoadingSpinner size="medium" />
        <LoadingSpinner size="large" />
    </div>
));
