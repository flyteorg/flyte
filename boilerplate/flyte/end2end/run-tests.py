#!/usr/bin/env python3
import json
import sys
import time
import traceback
from typing import Dict, List, Optional

import click
import requests
from flytekit.configuration import Config
from flytekit.models.core.execution import WorkflowExecutionPhase
from flytekit.remote import FlyteRemote
from flytekit.remote.executions import FlyteWorkflowExecution

WAIT_TIME = 10
MAX_ATTEMPTS = 200


def execute_workflow(
    remote: FlyteRemote,
    version,
    workflow_name,
    inputs,
    cluster_pool_name: Optional[str] = None,
):
    print(f"Fetching workflow={workflow_name} and version={version}")
    wf = remote.fetch_workflow(name=workflow_name, version=version)
    return remote.execute(wf, inputs=inputs, wait=False, cluster_pool=cluster_pool_name)


def executions_finished(
    executions_by_wfgroup: Dict[str, List[FlyteWorkflowExecution]]
) -> bool:
    for executions in executions_by_wfgroup.values():
        if not all([execution.is_done for execution in executions]):
            return False
    return True


def sync_executions(
    remote: FlyteRemote, executions_by_wfgroup: Dict[str, List[FlyteWorkflowExecution]]
):
    try:
        for executions in executions_by_wfgroup.values():
            for execution in executions:
                print(f"About to sync execution_id={execution.id.name}")
                remote.sync(execution)
    except Exception:
        print(traceback.format_exc())
        print("GOT TO THE EXCEPT")
        print("COUNT THIS!")


def report_executions(executions_by_wfgroup: Dict[str, List[FlyteWorkflowExecution]]):
    for executions in executions_by_wfgroup.values():
        for execution in executions:
            print(execution)


def schedule_workflow_groups(
    tag: str,
    workflow_groups: List[str],
    remote: FlyteRemote,
    terminate_workflow_on_failure: bool,
    parsed_manifest: List[dict],
    cluster_pool_name: Optional[str] = None,
) -> Dict[str, bool]:
    """
    Schedule workflows executions for all workflow groups and return True if all executions succeed, otherwise
    return False.
    """
    executions_by_wfgroup = {}
    # Schedule executions for each workflow group,
    for wf_group in workflow_groups:
        workflow_group_item = list(
            filter(lambda item: item["name"] == wf_group, parsed_manifest)
        )
        if not workflow_group_item:
            continue
        workflows = workflow_group_item[0].get("examples")
        if not workflows:
            continue
        executions_by_wfgroup[wf_group] = [
            execute_workflow(remote, tag, workflow[0], workflow[1], cluster_pool_name)
            for workflow in workflows
        ]

    # Wait for all executions to finish
    attempt = 0
    while attempt == 0 or (
        not executions_finished(executions_by_wfgroup) and attempt < MAX_ATTEMPTS
    ):
        attempt += 1
        print(
            f"Not all executions finished yet. Sleeping for some time, will check again in {WAIT_TIME}s"
        )
        time.sleep(WAIT_TIME)
        sync_executions(remote, executions_by_wfgroup)

    report_executions(executions_by_wfgroup)

    results = {}
    for wf_group, executions in executions_by_wfgroup.items():
        non_succeeded_executions = []
        for execution in executions:
            if execution.closure.phase != WorkflowExecutionPhase.SUCCEEDED:
                non_succeeded_executions.append(execution)
        # Report failing cases
        if len(non_succeeded_executions) != 0:
            print(f"Failed executions for {wf_group}:")
            for execution in non_succeeded_executions:
                print(
                    f"    workflow={execution.spec.launch_plan.name}, execution_id={execution.id.name}"
                )
                if terminate_workflow_on_failure:
                    remote.terminate(
                        execution, "aborting execution scheduled in functional test"
                    )
        # A workflow group succeeds iff all of its executions succeed
        results[wf_group] = len(non_succeeded_executions) == 0
    return results


def valid(workflow_group, parsed_manifest):
    """
    Return True if a workflow group is contained in parsed_manifest,
    False otherwise.
    """
    return workflow_group in set(wf_group["name"] for wf_group in parsed_manifest)


def run(
    flytesnacks_release_tag: str,
    priorities: List[str],
    config_file_path,
    terminate_workflow_on_failure: bool,
    test_project_name: str,
    test_project_domain: str,
    cluster_pool_name: Optional[str] = None,
) -> List[Dict[str, str]]:
    remote = FlyteRemote(
        Config.auto(config_file=config_file_path),
        test_project_name,
        test_project_domain,
    )

    # For a given release tag and priority, this function filters the workflow groups from the flytesnacks
    # manifest file. For example, for the release tag "v0.2.224" and the priority "P0" it returns [ "core" ].
    manifest_url = (
        "https://raw.githubusercontent.com/flyteorg/flytesnacks/"
        f"{flytesnacks_release_tag}/flyte_tests_manifest.json"
    )
    r = requests.get(manifest_url)
    parsed_manifest = r.json()
    workflow_groups = []
    workflow_groups = (
        ["lite"]
        if "lite" in priorities
        else [
            group["name"]
            for group in parsed_manifest
            if group["priority"] in priorities
        ]
    )

    results = []
    valid_workgroups = []
    for workflow_group in workflow_groups:
        if not valid(workflow_group, parsed_manifest):
            results.append(
                {
                    "label": workflow_group,
                    "status": "coming soon",
                    "color": "grey",
                }
            )
            continue
        valid_workgroups.append(workflow_group)

    results_by_wfgroup = schedule_workflow_groups(
        flytesnacks_release_tag,
        valid_workgroups,
        remote,
        terminate_workflow_on_failure,
        parsed_manifest,
        cluster_pool_name,
    )

    for workflow_group, succeeded in results_by_wfgroup.items():
        if succeeded:
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


@click.command()
@click.argument("flytesnacks_release_tag")
@click.argument("priorities")
@click.argument("config_file")
@click.option(
    "--return_non_zero_on_failure",
    default=False,
    is_flag=True,
    help="Return a non-zero exit status if any workflow fails",
)
@click.option(
    "--terminate_workflow_on_failure",
    default=False,
    is_flag=True,
    help="Abort failing workflows upon exit",
)
@click.option(
    "--test_project_name",
    default="flytesnacks",
    type=str,
    is_flag=False,
    help="Name of project to run functional tests on",
)
@click.option(
    "--test_project_domain",
    default="development",
    type=str,
    is_flag=False,
    help="Name of domain in project to run functional tests on",
)
@click.argument(
    "cluster_pool_name",
    required=False,
    type=str,
    default=None,
)
def cli(
    flytesnacks_release_tag,
    priorities,
    config_file,
    return_non_zero_on_failure,
    terminate_workflow_on_failure,
    test_project_name,
    test_project_domain,
    cluster_pool_name,
):
    print(f"return_non_zero_on_failure={return_non_zero_on_failure}")
    results = run(
        flytesnacks_release_tag,
        priorities,
        config_file,
        terminate_workflow_on_failure,
        test_project_name,
        test_project_domain,
        cluster_pool_name,
    )

    # Write a json object in its own line describing the result of this run to stdout
    print(f"Result of run:\n{json.dumps(results)}")

    # Return a non-zero exit code if core fails
    if return_non_zero_on_failure:
        for result in results:
            if result["status"] not in ("passing", "coming soon"):
                sys.exit(1)


if __name__ == "__main__":
    cli()
