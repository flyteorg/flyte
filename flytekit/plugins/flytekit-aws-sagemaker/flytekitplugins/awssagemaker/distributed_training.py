from __future__ import annotations

import json
import os
import typing
from dataclasses import dataclass

import retry

SM_RESOURCE_CONFIG_FILE = "/opt/ml/input/config/resourceconfig.json"
SM_ENV_VAR_CURRENT_HOST = "SM_CURRENT_HOST"
SM_ENV_VAR_HOSTS = "SM_HOSTS"
SM_ENV_VAR_NETWORK_INTERFACE_NAME = "SM_NETWORK_INTERFACE_NAME"


def setup_envars_for_testing():
    """
    This method is useful in simulating the env variables that sagemaker will set on the execution environment
    """
    os.environ[SM_ENV_VAR_CURRENT_HOST] = "host"
    os.environ[SM_ENV_VAR_HOSTS] = '["host1","host2"]'
    os.environ[SM_ENV_VAR_NETWORK_INTERFACE_NAME] = "nw"


@dataclass
class DistributedTrainingContext(object):
    current_host: str
    hosts: typing.List[str]
    network_interface_name: str

    @classmethod
    @retry.retry(exceptions=KeyError, delay=1, tries=10, backoff=1)
    def from_env(cls) -> DistributedTrainingContext:
        """
        SageMaker suggests "Hostname information might not be immediately available to the processing container.
        We recommend adding a retry policy on hostname resolution operations as nodes become available in the cluster."
        https://docs.aws.amazon.com/sagemaker/latest/dg/build-your-own-processing-container.html#byoc-config
        This is why we have an automatic retry policy
        """
        curr_host = os.environ.get(SM_ENV_VAR_CURRENT_HOST)
        raw_hosts = os.environ.get(SM_ENV_VAR_HOSTS)
        nw_iface = os.environ.get(SM_ENV_VAR_NETWORK_INTERFACE_NAME)
        if not (curr_host and raw_hosts and nw_iface):
            raise KeyError("Unable to locate Sagemaker Environment variables!")
        hosts = json.loads(raw_hosts)
        return DistributedTrainingContext(curr_host, hosts, nw_iface)

    @classmethod
    @retry.retry(exceptions=FileNotFoundError, delay=1, tries=10, backoff=1)
    def from_sagemaker_context_file(cls) -> DistributedTrainingContext:
        with open(SM_RESOURCE_CONFIG_FILE, "r") as rc_file:
            d = json.load(rc_file)
            curr_host = d["current_host"]
            hosts = d["hosts"]
            nw_iface = d["network_interface_name"]

            if not (curr_host and hosts and nw_iface):
                raise KeyError

            return DistributedTrainingContext(curr_host, hosts, nw_iface)

    @classmethod
    def local_execute(cls) -> DistributedTrainingContext:
        """
        Creates a dummy local execution context for distributed execution.
        TODO revisit if this is a good idea
        """
        return DistributedTrainingContext(hosts=["localhost"], current_host="localhost", network_interface_name="dummy")


DISTRIBUTED_TRAINING_CONTEXT_KEY = "DISTRIBUTED_TRAINING_CONTEXT"
"""
Use this key to retrieve the distributed training context of type :py:class:`DistributedTrainingContext`.
Usage:

.. code-block:: python

    ctx = flytekit.current_context().distributed_training_context
    # OR
    ctx = flytekit.current_context().get(sagemaker.DISTRIBUTED_TRAINING_CONTEXT_KEY)

"""
