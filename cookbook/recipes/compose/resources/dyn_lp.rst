.. _dyn-lp:

.. code-block:: bash

    id {
      resource_type: WORKFLOW
      project: "flytesnacks"
      domain: "development"
      name: "workflows.recipe_1.outer.DynamicLaunchPlanCaller"
      version: "cd5dd92280a7482164f00ab6e1d0e07d20e5f4df"
    }
    closure {
      compiled_workflow {
        primary {
          template {
            id {
              resource_type: WORKFLOW
              project: "flytesnacks"
              domain: "development"
              name: "workflows.recipe_1.outer.DynamicLaunchPlanCaller"
              version: "cd5dd92280a7482164f00ab6e1d0e07d20e5f4df"
            }
            metadata {
            }
            interface {
              inputs {
                variables {
                  key: "outer_a"
                  value {
                    type {
                      simple: INTEGER
                    }
                    description: "Input for inner workflow"
                  }
                }
              }
              outputs {
                variables {
                  key: "wf_output"
                  value {
                    type {
                      simple: INTEGER
                    }
                  }
                }
              }
            }
            nodes {
              id: "start-node"
              metadata {
                timeout {
                }
                retries {
                }
                interruptible: false
              }
            }
            nodes {
              id: "end-node"
              metadata {
                timeout {
                }
                retries {
                }
                interruptible: false
              }
              inputs {
                var: "wf_output"
                binding {
                  promise {
                    node_id: "lp-task"
                    var: "out"
                  }
                }
              }
            }
            nodes {
              id: "lp-task"
              metadata {
                name: "lp_task"
                timeout {
                }
                retries {
                }
                interruptible: false
              }
              inputs {
                var: "num"
                binding {
                  promise {
                    node_id: "start-node"
                    var: "outer_a"
                  }
                }
              }
              task_node {
                reference_id {
                  resource_type: TASK
                  project: "flytesnacks"
                  domain: "development"
                  name: "workflows.recipe_1.outer.lp_yield_task"
                  version: "cd5dd92280a7482164f00ab6e1d0e07d20e5f4df"
                }
              }
            }
            outputs {
              var: "wf_output"
              binding {
                promise {
                  node_id: "lp-task"
                  var: "out"
                }
              }
            }
            metadata_defaults {
            }
          }
          connections {
            downstream {
              key: "end-node"
              value {
                ids: "lp-task"
              }
            }
            downstream {
              key: "lp-task"
              value {
                ids: "start-node"
              }
            }
            upstream {
              key: "end-node"
              value {
                ids: "lp-task"
              }
            }
            upstream {
              key: "lp-task"
              value {
                ids: "start-node"
              }
            }
          }
        }
        tasks {
          template {
            id {
              resource_type: TASK
              project: "flytesnacks"
              domain: "development"
              name: "workflows.recipe_1.outer.lp_yield_task"
              version: "cd5dd92280a7482164f00ab6e1d0e07d20e5f4df"
            }
            type: "dynamic-task"
            metadata {
              runtime {
                type: FLYTE_SDK
                version: "0.7.1b1"
                flavor: "python"
              }
              timeout {
              }
              retries {
              }
            }
            interface {
              inputs {
                variables {
                  key: "num"
                  value {
                    type {
                      simple: INTEGER
                    }
                  }
                }
              }
              outputs {
                variables {
                  key: "imported_output"
                  value {
                    type {
                      simple: INTEGER
                    }
                  }
                }
                variables {
                  key: "out"
                  value {
                    type {
                      simple: INTEGER
                    }
                  }
                }
              }
            }
            container {
              image: "docker.io/lyft/flytecookbook:cd5dd92280a7482164f00ab6e1d0e07d20e5f4df"
              args: "flytekit_venv"
              args: "pyflyte-execute"
              args: "--task-module"
              args: "workflows.recipe_1.outer"
              args: "--task-name"
              args: "lp_yield_task"
              args: "--inputs"
              args: "{{.input}}"
              args: "--output-prefix"
              args: "{{.outputPrefix}}"
              resources {
              }
              env {
                key: "FLYTE_INTERNAL_CONFIGURATION_PATH"
                value: "/root/sandbox.config"
              }
              env {
                key: "FLYTE_INTERNAL_IMAGE"
                value: "docker.io/lyft/flytecookbook:cd5dd92280a7482164f00ab6e1d0e07d20e5f4df"
              }
              env {
                key: "FLYTE_INTERNAL_PROJECT"
                value: "flytesnacks"
              }
              env {
                key: "FLYTE_INTERNAL_DOMAIN"
                value: "development"
              }
              env {
                key: "FLYTE_INTERNAL_NAME"
              }
              env {
                key: "FLYTE_INTERNAL_VERSION"
                value: "cd5dd92280a7482164f00ab6e1d0e07d20e5f4df"
              }
            }
          }
        }
      }
    }


