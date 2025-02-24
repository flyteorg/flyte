def iterate_node_executions(
    client,
    workflow_execution_identifier=None,
    task_execution_identifier=None,
    limit=None,
    filters=None,
    unique_parent_id=None,
):
    """
    This returns a generator for node executions.
    :param flytekit.clients.friendly.SynchronousFlyteClient client:
    :param flytekit.models.core.identifier.WorkflowExecutionIdentifier workflow_execution_identifier:
    :param flytekit.models.core.identifier.TaskExecutionIdentifier task_execution_identifier:
    :param int limit: The maximum number of elements to retrieve
    :param list[flytekit.models.filters.Filter] filters:
    :rtype: Iterator[flytekit.models.node_execution.NodeExecution]
    """
    token = ""
    num_to_fetch = 100
    if limit is not None and limit < num_to_fetch:
        num_to_fetch = limit
    counter = 0
    while True:
        if workflow_execution_identifier is not None:
            node_execs, next_token = client.list_node_executions(
                workflow_execution_identifier=workflow_execution_identifier,
                limit=num_to_fetch,
                token=token,
                filters=filters,
                unique_parent_id=unique_parent_id,
            )
        else:
            node_execs, next_token = client.list_node_executions_for_task_paginated(
                task_execution_identifier=task_execution_identifier,
                limit=num_to_fetch,
                token=token,
                filters=filters,
            )
        for n in node_execs:
            counter += 1
            if limit is not None and counter > limit:
                return
            yield n
        if not next_token:
            break
        token = next_token


def iterate_task_executions(client, node_execution_identifier, limit=None, filters=None):
    """
    This returns a generator for task executions, given a node execution identifier
    :param flytekit.clients.friendly.SynchronousFlyteClient client:
    :param flytekit.models.core.identifier.NodeExecutionIdentifier node_execution_identifier:
    :param int limit: The maximum number of elements to retrieve
    :param list[flytekit.models.filters.Filter] filters:
    :rtype: Iterator[flytekit.models.admin.task_execution.TaskExecution]
    """
    token = ""
    num_to_fetch = 100
    if limit is not None and limit < num_to_fetch:
        num_to_fetch = limit
    counter = 0
    while True:
        task_execs, next_token = client.list_task_executions_paginated(
            node_execution_identifier=node_execution_identifier,
            limit=num_to_fetch,
            token=token,
            filters=filters,
        )
        for t in task_execs:
            counter += 1
            if limit is not None and counter > limit:
                return
            yield t
        if not next_token:
            break
        token = next_token
