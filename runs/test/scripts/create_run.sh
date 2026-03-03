#!/bin/bash
# Creates a run directly with an inline task spec (skips deploy task).
# Run from the repo root: bash runs/test/scripts/create_run.sh

ENDPOINT="${ENDPOINT:-http://localhost:8090}"
ORG="${ORG:-testorg}"
PROJECT="${PROJECT:-flytesnacks}"
DOMAIN="${DOMAIN:-development}"

buf curl --schema . "$ENDPOINT/flyteidl2.workflow.RunService/CreateRun" --data @- <<'EOF'
{
    "projectId": {
        "organization": "testorg",
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
                "org": "testorg"
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
                "name": "x_list",
                "value": {
                    "collection": {
                        "literals": [
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
