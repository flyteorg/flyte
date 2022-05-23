import { action } from '@storybook/addon-actions';
import { storiesOf } from '@storybook/react';
import { RENDER_ORDER } from 'components/Executions/TaskExecutionsList/constants';
import { NodeExecutionDisplayType } from 'components/Executions/types';
import { getTaskExecutionPhaseConstants } from 'components/Executions/utils';
import { CatalogCacheStatus, NodeExecutionPhase } from 'models/Execution/enums';
import { TaskType } from 'models/Task/constants';
import * as React from 'react';
import { ReactFlowProvider } from 'react-flow-renderer';
import { ReactFlowCustomTaskNode } from '../ReactFlow/customNodeComponents';

const stories = storiesOf('CustomNodes', module);
stories.addDecorator((story) => (
  <>
    <ReactFlowProvider>{story()}</ReactFlowProvider>
  </>
));

const commonData = {
  onNodeSelectionChanged: action('nodeSelected'),
  onPhaseSelectionChanged: action('phaseSelected'),
  scopedId: 'n0',
};

const taskData = {
  ...commonData,
  nodeType: NodeExecutionDisplayType.PythonTask,
  taskType: TaskType.PYTHON,
  cacheStatus: 0,
};

stories.add('Task Node', () => (
  <>
    {RENDER_ORDER.map((phase, i) => (
      <div
        style={{
          position: 'absolute',
          top: 40 * i + 20,
          left: 20,
        }}
        key={phase}
      >
        <ReactFlowCustomTaskNode
          data={{
            ...taskData,
            nodeExecutionStatus: phase,
            text: getTaskExecutionPhaseConstants(phase).text,
          }}
        />
      </div>
    ))}
  </>
));

const cachedTaskData = {
  ...commonData,
  nodeType: NodeExecutionDisplayType.PythonTask,
  nodeExecutionStatus: NodeExecutionPhase.SUCCEEDED,
  taskType: TaskType.PYTHON,
};

const CACHE_STATUSES = [
  { status: CatalogCacheStatus.CACHE_DISABLED, text: 'cache disabled' },
  { status: CatalogCacheStatus.CACHE_HIT, text: 'cache hit' },
  { status: CatalogCacheStatus.CACHE_LOOKUP_FAILURE, text: 'cache lookup failure' },
  { status: CatalogCacheStatus.CACHE_MISS, text: 'cache miss' },
  { status: CatalogCacheStatus.CACHE_POPULATED, text: 'cache populated' },
  { status: CatalogCacheStatus.CACHE_PUT_FAILURE, text: 'cache put failure' },
];

stories.add('Task Node by Cache Status', () => (
  <>
    {CACHE_STATUSES.map((cacheStatus, i) => (
      <div
        style={{
          position: 'absolute',
          top: 40 * i + 20,
          left: 20,
        }}
        key={cacheStatus.text}
      >
        <ReactFlowCustomTaskNode
          data={{
            ...cachedTaskData,
            cacheStatus: cacheStatus.status,
            text: cacheStatus.text,
          }}
        />
      </div>
    ))}
  </>
));

const logsByPhase = new Map();
logsByPhase.set(5, [
  {
    uri: '#',
    name: 'Kubernetes Logs #0-0-1',
    messageFormat: 2,
  },
  {
    uri: '#',
    name: 'Kubernetes Logs #0-1',
    messageFormat: 2,
  },
]);
logsByPhase.set(2, [
  {
    uri: '#',
    name: 'Kubernetes Logs #0-2-1',
    messageFormat: 2,
  },
]);
logsByPhase.set(6, [
  {
    name: 'ff21a6480a4c84742ad4-n0-0-3',
  },
]);

const mapTaskData = {
  ...commonData,
  nodeType: NodeExecutionDisplayType.MapTask,
  taskType: TaskType.ARRAY,
  cacheStatus: 0,
  nodeLogsByPhase: logsByPhase,
};

stories.add('Map Task Node', () => (
  <>
    {RENDER_ORDER.map((phase, i) => (
      <div
        style={{
          position: 'absolute',
          top: 45 * i + 20,
          left: 20,
        }}
        key={phase}
      >
        <ReactFlowCustomTaskNode
          data={{
            ...mapTaskData,
            nodeExecutionStatus: phase,
            text: getTaskExecutionPhaseConstants(phase).text,
          }}
        />
      </div>
    ))}
  </>
));
