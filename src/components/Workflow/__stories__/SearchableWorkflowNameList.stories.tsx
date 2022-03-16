import { storiesOf } from '@storybook/react';
import { sampleWorkflowNames } from 'models/__mocks__/sampleWorkflowNames';
import * as React from 'react';
import { SearchableWorkflowNameList } from '../SearchableWorkflowNameList';

const baseProps = { workflows: [...sampleWorkflowNames] };

const stories = storiesOf('Workflow/SearchableWorkflowNameList', module);
stories.addDecorator((story) => <div style={{ width: '650px' }}>{story()}</div>);
stories.add('basic', () => <SearchableWorkflowNameList {...baseProps} />);
