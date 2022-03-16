import { storiesOf } from '@storybook/react';
import { resolveAfter } from 'common/promiseUtils';
import { createMockWorkflow, createMockWorkflowVersions } from 'models/__mocks__/workflowData';
import { WorkflowId } from 'models/Workflow/types';
import * as moment from 'moment';
import * as React from 'react';
import { SearchableSelector, SearchableSelectorOption } from '../SearchableSelector';

const mockWorkflow = createMockWorkflow('MyWorkflow');
const mockWorkflowVersions = createMockWorkflowVersions(mockWorkflow.id.name, 10);

const options = mockWorkflowVersions.map<SearchableSelectorOption<WorkflowId>>((wf, index) => ({
  data: wf.id,
  id: wf.id.version,
  name: wf.id.version,
  description: index === 0 ? 'Latest' : moment().subtract(index, 'days').format('DD MMM YYYY'),
}));

const stories = storiesOf('Launch/WorkflowSelector', module);

stories.addDecorator((story) => {
  return <div style={{ width: 600, height: '95vh' }}>{story()}</div>;
});

stories.add('Basic', () => {
  const [selectedItem, setSelectedItem] = React.useState(options[0]);
  const fetch = (query: string) =>
    resolveAfter(
      500,
      options.filter(({ name }) => name.includes(query)),
    );

  return (
    <SearchableSelector
      label="Workflow Version"
      fetchSearchResults={fetch}
      onSelectionChanged={setSelectedItem}
      selectedItem={selectedItem}
      options={options}
    />
  );
});
