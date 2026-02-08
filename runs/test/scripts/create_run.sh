#!/bin/bash
# Creates a run directly with an inline task spec (skips deploy task).
# Run from the repo root: bash runs/test/scripts/create_run.sh

ENDPOINT="${ENDPOINT:-http://localhost:8090}"
ORG="${ORG:-dogfood-gcp}"
PROJECT="${PROJECT:-flytesnacks}"
DOMAIN="${DOMAIN:-development}"

buf curl --schema . "$ENDPOINT/flyteidl2.workflow.RunService/CreateRun" --data @- <<'EOF'
{
    "projectId": {
        "organization": "dogfood-gcp",
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
                "org": "dogfood-gcp"
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
                "value": {
                    "collection": {
                        "literals": [
                            {"scalar": {"primitive": {"integer": "1"}}},
                            {"scalar": {"primitive": {"integer": "2"}}},
                            {"scalar": {"primitive": {"integer": "3"}}}
                        ]
                    }
                }
            }
        ]
    }
}
EOF
