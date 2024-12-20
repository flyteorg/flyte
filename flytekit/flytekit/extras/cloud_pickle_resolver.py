from base64 import b64decode, b64encode
from typing import List

import cloudpickle

from flytekit.configuration import SerializationSettings
from flytekit.core.base_task import TaskResolverMixin
from flytekit.core.python_auto_container import PythonAutoContainerTask
from flytekit.core.tracker import TrackedInstance


class ExperimentalNaiveCloudPickleResolver(TrackedInstance, TaskResolverMixin):
    """
    Please do not use this resolver, basically ever. This is here for demonstration purposes only. The critical flaw
    of this resolver is that pretty much any task that it resolves results in loader_args that are enormous. This
    payload is serialized as part of the ``TaskTemplate`` protobuf object and will live in Admin and then be loaded
    into Flyte Propeller memory and will pretty much clog up performance along the entire platform.

    TODO: Replace this with a version that will upload the data to S3 or some other durable store upon ``loader_args``
      and will download the data upon ``load_task``. This will require additional changes to Admin however.
    """

    def name(self) -> str:
        return "cloud pickling task resolver"

    def load_task(self, loader_args: List[str]) -> PythonAutoContainerTask:
        raw_bytes = loader_args[0].encode("ascii")
        pickled = b64decode(raw_bytes)
        return cloudpickle.loads(pickled)

    def loader_args(self, settings: SerializationSettings, t: PythonAutoContainerTask) -> List[str]:
        return [b64encode(cloudpickle.dumps(t)).decode("ascii")]

    def get_all_tasks(self) -> List[PythonAutoContainerTask]:
        pass


experimental_cloud_pickle_resolver = ExperimentalNaiveCloudPickleResolver()
