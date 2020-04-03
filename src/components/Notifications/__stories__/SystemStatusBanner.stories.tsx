import { storiesOf } from '@storybook/react';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { APIContext } from 'components/data/apiContext';
import { SystemStatus } from 'models';
import * as React from 'react';
import { SystemStatusBanner } from '../SystemStatusBanner';

const normalStatus: SystemStatus = {
    status: 'normal',
    message:
        'This is a test. It is only a test. Check out https://flyte.org for more information.'
};

const degradedStatus: SystemStatus = {
    status: 'degraded',
    message:
        'Something is a bit wrong with the system. We can probably still do some things, but may be slow about it.'
};

const downStatus: SystemStatus = {
    status: 'down',
    message:
        'We are down. You should probably go get a cup of coffee while the administrators attempt to troubleshoot.'
};

function renderBanner(status: SystemStatus) {
    const mockApi = mockAPIContextValue({
        getSystemStatus: () => Promise.resolve(status)
    });
    return (
        <APIContext.Provider value={mockApi}>
            <SystemStatusBanner />
        </APIContext.Provider>
    );
}

const stories = storiesOf('Notifications/SystemStatusBanner', module);
stories.add('Normal Status', () => renderBanner(normalStatus));
stories.add('Degraded Status', () => renderBanner(degradedStatus));
stories.add('Down Status', () => renderBanner(downStatus));
