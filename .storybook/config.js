// import { setOptions } from '@storybook/addon-options';
import { addDecorator, configure } from '@storybook/react';
import StoryRouter from 'storybook-react-router';
import { withKnobs } from '@storybook/addon-knobs';

import * as React from 'react';

import { StorybookContainer } from './StorybookContainer';

addDecorator(withKnobs);
addDecorator(StoryRouter());
addDecorator(story => React.createElement(StorybookContainer, null, story()));

const req = require.context(
    '../src/components',
    true,
    /\.stories\.(ts|tsx|js|jsx)$/
);

function loadStories() {
    require('../src/protobuf');
    req.keys().forEach(path => req(path));
}

configure(loadStories, module);
