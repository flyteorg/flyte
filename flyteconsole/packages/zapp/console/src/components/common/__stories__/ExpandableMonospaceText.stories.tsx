import { storiesOf } from '@storybook/react';
import { sampleError } from 'models/Execution/__mocks__/sampleExecutionError';
import * as React from 'react';
import { ExpandableMonospaceText } from '../ExpandableMonospaceText';

const baseProps = {
  text: sampleError,
};

const stories = storiesOf('Common', module);
stories.add('ExpandableMonospaceText', () => <ExpandableMonospaceText {...baseProps} />);
