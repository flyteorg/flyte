import { storiesOf } from '@storybook/react';
import { sampleWorkflowIds } from 'models/__mocks__/sampleWorkflowIds';
import * as React from 'react';
import { SearchableWorkflowIdList } from '../SearchableWorkflowIdList';

const baseProps = { workflowIds: [...sampleWorkflowIds] };

const stories = storiesOf('Workflow/SearchableWorkflowIdList', module);
stories.addDecorator(story => <div style={{ width: '650px' }}>{story()}</div>);
stories.add('basic', () => <SearchableWorkflowIdList {...baseProps} />);
