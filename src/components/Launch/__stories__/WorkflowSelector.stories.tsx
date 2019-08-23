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
        id: wf.id.name,
        name: wf.id.name,
        description:
            index === 0
                ? 'latest'
                : moment()
                      .subtract(index, 'days')
                      .format('DD MMM YYYY')
    })
);

const stories = storiesOf('Launch/WorkflowSelector', module);

stories.addDecorator(story => {
    return <div style={{ width: 600, height: '95vh' }}>{story()}</div>;
});

const onSelectionChanged = (...args: any[]) => console.log(args);

stories.add('Basic', () => {
    const [searchValue, setSearchValue] = React.useState<string>();

    const fetch = () => resolveAfter(1000, []);
    const searchResults = createMockFetchable<WorkflowSelectorOption[]>(
        [],
        fetch
    );

    return (
        <WorkflowSelector
            searchResults={searchResults}
            onSelectionChanged={onSelectionChanged}
            onSearchStringChanged={setSearchValue}
            searchValue={searchValue}
            selectedItem={options[0]}
            options={options}
            workflowId={mockWorkflow.id}
        />
    );
});
