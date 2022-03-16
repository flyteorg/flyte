import * as React from 'react';

import { action } from '@storybook/addon-actions';
import { storiesOf } from '@storybook/react';

import { ExpandableContentLink } from '../ExpandableContentLink';

const baseProps = {
  collapsedText: 'show',
  expandedText: 'hide',
  onExpand: action('expand'),
  renderContent: () => <p>This is the content.</p>,
};

const stories = storiesOf('Common/ExpandableContentLink', module);
stories.add('as link', () => <ExpandableContentLink {...baseProps} />);
stories.add('as button', () => <ExpandableContentLink button={true} {...baseProps} />);
