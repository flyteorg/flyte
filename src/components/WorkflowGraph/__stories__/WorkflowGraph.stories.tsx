import { action } from '@storybook/addon-actions';
import { storiesOf } from '@storybook/react';
import { CacheContext } from 'components/Cache/CacheContext';
import { createCache } from 'components/Cache/createCache';
import { DetailsPanel } from 'components/common/DetailsPanel';
import { extractTaskTemplates } from 'components/hooks/utils';
import { CompiledWorkflowClosure, Workflow } from 'models/Workflow/types';
import * as React from 'react';
import { WorkflowGraph } from '../WorkflowGraph';
import * as graphData from './rich.json';

const graphDataClosure = (graphData as unknown) as CompiledWorkflowClosure;

const workflow: Workflow = {
    closure: { compiledWorkflow: graphDataClosure },
    id: {
        project: 'test',
        domain: 'test',
        name: 'test',
        version: '1'
    }
};

const onNodeSelectionChanged = action('nodeSelected');

const cache = createCache();
const taskTemplates = extractTaskTemplates(workflow);
cache.mergeArray(taskTemplates);

const stories = storiesOf('WorkflowGraph', module);
stories.addDecorator(story => (
    <>
        <div
            style={{
                position: 'absolute',
                top: 0,
                right: '35vw',
                left: 0,
                bottom: 0
            }}
        >
            <CacheContext.Provider value={cache}>
                {story()}
            </CacheContext.Provider>
        </div>
        <DetailsPanel />
    </>
));

stories.add('TaskNodeRenderer', () => (
    <WorkflowGraph
        onNodeSelectionChanged={onNodeSelectionChanged}
        workflow={workflow}
    />
));
