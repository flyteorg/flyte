import { Protobuf, Event } from 'flyteidl';
import { MessageFormat, ResourceType, TaskLog } from 'models/Common/types';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { TaskExecution } from 'models/Execution/types';

import * as Long from 'long';
import { TaskType } from 'models/Task/constants';

// we probably will create a new helper function in future, to make testing/storybooks closer to what we see in API Json responses
const getProtobufTimestampFromIsoTime = (isoDateTime: string): Protobuf.ITimestamp => {
  const timeMs = Date.parse(isoDateTime);
  const timestamp = new Protobuf.Timestamp();
  timestamp.seconds = Long.fromInt(Math.floor(timeMs / 1000));
  timestamp.nanos = (timeMs % 1000) * 1e6;
  return timestamp;
};

const getProtobufDurationFromString = (durationSec: string): Protobuf.Duration => {
  const secondsInt = parseInt(durationSec, 10);
  const duration = new Protobuf.Duration();
  duration.seconds = Long.fromInt(secondsInt);
  return duration;
};

export const MockTaskExceutionLog: TaskLog = {
  uri: '#',
  name: 'Cloudwatch Logs (User)',
  messageFormat: MessageFormat.JSON,
};

export const MockPythonTaskExecution: TaskExecution = {
  id: {
    taskId: {
      resourceType: ResourceType.TASK,
      project: 'flytesnacks',
      domain: 'development',
      name: 'athena.workflows.example.say_hello',
      version: 'v13',
    },
    nodeExecutionId: {
      nodeId: 'ff65vi3y',
      executionId: {
        project: 'flytesnacks',
        domain: 'development',
        name: 'ogaayir2e3',
      },
    },
  },
  inputUri:
    's3://flyte-demo/metadata/propeller/flytesnacks-development-ogaayir2e3/athenaworkflowsexamplesayhello/data/inputs.pb',
  closure: {
    outputUri:
      's3://flyte-demo/metadata/propeller/flytesnacks-development-ogaayir2e3/athenaworkflowsexamplesayhello/data/0/outputs.pb',
    phase: TaskExecutionPhase.SUCCEEDED,
    logs: [MockTaskExceutionLog],
    createdAt: getProtobufTimestampFromIsoTime('2022-03-17T21:30:53.469624134Z'),
    updatedAt: getProtobufTimestampFromIsoTime('2022-03-17T21:31:04.011303736Z'),
    reason: 'task submitted to K8s',
    taskType: TaskType.PYTHON,
  },
};

export const getMockMapTaskLogItem = (
  phase: TaskExecutionPhase,
  hasLogs: boolean,
  index?: number,
  retryAttempt?: number,
): Event.IExternalResourceInfo => {
  const retryString = retryAttempt && retryAttempt > 0 ? `-${retryAttempt}` : '';
  return {
    externalId: `y286hpfvwh-n0-0-${index ?? 0}`,
    index: index,
    phase: phase,
    retryAttempt: retryAttempt,
    logs: hasLogs
      ? [
          {
            uri: '#',
            name: `Kubernetes Logs #0-${index ?? 0}${retryString} (State)`,
            messageFormat: MessageFormat.JSON,
          },
        ]
      : [],
  };
};

export const MockMapTaskExecution: TaskExecution = {
  id: {
    taskId: {
      resourceType: ResourceType.TASK,
      project: 'flytesnacks',
      domain: 'development',
      name: 'flyte.workflows.example.mapper_a_mappable_task_0',
      version: 'v2',
    },
    nodeExecutionId: {
      nodeId: 'n0',
      executionId: {
        project: 'flytesnacks',
        domain: 'development',
        name: 'y286hpfvwh',
      },
    },
  },
  inputUri:
    's3://my-s3-bucket/metadata/propeller/sandbox/flytesnacks-development-y286hpfvwh/n0/data/inputs.pb',
  closure: {
    outputUri:
      's3://my-s3-bucket/metadata/propeller/sandbox/flytesnacks-development-y286hpfvwh/n0/data/0/outputs.pb',
    phase: TaskExecutionPhase.SUCCEEDED,
    startedAt: getProtobufTimestampFromIsoTime('2022-03-30T19:31:09.487343Z'),
    duration: getProtobufDurationFromString('190.302384340s'),
    createdAt: getProtobufTimestampFromIsoTime('2022-03-30T19:31:09.487343693Z'),
    updatedAt: getProtobufTimestampFromIsoTime('2022-03-30T19:34:19.789727340Z'),
    taskType: 'container_array',
    metadata: {
      generatedName: 'y286hpfvwh-n0-0',
      externalResources: [
        getMockMapTaskLogItem(TaskExecutionPhase.SUCCEEDED, true),
        getMockMapTaskLogItem(TaskExecutionPhase.SUCCEEDED, true, 1),
        getMockMapTaskLogItem(TaskExecutionPhase.SUCCEEDED, true, 2),
        getMockMapTaskLogItem(TaskExecutionPhase.FAILED, true, 3),
        getMockMapTaskLogItem(TaskExecutionPhase.SUCCEEDED, true, 3, 1),
        getMockMapTaskLogItem(TaskExecutionPhase.SUCCEEDED, true, 4),
        getMockMapTaskLogItem(TaskExecutionPhase.FAILED, false, 5),
      ],
      pluginIdentifier: 'k8s-array',
    },
  },
};
