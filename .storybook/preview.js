import '../src/protobuf';
import { withKnobs } from '@storybook/addon-knobs';
import { addDecorator } from '@storybook/react';
import * as React from 'react';
import StoryRouter from 'storybook-react-router';
import { StorybookContainer } from './StorybookContainer';

addDecorator(withKnobs);
addDecorator(StoryRouter());
addDecorator(Story => <StorybookContainer><Story /></StorybookContainer>);

// Storybook executes this module in both bootstap phase (Node)
// and a story's runtime (browser). However, cannot call `setupWorker`
// in Node environment, so need to check if we're in a browser.
if (typeof global.process === 'undefined') {
    const { worker } = require('./worker');

    // Start the mocking when each story is loaded.
    // Repetitive calls to the `.start()` method will check for an existing
    // worker and reuse it.
    worker.start();
}
