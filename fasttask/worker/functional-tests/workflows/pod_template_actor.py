import os
from flytekit import PodTemplate
from kubernetes.client.models import (
    V1PodSpec,
    V1Container,
)
from union.actor import ActorEnvironment

pod_template = PodTemplate(
    primary_container_name="primary",
    pod_spec=V1PodSpec(
        containers=[
            V1Container(
                name="primary",
                image=os.getenv("UNION_RUNTIME_TEST_IMAGE"),
                termination_message_policy="FallbackToLogsOnError",
            ),
        ],
    ),
)
actor_env = ActorEnvironment(
    name="template",
    pod_template=pod_template,
    ttl_seconds=10,
)


@actor_env.task
def wf() -> str:
    return "hello"
