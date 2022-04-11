
#!/usr/bin/env python3

import json
import sys
import time
import traceback
import requests
from pathlib import Path
from typing import List, Mapping, Tuple, Dict
from flytekit.remote import FlyteRemote
from flytekit.models.core.execution import WorkflowExecutionPhase

WAIT_TIME = 10
MAX_ATTEMPTS = 60

# This dictionary maps the names found in the flytesnacks manifest to a list of workflow names and
# inputs. This is so we can progressively cover all priorities in the original flytesnacks manifest,
# starting with "core".
FLYTESNACKS_WORKFLOW_GROUPS: Mapping[str, List[Tuple[str, dict]]] = {
    "core": [
        ("core.control_flow.chain_tasks.chain_tasks_wf", {}),
        ("core.control_flow.dynamics.wf", {"s1": "Pear", "s2": "Earth"}),
        ("core.control_flow.map_task.my_map_workflow", {"a": [1, 2, 3, 4, 5]}),
        # Workflows that use nested executions cannot be launched via flyteremote.
        # This issue is being tracked in https://github.com/flyteorg/flyte/issues/1482.
        # ("core.control_flow.run_conditions.multiplier", {"my_input": 0.5}),
        # ("core.control_flow.run_conditions.multiplier_2", {"my_input": 10}),
        # ("core.control_flow.run_conditions.multiplier_3", {"my_input": 5}),
        # ("core.control_flow.run_conditions.basic_boolean_wf", {"seed": 5}),
        # ("core.control_flow.run_conditions.bool_input_wf", {"b": True}),
        # ("core.control_flow.run_conditions.nested_conditions", {"my_input": 0.4}),
        # ("core.control_flow.run_conditions.consume_outputs", {"my_input": 0.4, "seed": 7}),
        # ("core.control_flow.run_merge_sort.merge_sort", {"numbers": [5, 4, 3, 2, 1], "count": 5}),
        ("core.control_flow.subworkflows.parent_wf", {"a": 3}),
        ("core.control_flow.subworkflows.nested_parent_wf", {"a": 3}),
        ("core.flyte_basics.basic_workflow.my_wf", {"a": 50, "b": "hello"}),
        # Getting a 403 for the wikipedia image
        # ("core.flyte_basics.files.rotate_one_workflow", {"in_image": "https://upload.wikimedia.org/wikipedia/commons/d/d2/Julia_set_%28C_%3D_0.285%2C_0.01%29.jpg"}),
        ("core.flyte_basics.folders.download_and_rotate", {}),
        ("core.flyte_basics.hello_world.my_wf", {}),
        ("core.flyte_basics.lp.my_wf", {"val": 4}),
        ("core.flyte_basics.lp.go_greet", {"day_of_week": "5", "number": 3, "am": True}),
        ("core.flyte_basics.named_outputs.my_wf", {}),
        # # Getting a 403 for the wikipedia image
        # # ("core.flyte_basics.reference_task.wf", {}),
        ("core.type_system.custom_objects.wf", {"x": 10, "y": 20}),
        # Enums are not supported in flyteremote
        # ("core.type_system.enums.enum_wf", {"c": "red"}),
        ("core.type_system.schema.df_wf", {"a": 42}),
        ("core.type_system.typed_schema.wf", {}),
        ("my.imperative.workflow.example", {"in1": "hello", "in2": "foo"}),
    ],
}


def run_launch_plan(remote, version, workflow_name, inputs):
    print(f"Fetching workflow={workflow_name} and version={version}")
    lp = remote.fetch_workflow(name=workflow_name, version=version)
    return remote.execute(lp, inputs=inputs, wait=False)


def schedule_workflow_group(tag: str, workflow_group: str, remote: FlyteRemote) -> bool:
    """
    Schedule all workflows executions and return True if all executions succeed, otherwise
    return False.
    """
    workflows = FLYTESNACKS_WORKFLOW_GROUPS.get(workflow_group, [])

    launch_plans = [
        run_launch_plan(remote, tag, workflow[0], workflow[1]) for workflow in workflows
    ]

    # Wait for all launch plans to finish
    attempt = 0
    while attempt == 0 or (
        not all([lp.is_complete for lp in launch_plans]) and attempt < MAX_ATTEMPTS
    ):
        attempt += 1
        print(
            f"Not all executions finished yet. Sleeping for some time, will check again in {WAIT_TIME}s"
        )
        time.sleep(WAIT_TIME)
        # Need to sync to refresh status of executions
        for lp in launch_plans:
            print(f"About to sync execution_id={lp.id.name}")
            remote.sync(lp)

    # Report result of each launch plan
    for lp in launch_plans:
        print(lp)

    # Collect all failing launch plans
    non_succeeded_lps = [
        lp
        for lp in launch_plans
        if lp.closure.phase != WorkflowExecutionPhase.SUCCEEDED
    ]

    if len(non_succeeded_lps) == 0:
        print("All executions succeeded.")
        return True

    print("Failed executions:")
    # Report failing cases
    for lp in non_succeeded_lps:
        print(f"    workflow={lp.spec.launch_plan.name}, execution_id={lp.id.name}")
    return False


def valid(workflow_group):
    """
    Return True if a workflow group is contained in FLYTESNACKS_WORKFLOW_GROUPS,
    False otherwise.
    """
    return workflow_group in FLYTESNACKS_WORKFLOW_GROUPS.keys()


def run(release_tag: str, priorities: List[str], config_file_path) -> List[Dict[str, str]]:
    remote = FlyteRemote.from_config(
        default_project="flytesnacks",
        default_domain="development",
        config_file_path=config_file_path,
    )

    # For a given release tag and priority, this function filters the workflow groups from the flytesnacks manifest file. For
    # example, for the release tag "v0.2.224" and the priority "P0" it returns [ "core" ].
    manifest_url = f"https://raw.githubusercontent.com/flyteorg/flytesnacks/{release_tag}/cookbook/flyte_tests_manifest.json"
    r = requests.get(manifest_url)
    parsed_manifest = r.json()

    workflow_groups = [group["name"] for group in parsed_manifest if group["priority"] in priorities]
    results = []
    for workflow_group in workflow_groups:
        if not valid(workflow_group):
            results.append({
                "label": workflow_group,
                "status": "coming soon",
                "color": "grey",
            })
            continue

        try:
            workflows_succeeded = schedule_workflow_group(flytesnacks_release_tag, workflow_group, remote)
        except Exception:
            print(traceback.format_exc())

            workflows_succeeded = False

        if workflows_succeeded:
            background_color = "green"
            status = "passing"
        else:
            background_color = "red"
            status = "failing"

        # Workflow groups can be only in one of three states:
        #   1. passing: this indicates all the workflow executions for that workflow group
        #               executed successfully
        #   2. failing: this state indicates that at least one execution failed in that
        #               workflow group
        #   3. coming soon: this state is used to indicate that the workflow group was not
        #                   implemented yet.
        #
        # Each state has a corresponding status and color to be used in the badge for that
        # workflow group.
        result = {
            "label": workflow_group,
            "status": status,
            "color": background_color,
        }
        results.append(result)
    return results


if __name__ == "__main__":
    # Assume that the first argument passed to the script is a flytesnacks release tag and
    # the second one is a comma-separated list of priorities, as defined in the flytesnacks
    # tests manifest.
    flytesnacks_release_tag = sys.argv[1]
    priorities = sys.argv[2].split(',')
    config_file = sys.argv[3]
    return_non_zero_on_failure = sys.argv[4]

    results = run(flytesnacks_release_tag, priorities, config_file)

    # Write a json object in its own line describing the result of this run to stdout
    print(f"Result of run:\n{json.dumps(results)}")

    # Return a non-zero exit code if core fails
    if return_non_zero_on_failure is not None:
        # find the result
        for result in results:
            if result['label'] == return_non_zero_on_failure and result['status'] != 'passing':
                sys.exit(1)
