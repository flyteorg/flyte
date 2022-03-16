import { storiesOf } from '@storybook/react';
import { sampleTaskNames } from 'models/__mocks__/sampleTaskNames';
import * as React from 'react';
import { SearchableTaskNameList } from '../SearchableTaskNameList';

const baseProps = { names: [...sampleTaskNames] };

const stories = storiesOf('Task/SearchableTaskNameList', module);
stories.addDecorator((story) => <div style={{ width: '650px' }}>{story()}</div>);
stories.add('basic', () => <SearchableTaskNameList {...baseProps} />);
