import { storiesOf } from '@storybook/react';
import { convertFlyteGraphToDAG } from 'models/Graph/convertFlyteGraphToDAG';
import { CompiledWorkflowClosure } from 'models/Workflow/types';
import * as React from 'react';
import { Graph } from '../Graph';
import * as batchTasks from './batchTasks.json';
import * as largeGraph from './largeGraph.json';
import * as rich from './rich.json';
import * as simple from './simple.json';

const simpleData = convertFlyteGraphToDAG(
    (simple as unknown) as CompiledWorkflowClosure
);
const batchData = convertFlyteGraphToDAG(
    (batchTasks as unknown) as CompiledWorkflowClosure
);
const richData = convertFlyteGraphToDAG(
    (rich as unknown) as CompiledWorkflowClosure
);
const largeData = convertFlyteGraphToDAG(
    (largeGraph as unknown) as CompiledWorkflowClosure
);

const stories = storiesOf('flytegraph/Graph', module);
stories.addDecorator(story => (
    <div style={{ position: 'absolute', top: 0, left: 0, right: 0, bottom: 0 }}>
        {story()}
    </div>
));
stories.add('simple', () => <Graph data={simpleData} />);
stories.add('batchTasks', () => <Graph data={batchData} />);
stories.add('rich', () => <Graph data={richData} />);
stories.add('large graph', () => <Graph data={largeData} />);
