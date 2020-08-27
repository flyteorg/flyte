import { makeStyles, Theme } from '@material-ui/core/styles';
import { action } from '@storybook/addon-actions';
import { storiesOf } from '@storybook/react';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { APIContext } from 'components/data/apiContext';
import { createMockExecutionEntities } from 'components/Executions/__mocks__/createMockExecutionEntities';
import { ExecutionDataCacheContext } from 'components/Executions/contexts';
import { createExecutionDataCache } from 'components/Executions/useExecutionDataCache';
import { fetchStates } from 'components/hooks/types';
import { keyBy } from 'lodash';
import { createMockTaskExecutionForNodeExecution } from 'models/Execution/__mocks__/mockTaskExecutionsData';
import * as React from 'react';
import { State } from 'xstate';
import {
    NodeExecutionsTable,
    NodeExecutionsTableProps
} from '../NodeExecutionsTable';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        borderLeft: `1px solid ${theme.palette.grey[400]}`,
        display: 'flex',
        height: '100vh',
        padding: `${theme.spacing(2)}px 0`,
        width: '100vw'
    }
}));

const {
    nodes,
    nodeExecutions,
    workflow,
    workflowExecution
} = createMockExecutionEntities({
    workflowName: 'SampleWorkflow',
    nodeExecutionCount: 10
});

const nodesById = keyBy(nodes, n => n.id);
const nodesWithChildren = {
    [nodes[0].id]: true,
    [nodes[1].id]: true
};
const nodeRetryAttempts = {
    [nodes[1].id]: 2
};

const apiContext = mockAPIContextValue({
    getExecution: () => Promise.resolve(workflowExecution),
    getNodeExecutionData: () => Promise.resolve({ inputs: {}, outputs: {} }),
    listTaskExecutions: nodeExecutionId => {
        const length = nodeRetryAttempts[nodeExecutionId.nodeId] || 1;
        const entities = Array.from({ length }, (_, retryAttempt) =>
            createMockTaskExecutionForNodeExecution(
                nodeExecutionId,
                nodesById[nodeExecutionId.nodeId],
                retryAttempt,
                { isParent: !!nodesWithChildren[nodeExecutionId.nodeId] }
            )
        );
        return Promise.resolve({ entities });
    },
    listTaskExecutionChildren: ({ retryAttempt }) =>
        Promise.resolve({
            entities: nodeExecutions.slice(0, 2).map(ne => ({
                ...ne,
                id: {
                    ...ne.id,
                    nodeId: `${ne.id.nodeId}_${retryAttempt}`
                }
            }))
        }),
    getWorkflow: () => Promise.resolve(workflow)
});
const dataCache = createExecutionDataCache(apiContext);
dataCache.insertWorkflow(workflow);
dataCache.insertWorkflowExecutionReference(workflowExecution.id, workflow.id);

const fetchAction = action('fetch');

const props: NodeExecutionsTableProps = {
    value: nodeExecutions,
    lastError: null,
    state: State.from(fetchStates.LOADED),
    moreItemsAvailable: false,
    fetch: () => Promise.resolve(() => fetchAction() as unknown)
};

const stories = storiesOf('Tables/NodeExecutionsTable', module);
stories.addDecorator(story => {
    return (
        <APIContext.Provider value={apiContext}>
            <ExecutionDataCacheContext.Provider value={dataCache}>
                <div className={useStyles().container}>{story()}</div>
            </ExecutionDataCacheContext.Provider>
        </APIContext.Provider>
    );
});
stories.add('Basic', () => <NodeExecutionsTable {...props} />);
stories.add('With no items', () => (
    <NodeExecutionsTable {...props} value={[]} />
));
