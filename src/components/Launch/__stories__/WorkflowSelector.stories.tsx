import { storiesOf } from '@storybook/react';
import { resolveAfter } from 'common/promiseUtils';
import { createMockFetchable } from 'components/hooks/__mocks__/fetchableData';
import {
    createMockWorkflow,
    createMockWorkflowVersions
} from 'models/__mocks__/workflowData';
import * as moment from 'moment';
import * as React from 'react';
import { WorkflowSelector, WorkflowSelectorOption } from '../WorkflowSelector';

const mockWorkflow = createMockWorkflow('MyWorkflow');
const mockWorkflowVersions = createMockWorkflowVersions(
    mockWorkflow.id.name,
    10
);

const options = mockWorkflowVersions.map<WorkflowSelectorOption>(
    (wf, index) => ({
        data: wf.id,
        id: wf.id.version,
        name: wf.id.version,
        description:
            index === 0
                ? 'Latest'
                : moment()
                      .subtract(index, 'days')
                      .format('DD MMM YYYY')
    })
);

const stories = storiesOf('Launch/WorkflowSelector', module);

stories.addDecorator(story => {
    return <div style={{ width: 600, height: '95vh' }}>{story()}</div>;
});

stories.add('Basic', () => {
    const [searchValue, setSearchValue] = React.useState<string>();
    const [selectedItem, setSelectedItem] = React.useState(options[0]);

    const fetch = () => resolveAfter(1000, []);
    const searchResults = createMockFetchable<WorkflowSelectorOption[]>(
        [],
        fetch
    );

    return (
        <WorkflowSelector
            searchResults={searchResults}
            onSelectionChanged={setSelectedItem}
            onSearchStringChanged={setSearchValue}
            searchValue={searchValue}
            selectedItem={selectedItem}
            options={options}
            workflowId={mockWorkflow.id}
        />
    );
});
