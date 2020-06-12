.. _st-swf:

.. code-block:: bash

    id {
      resource_type: WORKFLOW
      project: "flytesnacks"
      domain: "development"
      name: "workflows.recipe_1.outer.StaticSubWorkflowCaller"
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
              name: "workflows.recipe_1.outer.StaticSubWorkflowCaller"
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
                    node_id: "identity-wf-execution"
                    var: "task_output"
                  }
                }
              }
            }
            nodes {
              id: "identity-wf-execution"
              metadata {
                name: "identity_wf_execution"
                timeout {
                }
                retries {
                }
                interruptible: false
              }
              inputs {
                var: "a"
                binding {
                  promise {
                    node_id: "start-node"
                    var: "outer_a"
                  }
                }
              }
              workflow_node {
                sub_workflow_ref {
                  resource_type: WORKFLOW
                  project: "flytesnacks"
                  domain: "development"
                  name: "workflows.recipe_1.inner.IdentityWorkflow"
                  version: "cd5dd92280a7482164f00ab6e1d0e07d20e5f4df"
                }
              }
            }
            outputs {
              var: "wf_output"
              binding {
                promise {
                  node_id: "identity-wf-execution"
                  var: "task_output"
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
                ids: "identity-wf-execution"
              }
            }
            downstream {
              key: "identity-wf-execution"
              value {
                ids: "start-node"
              }
            }
            upstream {
              key: "end-node"
              value {
                ids: "identity-wf-execution"
              }
            }
            upstream {
              key: "identity-wf-execution"
              value {
                ids: "start-node"
              }
            }
          }
        }
        sub_workflows {
          template {
            id {
              resource_type: WORKFLOW
              project: "flytesnacks"
              domain: "development"
              name: "workflows.recipe_1.inner.IdentityWorkflow"
              version: "cd5dd92280a7482164f00ab6e1d0e07d20e5f4df"
            }
            metadata {
            }
            interface {
              inputs {
                variables {
                  key: "a"
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
                  key: "task_output"
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
                var: "task_output"
                binding {
                  promise {
                    node_id: "odd-nums-task"
                    var: "out"
                  }
                }
              }
            }
            nodes {
              id: "odd-nums-task"
              metadata {
                name: "odd_nums_task"
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
                    var: "a"
                  }
                }
              }
              task_node {
                reference_id {
                  resource_type: TASK
                  project: "flytesnacks"
                  domain: "development"
                  name: "workflows.recipe_1.inner.inner_task"
                  version: "cd5dd92280a7482164f00ab6e1d0e07d20e5f4df"
                }
              }
            }
            outputs {
              var: "task_output"
              binding {
                promise {
                  node_id: "odd-nums-task"
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
                ids: "odd-nums-task"
              }
            }
            downstream {
              key: "odd-nums-task"
              value {
                ids: "start-node"
              }
            }
            upstream {
              key: "end-node"
              value {
                ids: "odd-nums-task"
              }
            }
            upstream {
              key: "odd-nums-task"
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
              name: "workflows.recipe_1.inner.inner_task"
              version: "cd5dd92280a7482164f00ab6e1d0e07d20e5f4df"
            }
            type: "python-task"
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
              args: "workflows.recipe_1.inner"
              args: "--task-name"
              args: "inner_task"
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


