import { Protobuf } from 'flyteidl';
import { MessageFormat, ResourceType } from 'models/Common/types';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { TaskExecution } from 'models/Execution/types';

import * as Long from 'long';

// we probably will create a new helper function in future, to make testing/storybooks closer to what we see in API Json responses
const getProtobufTimestampFromIsoTime = (isoDateTime: string): Protobuf.ITimestamp => {
  const timeMs = Date.parse(isoDateTime);
  const timestamp = new Protobuf.Timestamp();
  timestamp.seconds = Long.fromInt(Math.floor(timeMs / 1000));
  timestamp.nanos = (timeMs % 1000) * 1e6;
  return timestamp;
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
    logs: [
      {
        uri: 'https://console.aws.amazon.com/cloudwatch/home?region=us-east-2#logEventViewer:group=/aws/containerinsights/flyte-demo-2/application;stream=var.log.containers.ogaayir2e3-ff65vi3y-0_flytesnacks-development_ogaayir2e3-ff65vi3y-0-380d210ccaac45a6e2314a155822b36a67e044914069d01323bc18832487ac4a.log',
        name: 'Cloudwatch Logs (User)',
        messageFormat: MessageFormat.JSON,
      },
    ],
    createdAt: getProtobufTimestampFromIsoTime('2022-03-17T21:30:53.469624134Z'),
    updatedAt: getProtobufTimestampFromIsoTime('2022-03-17T21:31:04.011303736Z'),
    reason: 'task submitted to K8s',
    taskType: 'python-task',
  },
};
