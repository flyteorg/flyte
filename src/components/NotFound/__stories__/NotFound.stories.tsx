import * as React from 'react';

import { storiesOf } from '@storybook/react';
import { withNavigation } from '__stories__/decorators';
import { NotFound } from '../NotFound';

const stories = storiesOf('Views', module);
stories.addDecorator(withNavigation);
stories.add('Not Found', () => <NotFound />);
