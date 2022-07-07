from flytekit.configuration import Config
from flytekit.remote import FlyteRemote

from basic_workflow import my_wf

remote = FlyteRemote(
    config=Config.auto(),
    default_project="flytesnacks",
    default_domain="development",
)

registered_workflow = remote.register_script(my_wf)

execution = remote.execute(registered_workflow, inputs={"a": 100, "b": "hello"})
print(f"Execution successfully started: {execution.id.name}")
