.. _st-lp:

.. code-block:: bash

    id {
      resource_type: WORKFLOW
      project: "flytesnacks"
      domain: "development"
      name: "workflows.recipe_1.outer.StaticLaunchPlanCaller"
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
              name: "workflows.recipe_1.outer.StaticLaunchPlanCaller"
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
                    node_id: "identity-lp-execution"
                    var: "task_output"
                  }
                }
              }
            }
            nodes {
              id: "identity-lp-execution"
              metadata {
                name: "identity_lp_execution"
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
                launchplan_ref {
                  resource_type: LAUNCH_PLAN
                  project: "flytesnacks"
                  domain: "development"
                  name: "workflows.recipe_1.outer.id_lp"
                  version: "cd5dd92280a7482164f00ab6e1d0e07d20e5f4df"
                }
              }
            }
            outputs {
              var: "wf_output"
              binding {
                promise {
                  node_id: "identity-lp-execution"
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
                ids: "identity-lp-execution"
              }
            }
            downstream {
              key: "identity-lp-execution"
              value {
                ids: "start-node"
              }
            }
            upstream {
              key: "end-node"
              value {
                ids: "identity-lp-execution"
              }
            }
            upstream {
              key: "identity-lp-execution"
              value {
                ids: "start-node"
              }
            }
          }
        }
      }
    }


