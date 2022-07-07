from flytekit.configuration import Config
from flytekit.remote import FlyteRemote

from task import wf

remote = FlyteRemote(
    config=Config.auto(),
    default_project="flytesnacks",
    default_domain="development",
)

registered_workflow = remote.register_script(wf)

execution = remote.execute(registered_workflow, inputs={"n": 2})
print(f"Execution successfully started: {execution.id.name}")
