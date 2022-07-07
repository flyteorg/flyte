from flytekit.configuration import Config
from flytekit.remote import FlyteRemote

from imperative_wf_style import wf

remote = FlyteRemote(
    config=Config.auto(),
    default_project="flytesnacks",
    default_domain="development",
)

registered_workflow = remote.register_script(wf)

execution = remote.execute(registered_workflow, inputs={"in1": "hello", "in2": "foo"})
print(f"Execution successfully started: {execution.id.name}")
