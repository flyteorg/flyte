import * as React from 'react';
import { storiesOf } from '@storybook/react';
import { action } from '@storybook/addon-actions';
import { SystemStatus } from 'models/Common/types';
import { RenderSystemStatusBanner } from '../SystemStatusBanner';

const normalStatus: SystemStatus = {
  status: 'normal',
  message: 'This is a test. It is only a test. Check out https://flyte.org for more information.'
};

// Max length for a status should be 500 characters
const longTextStatus: SystemStatus = {
  status: 'normal',
  message:
    'Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibu'
};

const degradedStatus: SystemStatus = {
  status: 'degraded',
  message: 'Something is a bit wrong with the system. We can probably still do some things, but may be slow about it.'
};

const downStatus: SystemStatus = {
  status: 'down',
  message: 'We are down. You should probably go get a cup of coffee while the administrators attempt to troubleshoot.'
};

const renderBanner = (status: SystemStatus) => (
  <RenderSystemStatusBanner systemStatus={status} onClose={action('onClose')} />
);

const stories = storiesOf('Notifications/SystemStatusBanner', module);
stories.add('Normal Status', () => renderBanner(normalStatus));
stories.add('Degraded Status', () => renderBanner(degradedStatus));
stories.add('Down Status', () => renderBanner(downStatus));
stories.add('Long Message', () => renderBanner(longTextStatus));
