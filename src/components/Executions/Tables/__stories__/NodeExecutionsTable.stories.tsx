import { makeStyles, Theme } from '@material-ui/core/styles';
import { action } from '@storybook/addon-actions';
import { storiesOf } from '@storybook/react';
import axios from 'axios';
// tslint:disable-next-line:import-name
import AxiosMockAdapter from 'axios-mock-adapter';
import { mapNodeExecutionDetails } from 'components/Executions/utils';
import { NavBar } from 'components/Navigation';
import { Admin } from 'flyteidl';
import { encodeProtoPayload } from 'models';
import {
    createMockWorkflow,
    createMockWorkflowClosure
} from 'models/__mocks__/workflowData';
import { createMockNodeExecutions } from 'models/Execution/__mocks__/mockNodeExecutionsData';
import { createMockTaskExecutionsListResponse } from 'models/Execution/__mocks__/mockTaskExecutionsData';
import { mockTasks } from 'models/Task/__mocks__/mockTaskData';
import * as React from 'react';
import {
    NodeExecutionsTable,
    NodeExecutionsTableProps
} from '../NodeExecutionsTable';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        borderLeft: `1px solid ${theme.palette.grey[400]}`,
        display: 'flex',
        height: '100vh',
        paddingTop: theme.spacing(10),
        width: '100vw'
    }
}));

const {
    executions: mockExecutions,
    nodes: mockNodes
} = createMockNodeExecutions(10);

const mockWorkflow = createMockWorkflow('SampleWorkflow');
const mockWorkflowClosure = createMockWorkflowClosure();
const compiledWorkflow = mockWorkflowClosure.compiledWorkflow!;
const {
    primary: { template },
    tasks
} = compiledWorkflow;
template.nodes = template.nodes.concat(mockNodes);
compiledWorkflow.tasks = tasks.concat(mockTasks);
mockWorkflow.closure = mockWorkflowClosure;

const fetchAction = action('fetch');

const props: NodeExecutionsTableProps = {
    value: mapNodeExecutionDetails(mockExecutions, mockWorkflow),
    lastError: null,
    loading: false,
    moreItemsAvailable: false,
    fetch: () => Promise.resolve(() => fetchAction() as unknown)
};

const stories = storiesOf('Tables/NodeExecutionsTable', module);
stories.addDecorator(story => {
    React.useEffect(() => {
        const executionsResponse = createMockTaskExecutionsListResponse(3);
        const mock = new AxiosMockAdapter(axios);
        mock.onGet(/.*\/task_executions\/.*/).reply(() => [
            200,
            encodeProtoPayload(executionsResponse, Admin.TaskExecutionList),
            { 'Content-Type': 'application/octet-stream' }
        ]);
        return () => mock.restore();
    });
    return (
        <>
            <NavBar />
            <div className={useStyles().container}>{story()}</div>
        </>
    );
});
stories.add('Basic', () => <NodeExecutionsTable {...props} />);
stories.add('With more items available', () => (
    <NodeExecutionsTable {...props} moreItemsAvailable={true} />
));
stories.add('With no items', () => (
    <NodeExecutionsTable {...props} value={[]} />
));
