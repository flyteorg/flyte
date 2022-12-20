export const nodeExecutionsById = {
  n0: {
    id: {
      nodeId: 'n0',
      executionId: {
        project: 'flytesnacks',
        domain: 'development',
        name: 'fc027ce9fe4cf4f5eba8',
      },
    },
    inputUri:
      's3://flyte-demo/metadata/propeller/flytesnacks-development-fc027ce9fe4cf4f5eba8/n0/data/inputs.pb',
    closure: {
      phase: 3,
      startedAt: {
        seconds: {
          low: 1649888546,
          high: 0,
          unsigned: false,
        },
        nanos: 773100279,
      },
      duration: {
        seconds: {
          low: 22,
          high: 0,
          unsigned: false,
        },
        nanos: 800572640,
      },
      createdAt: {
        seconds: {
          low: 1649888546,
          high: 0,
          unsigned: false,
        },
        nanos: 697168683,
      },
      updatedAt: {
        seconds: {
          low: 1649888569,
          high: 0,
          unsigned: false,
        },
        nanos: 573672640,
      },
      outputUri:
        's3://flyte-demo/metadata/propeller/flytesnacks-development-fc027ce9fe4cf4f5eba8/n0/data/0/outputs.pb',
    },
    metadata: {
      specNodeId: 'n0',
    },
    scopedId: 'n0',
    logsByPhase: new Map([
      [
        3,
        [
          {
            logs: [
              {
                uri: 'http://localhost:30082/#!/log/flytesnacks-development/fc027ce9fe4cf4f5eba8-n0-0-0/pod?namespace=flytesnacks-development',
                name: 'Kubernetes Logs #0-0',
                messageFormat: 2,
              },
            ],
            externalId: 'fc027ce9fe4cf4f5eba8-n0-0-0',
            phase: 3,
          },
          {
            logs: [
              {
                uri: 'http://localhost:30082/#!/log/flytesnacks-development/fc027ce9fe4cf4f5eba8-n0-0-1/pod?namespace=flytesnacks-development',
                name: 'Kubernetes Logs #0-1',
                messageFormat: 2,
              },
            ],
            externalId: 'fc027ce9fe4cf4f5eba8-n0-0-1',
            index: 1,
            phase: 3,
          },
          {
            logs: [
              {
                uri: 'http://localhost:30082/#!/log/flytesnacks-development/fc027ce9fe4cf4f5eba8-n0-0-2/pod?namespace=flytesnacks-development',
                name: 'Kubernetes Logs #0-2',
                messageFormat: 2,
              },
            ],
            externalId: 'fc027ce9fe4cf4f5eba8-n0-0-2',
            index: 2,
            phase: 3,
          },
        ],
      ],
    ]),
  },
  n1: {
    id: {
      nodeId: 'n1',
      executionId: {
        project: 'flytesnacks',
        domain: 'development',
        name: 'fc027ce9fe4cf4f5eba8',
      },
    },
    inputUri:
      's3://flyte-demo/metadata/propeller/flytesnacks-development-fc027ce9fe4cf4f5eba8/n1/data/inputs.pb',
    closure: {
      phase: 3,
      startedAt: {
        seconds: {
          low: 1649888569,
          high: 0,
          unsigned: false,
        },
        nanos: 782695018,
      },
      duration: {
        seconds: {
          low: 9,
          high: 0,
          unsigned: false,
        },
        nanos: 811268323,
      },
      createdAt: {
        seconds: {
          low: 1649888569,
          high: 0,
          unsigned: false,
        },
        nanos: 685160925,
      },
      updatedAt: {
        seconds: {
          low: 1649888579,
          high: 0,
          unsigned: false,
        },
        nanos: 593963323,
      },
      outputUri:
        's3://flyte-demo/metadata/propeller/flytesnacks-development-fc027ce9fe4cf4f5eba8/n1/data/0/outputs.pb',
    },
    metadata: {
      specNodeId: 'n1',
    },
    scopedId: 'n1',
  },
};
