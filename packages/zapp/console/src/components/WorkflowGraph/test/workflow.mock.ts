import { Workflow } from 'models/Workflow/types';
import { long } from 'test/utils';

export const workflow: Workflow = {
  id: {
    resourceType: 2,
    project: 'flytesnacks',
    domain: 'development',
    name: 'core.control_flow.map_task.my_map_workflow',
    version: 'MT76cyUZeeYX-hA6qeIotA==',
  },
  closure: {
    compiledWorkflow: {
      subWorkflows: [],
      tasks: [
        {
          template: {
            config: {},
            id: {
              resourceType: 1,
              project: 'flytesnacks',
              domain: 'development',
              name: 'core.control_flow.map_task.coalesce',
              version: 'MT76cyUZeeYX-hA6qeIotA==',
            },
            type: 'python-task',
            metadata: {
              runtime: {
                type: 1,
                version: '0.0.0+develop',
                flavor: 'python',
              },
              retries: {},
            },
            interface: {
              inputs: {
                variables: {
                  b: {
                    type: {
                      collectionType: {
                        simple: 3,
                      },
                    },
                    description: 'b',
                  },
                },
              },
              outputs: {
                variables: {
                  o0: {
                    type: {
                      simple: 3,
                    },
                    description: 'o0',
                  },
                },
              },
            },
            container: {
              command: [],
              args: [
                'pyflyte-fast-execute',
                '--additional-distribution',
                's3://flyte-demo/zn/flytesnacks/development/MT76cyUZeeYX+hA6qeIotA==/scriptmode.tar.gz',
                '--dest-dir',
                '/root',
                '--',
                'pyflyte-execute',
                '--inputs',
                '{{.input}}',
                '--output-prefix',
                '{{.outputPrefix}}',
                '--raw-output-data-prefix',
                '{{.rawOutputDataPrefix}}',
                '--checkpoint-path',
                '{{.checkpointOutputPrefix}}',
                '--prev-checkpoint',
                '{{.prevCheckpointPrefix}}',
                '--resolver',
                'flytekit.core.python_auto_container.default_task_resolver',
                '--',
                'task-module',
                'core.control_flow.map_task',
                'task-name',
                'coalesce',
              ],
              env: [],
              config: [],
              ports: [],
              image: 'ghcr.io/flyteorg/flytekit:py3.9-latest',
              resources: {
                requests: [],
                limits: [],
              },
            },
          },
        },
        {
          template: {
            config: {},
            id: {
              resourceType: 1,
              project: 'flytesnacks',
              domain: 'development',
              name: 'core.control_flow.map_task.mapper_a_mappable_task_0',
              version: 'MT76cyUZeeYX-hA6qeIotA==',
            },
            type: 'container_array',
            metadata: {
              runtime: {
                type: 1,
                version: '0.0.0+develop',
                flavor: 'python',
              },
              retries: {},
            },
            interface: {
              inputs: {
                variables: {
                  a: {
                    type: {
                      collectionType: {
                        simple: 1,
                      },
                    },
                    description: 'a',
                  },
                },
              },
              outputs: {
                variables: {
                  o0: {
                    type: {
                      collectionType: {
                        simple: 3,
                      },
                    },
                    description: 'o0',
                  },
                },
              },
            },
            custom: {
              fields: {
                minSuccessRatio: {
                  numberValue: 1,
                },
              },
            },
            taskTypeVersion: 1,
            container: {
              command: [],
              args: [
                'pyflyte-fast-execute',
                '--additional-distribution',
                's3://flyte-demo/zn/flytesnacks/development/MT76cyUZeeYX+hA6qeIotA==/scriptmode.tar.gz',
                '--dest-dir',
                '/root',
                '--',
                'pyflyte-map-execute',
                '--inputs',
                '{{.input}}',
                '--output-prefix',
                '{{.outputPrefix}}',
                '--raw-output-data-prefix',
                '{{.rawOutputDataPrefix}}',
                '--checkpoint-path',
                '{{.checkpointOutputPrefix}}',
                '--prev-checkpoint',
                '{{.prevCheckpointPrefix}}',
                '--resolver',
                'flytekit.core.python_auto_container.default_task_resolver',
                '--',
                'task-module',
                'core.control_flow.map_task',
                'task-name',
                'a_mappable_task',
              ],
              env: [],
              config: [],
              ports: [],
              image: 'ghcr.io/flyteorg/flytekit:py3.9-latest',
              resources: {
                requests: [],
                limits: [],
              },
            },
          },
        },
      ],
      primary: {
        template: {
          nodes: [
            {
              inputs: [],
              upstreamNodeIds: [],
              outputAliases: [],
              id: 'start-node',
            },
            {
              inputs: [
                {
                  var: 'o0',
                  binding: {
                    promise: {
                      nodeId: 'n1',
                      var: 'o0',
                    },
                  },
                },
              ],
              upstreamNodeIds: [],
              outputAliases: [],
              id: 'end-node',
            },
            {
              inputs: [
                {
                  var: 'a',
                  binding: {
                    promise: {
                      nodeId: 'start-node',
                      var: 'a',
                    },
                  },
                },
              ],
              upstreamNodeIds: [],
              outputAliases: [],
              id: 'n0',
              metadata: {
                name: 'mapper_a_mappable_task_0',
                retries: {
                  retries: 1,
                },
              },
              taskNode: {
                overrides: {
                  resources: {
                    requests: [
                      {
                        name: 3,
                        value: '300Mi',
                      },
                    ],
                    limits: [
                      {
                        name: 3,
                        value: '500Mi',
                      },
                    ],
                  },
                },
                referenceId: {
                  resourceType: 1,
                  project: 'flytesnacks',
                  domain: 'development',
                  name: 'core.control_flow.map_task.mapper_a_mappable_task_0',
                  version: 'MT76cyUZeeYX-hA6qeIotA==',
                },
              },
            },
            {
              inputs: [
                {
                  var: 'b',
                  binding: {
                    promise: {
                      nodeId: 'n0',
                      var: 'o0',
                    },
                  },
                },
              ],
              upstreamNodeIds: ['n0'],
              outputAliases: [],
              id: 'n1',
              metadata: {
                name: 'coalesce',
                retries: {},
              },
              taskNode: {
                overrides: {},
                referenceId: {
                  resourceType: 1,
                  project: 'flytesnacks',
                  domain: 'development',
                  name: 'core.control_flow.map_task.coalesce',
                  version: 'MT76cyUZeeYX-hA6qeIotA==',
                },
              },
            },
          ],
          outputs: [
            {
              var: 'o0',
              binding: {
                promise: {
                  nodeId: 'n1',
                  var: 'o0',
                },
              },
            },
          ],
          id: {
            resourceType: 2,
            project: 'flytesnacks',
            domain: 'development',
            name: 'core.control_flow.map_task.my_map_workflow',
            version: 'MT76cyUZeeYX-hA6qeIotA==',
          },
          metadata: {},
          interface: {
            inputs: {
              variables: {
                a: {
                  type: {
                    collectionType: {
                      simple: 1,
                    },
                  },
                  description: 'a',
                },
              },
            },
            outputs: {
              variables: {
                o0: {
                  type: {
                    simple: 3,
                  },
                  description: 'o0',
                },
              },
            },
          },
          metadataDefaults: {},
        },
        connections: {
          downstream: {
            'start-node': {
              ids: ['n0'],
            },
            n0: {
              ids: ['n1'],
            },
            n1: {
              ids: ['end-node'],
            },
          },
          upstream: {
            'end-node': {
              ids: ['n1'],
            },
            n0: {
              ids: ['start-node'],
            },
            n1: {
              ids: ['n0'],
            },
          },
        },
      },
    },
    createdAt: {
      seconds: long(0),
      nanos: 343264000,
    },
  },
};
