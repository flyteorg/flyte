#!/bin/bash
# Creates a run directly with an inline task spec (skips deploy task).
# Run from the repo root: bash runs/test/scripts/create_run.sh

ENDPOINT="${ENDPOINT:-http://localhost:8090}"
<<<<<<< HEAD
ORG="${ORG:-testorg}"
=======
ORG="${ORG:-dogfood-gcp}"
>>>>>>> enghabu/state-etcd
PROJECT="${PROJECT:-flytesnacks}"
DOMAIN="${DOMAIN:-development}"

buf curl --schema . "$ENDPOINT/flyteidl2.workflow.RunService/CreateRun" --data @- <<'EOF'
{
    "projectId": {
<<<<<<< HEAD
        "organization": "testorg",
=======
        "organization": "dogfood-gcp",
>>>>>>> enghabu/state-etcd
        "name": "flytesnacks",
        "domain": "development"
    },
    "taskSpec": {
        "taskTemplate": {
            "id": {
                "resourceType": "TASK",
                "project": "flytesnacks",
                "domain": "development",
                "name": "hello_world.say_hello",
                "version": "ae90079690e275d2dd68fadbf4a66641",
<<<<<<< HEAD
                "org": "testorg"
=======
                "org": "dogfood-gcp"
>>>>>>> enghabu/state-etcd
            },
            "type": "python",
            "metadata": {
                "discoverable": false,
                "runtime": {
                    "type": "FLYTE_SDK",
                    "version": "2.0.0b54.dev3+g27699481e",
                    "flavor": "python"
                },
                "timeout": "0s",
                "retries": {
                    "retries": 0
                },
                "interruptible": false,
                "cacheSerializable": false,
                "tags": {},
                "debuggable": true,
                "generatesDeck": false,
                "cacheIgnoreInputVars": [],
                "isEager": false
            },
            "interface": {
                "inputs": {
<<<<<<< HEAD
                    "variables": [
                        {"key": "x_list", "value": {"type": {"collectionType": {"simple": "FLOAT"}}}}
                    ]
                },
                "outputs": {
                    "variables": [
                        {"key": "o0", "value": {"type": {"simple": "STRING"}}}
                    ]
                }
            },
            "container": {
                "image": "ghcr.io/flyteorg/flyte:py3.10-v2.0.0b55",
                "args": [
                  "a0",
                  "--inputs",
                  "{{.input}}",
                  "--outputs-path",
                  "{{.outputPrefix}}",
                  "--version",
                  "45b230d2b24b90cb7578333647e619c4",
                  "--raw-data-path",
                  "{{.rawOutputDataPrefix}}",
                  "--checkpoint-path",
                  "{{.checkpointOutputPrefix}}",
                  "--prev-checkpoint",
                  "{{.prevCheckpointPrefix}}",
                  "--run-name",
                  "{{.runName}}",
                  "--name",
                  "{{.actionName}}",
                  "--image-cache",
                  "H4sIAPdEimkC/6tWysxNTE+Nz8nPzy4tULKqVspIzcnJjy/PL8pJUbJSSs9ILtLLzNdPy6ksSc0vSocwrAoqjfUMDXTLjPQM9AySTE2VamsBGQCEuEoAAAA=",
                  "--tgz",
                  "s3://flyte-data/fast-registration/fast.tar.gz",
                  "--dest",
                  ".",
                  "--resolver",
                  "flyte._internal.resolvers.default.DefaultTaskResolver",
                  "mod",
                  "hello",
                  "instance",
                  "main"
=======
                    "variables": {
                        "data": {
                            "type": {
                                "simple": "STRING"
                            }
                        },
                        "lt": {
                            "type": {
                                "collectionType": {
                                    "simple": "INTEGER"
                                }
                            }
                        }
                    }
                },
                "outputs": {
                    "variables": {
                        "o0": {
                            "type": {
                                "simple": "STRING"
                            }
                        }
                    }
                }
            },
            "container": {
                "image": "us-docker.pkg.dev/dogfood-gcp-dataplane/orgs/dogfood-gcp/flyte:38b0fdf19f1b10b83e74eac5f5d241cb",
                "args": [
                    "a0",
                    "--inputs",
                    "{{.input}}",
                    "--outputs-path",
                    "{{.outputPrefix}}",
                    "--version",
                    "ae90079690e275d2dd68fadbf4a66641",
                    "--raw-data-path",
                    "{{.rawOutputDataPrefix}}",
                    "--checkpoint-path",
                    "{{.checkpointOutputPrefix}}",
                    "--prev-checkpoint",
                    "{{.prevCheckpointPrefix}}",
                    "--run-name",
                    "{{.runName}}",
                    "--name",
                    "{{.actionName}}",
                    "--image-cache",
                    "H4sIAAAAAAAC/03JUQ6CMAwA0Lv0f4wJRNxlyLa206zaZYDGEO7ur+/3HfB4hkyLqJa9gj/gTiK6fLQJgod9NaipUOtqyR3S26JmVkWTUzUYtlAlvMhqy+t/WZbvRn6YY8/I7sYuuj7OA11HCmniCS+jSxHO8weQe/MhggAAAA==",
                    "--tgz",
                    "gs://opta-gcp-dogfood-gcp-dp-dogfood-gcp-fast-registration/flytesnacks/development/V2IAPFUQ4J25FXLI7LN7JJTGIE======/fast1a8cc5bb1af09a1108b3ca1189b1c342.tar.gz",
                    "--dest",
                    ".",
                    "--resolver",
                    "flyte._internal.resolvers.default.DefaultTaskResolver",
                    "mod",
                    "examples.basics.devbox_one",
                    "instance",
                    "say_hello"
>>>>>>> enghabu/state-etcd
                ],
                "resources": {
                    "requests": [
                        {"name": "CPU", "value": "1"},
                        {"name": "MEMORY", "value": "1Gi"}
                    ]
                }
            },
            "taskTypeVersion": 0,
            "config": {}
        }
    },
    "inputs": {
        "literals": [
            {
<<<<<<< HEAD
                "name": "x_list",
=======
                "name": "data",
                "value": {
                    "scalar": {
                        "primitive": {
                            "stringValue": "hello world"
                        }
                    }
                }
            },
            {
                "name": "lt",
>>>>>>> enghabu/state-etcd
                "value": {
                    "collection": {
                        "literals": [
                            {"scalar": {"primitive": {"integer": "1"}}},
                            {"scalar": {"primitive": {"integer": "2"}}},
<<<<<<< HEAD
                            {"scalar": {"primitive": {"integer": "1"}}},
                            {"scalar": {"primitive": {"integer": "2"}}},
                            {"scalar": {"primitive": {"integer": "1"}}},
                            {"scalar": {"primitive": {"integer": "2"}}},
                            {"scalar": {"primitive": {"integer": "1"}}},
                            {"scalar": {"primitive": {"integer": "2"}}},
                            {"scalar": {"primitive": {"integer": "1"}}},
                            {"scalar": {"primitive": {"integer": "2"}}},
                            {"scalar": {"primitive": {"integer": "1"}}},
                            {"scalar": {"primitive": {"integer": "2"}}},
                            {"scalar": {"primitive": {"integer": "1"}}},
                            {"scalar": {"primitive": {"integer": "2"}}},
                            {"scalar": {"primitive": {"integer": "1"}}},
                            {"scalar": {"primitive": {"integer": "2"}}},
=======
>>>>>>> enghabu/state-etcd
                            {"scalar": {"primitive": {"integer": "3"}}}
                        ]
                    }
                }
            }
        ]
    }
}
EOF
