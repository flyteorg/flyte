.. _design-control-plane:

###################################################
FlyteRemote: A Programmatic Control Plane Interface
###################################################

.. tags:: Remote, Basic

For those who require programmatic access to the control plane, the :mod:`~flytekit.remote` module enables you to perform
certain operations in a Python runtime environment.

Since this section naturally deals with the control plane, this discussion is only relevant for those who have a Flyte
backend set up and have access to it (a local demo cluster will suffice as well).

*****************************
Creating a FlyteRemote Object
*****************************

The :class:`~flytekit.remote.remote.FlyteRemote` class is the entrypoint for programmatically performing operations in a Python
runtime. It can be initialized by passing in the:

* :py:class:`~flytekit.configuration.Config` object: the parent configuration object that holds all the configuration information to connect to the Flyte backend.
* :py:attr:`~flytekit.remote.remote.FlyteRemote.default_project`: the default project to use when fetching or executing flyte entities.
* :py:attr:`~flytekit.remote.remote.FlyteRemote.default_domain`: the default domain to use when fetching or executing flyte entities.
* :py:attr:`~flytekit.remote.remote.FlyteRemote.file_access`: the file access provider to use for offloading non-literal inputs/outputs.
* ``kwargs``: additional arguments that need to be passed to create ``SynchronousFlyteClient``.

A :class:`~flytekit.remote.remote.FlyteRemote` object can be created in various ways:

Auto
====

The :py:class:`~flytekit.configuration.Config` class's :py:meth:`~flytekit.configuration.Config.auto` method can be used to automatically
construct the ``Config`` object.

.. code-block:: python

    from flytekit.remote import FlyteRemote
    from flytekit.configuration import Config

    remote = FlyteRemote(config=Config.auto())

``auto`` also accepts a ``config_file`` argument, which is the path to the configuration file to use.
The order of precedence that ``auto`` follows is:

* Finds all the environment variables that match the configuration variables.
* If no environment variables are set, it looks for a configuration file at the path specified by the ``config_file`` argument.
* If no configuration file is found, it uses the default values.

Sandbox
=======

The :py:class:`~flytekit.configuration.Config` class's :py:meth:`~flytekit.configuration.Config.for_sandbox` method can be used to
construct the ``Config`` object, specifically to connect to the Flyte cluster.

.. code-block:: python

    from flytekit.remote import FlyteRemote
    from flytekit.configuration import Config

    remote = FlyteRemote(config=Config.for_sandbox())

The initialization is as simple as calling ``for_sandbox()`` on the ``Config`` class!
This, by default, uses ``localhost:30081`` as the endpoint, and the default minio credentials.

If the sandbox is in a hosted-like environment, then *port-forward* or *ingress URLs* need to be taken care of.

Any Endpoint
============

The :py:class:`~flytekit.configuration.Config` class's :py:meth:`~flytekit.configuration.Config.for_endpoint` method can be used to
construct the ``Config`` object to connect to a specific endpoint.

.. code-block:: python

    from flytekit.remote import FlyteRemote
    from flytekit.configuration import Config

    remote = FlyteRemote(
        config=Config.for_endpoint(endpoint="flyte.example.net"),
        default_project="flytesnacks",
        default_domain="development",
    )

The ``for_endpoint`` method also accepts:

* ``insecure``: whether to use insecure connections. Defaults to ``False``.
* ``data_config``: can be used to configure how data is downloaded or uploaded to a specific blob storage like S3, GCS, etc.
* ``config_file``: the path to the configuration file to use.

.. _general_initialization:

Generalized Initialization
==========================

The :py:class:`~flytekit.configuration.Config` class can be directly used to construct the ``Config`` object if additional configuration is needed.
You can send :py:class:`~flytekit.configuration.PlatformConfig`, :py:class:`~flytekit.configuration.DataConfig`,
:py:class:`~flytekit.configuration.SecretsConfig`, and :py:class:`~flytekit.configuration.StatsConfig` objects to the ``Config`` class.

.. list-table:: ``Config`` Attributes
   :widths: 50 50

   * - ``PlatformConfig``
     - Settings to talk to a Flyte backend.
   * - ``DataConfig``
     - Any data storage specific configuration.
   * - ``SecretsConfig``
     - Configuration for secrets.
   * - ``StatsConfig``
     - Configuration for sending statsd.

For example:

.. code-block:: python

    from flytekit.remote import FlyteRemote
    from flytekit.configuration import Config, PlatformConfig

    remote = FlyteRemote(
        config=Config(
            platform=PlatformConfig(
                endpoint="flyte.example.net",
                insecure=False,
                client_id="my-client-id",
                client_credentials_secret="my-client-secret",
                auth_mode="client_credentials",
            ),
            secrets=SecretsConfig(default_dir="/etc/secrets"),
        )
    )

*****************
Fetching Entities
*****************

Tasks, workflows, launch plans, and executions can be fetched using FlyteRemote.

.. code-block:: python

    flyte_task = remote.fetch_task(name="my_task", version="v1")
    flyte_workflow = remote.fetch_workflow(name="my_workflow", version="v1")
    flyte_launch_plan = remote.fetch_launch_plan(name="my_launch_plan", version="v1")
    flyte_execution = remote.fetch_execution(name="my_execution")

``project`` and ``domain`` can also be specified in all the ``fetch_*`` calls.
If not specified, the default values given during the creation of the FlyteRemote object will be used.

The following is an example that fetches :py:func:`~flytekit.task`s and creates a :py:func:`~flytekit.workflow`:

.. code-block:: python

    from flytekit import workflow

    task_1 = remote.fetch_task(name="core.basic.hello_world.say_hello", version="v1")
    task_2 = remote.fetch_task(
        name="core.basic.lp.greet",
        version="v13",
        project="flytesnacks",
        domain="development",
    )


    @workflow
    def my_remote_wf(name: str) -> int:
        return task_2(task_1(name=name))

Another example that dynamically creates a launch plan for the ``my_remote_wf`` workflow:

.. code-block:: python

    from flytekit import LaunchPlan

    flyte_workflow = remote.fetch_workflow(
        name="my_workflow", version="v1", project="flytesnacks", domain="development"
    )
    launch_plan = LaunchPlan.get_or_create(name="my_launch_plan", workflow=flyte_workflow)

********************
Registering Entities
********************

Tasks, workflows, and launch plans can be registered using FlyteRemote.

.. code-block:: python

    from flytekit.configuration import SerializationSettings

    flyte_entity = ...
    flyte_task = remote.register_task(
        entity=flyte_entity,
        serialization_settings=SerializationSettings(image_config=None),
        version="v1",
    )
    flyte_workflow = remote.register_workflow(
        entity=flyte_entity,
        serialization_settings=SerializationSettings(image_config=None),
        version="v1",
    )
    flyte_launch_plan = remote.register_launch_plan(entity=flyte_entity, version="v1")

* ``entity``: the entity to register.
* ``version``: the version that will be used to register. If not specified, the version used in serialization settings will be used.
* ``serialization_settings``: the serialization settings to use. Refer to :py:class:`~flytekit.configuration.SerializationSettings` to know all the acceptable parameters.

All the additional parameters which can be sent to the ``register_*`` methods can be found in the documentation for the corresponding method:
:py:meth:`~flytekit.remote.remote.FlyteRemote.register_task`, :py:meth:`~flytekit.remote.remote.FlyteRemote.register_workflow`,
and :py:meth:`~flytekit.remote.remote.FlyteRemote.register_launch_plan`.

The :py:class:`~flytekit.configuration.SerializationSettings` class accepts :py:class:`~flytekit.configuration.ImageConfig` which
holds the available images to use for the registration.

The following example showcases how to register a workflow using an existing image if the workflow is created locally:

.. code-block:: python

    from flytekit.configuration import ImageConfig

    img = ImageConfig.from_images(
        "docker.io/xyz:latest", {"spark": "docker.io/spark:latest"}
    )
    wf2 = remote.register_workflow(
        my_remote_wf,
        serialization_settings=SerializationSettings(image_config=img),
        version="v1",
    )

******************
Executing Entities
******************

You can execute a task, workflow, or launch plan using :meth:`~flytekit.remote.remote.FlyteRemote.execute` method
which returns a :class:`~flytekit.remote.executions.FlyteWorkflowExecution` object.
For more information on Flyte entities, see the :ref:`remote flyte entities <remote-flyte-execution-objects>` reference.

.. code-block:: python

    flyte_entity = ...  # one of FlyteTask, FlyteWorkflow, or FlyteLaunchPlan
    execution = remote.execute(
        flyte_entity, inputs={...}, execution_name="my_execution", wait=True
    )

* ``inputs``: the inputs to the entity.
* ``execution_name``: the name of the execution. This is useful to avoid de-duplication of executions.
* ``wait``: synchronously wait for the execution to complete.

Additional arguments include:

* ``project``: the project on which to execute the entity.
* ``domain``: the domain on which to execute the entity.
* ``type_hints``: a dictionary mapping Python types to their corresponding Flyte types.
* ``options``: options can be configured for a launch plan during registration or overridden during execution. Refer to :py:class:`~flytekit.remote.remote.Options` to know all the acceptable parameters.

The following is an example demonstrating how to use the :py:class:`~flytekit.remote.remote.Options` class to configure a Flyte entity:

.. code-block:: python

    from flytekit.models.common import AuthRole, Labels
    from flytekit.tools.translator import Options

    flyte_entity = ...  # one of FlyteTask, FlyteWorkflow, or FlyteLaunchPlan
    execution = remote.execute(
        flyte_entity,
        inputs={...},
        execution_name="my_execution",
        wait=True,
        options=Options(
            raw_data_prefix="s3://my-bucket/my-prefix",
            auth_role=AuthRole(assumable_iam_role="my-role"),
            labels=Labels({"my-label": "my-value"}),
        ),
    )

**********************************
Retrieving & Inspecting Executions
**********************************

After an execution is completed, you can retrieve the execution using the :meth:`~flytekit.remote.remote.FlyteRemote.fetch_execution` method.
The fetched execution can be used to retrieve the inputs and outputs of an execution.

.. code-block:: python

    execution = remote.fetch_execution(
        name="fb22e306a0d91e1c6000", project="flytesnacks", domain="development"
    )
    input_keys = execution.inputs.keys()
    output_keys = execution.outputs.keys()

The ``inputs`` and ``outputs`` correspond to the top-level execution or the workflow itself.

To fetch a specific output, say, a model file:

.. code-block:: python

    model_file = execution.outputs["model_file"]
    with open(model_file) as f:
        # use mode
        ...

You can use :meth:`~flytekit.remote.remote.FlyteRemote.sync` to sync the entity object's state with the remote state during the execution run:

.. code-block:: python

    synced_execution = remote.sync(execution, sync_nodes=True)
    node_keys = synced_execution.node_executions.keys()

.. note::

    During the sync, you may come across ``Received message larger than max (xxx vs. 4194304)`` error if the message size is too large. In that case, edit the ``flyte-admin-base-config`` config map using the command ``kubectl edit cm flyte-admin-base-config -n flyte`` to increase the ``maxMessageSizeBytes`` value. Refer to the :ref:`troubleshooting guide <troubleshoot>` in case you've queries about the command's usage.

``node_executions`` will fetch all the underlying node executions recursively.

To fetch output of a specific node execution:

.. code-block:: python

    node_execution_output = synced_execution.node_executions["n1"].outputs["model_file"]

:ref:`Node <divedeep-nodes>` here, can correspond to a task, workflow, or branch node.

****************
Listing Entities
****************

To list the recent executions, use the :meth:`~flytekit.remote.remote.FlyteRemote.recent_executions` method.

.. code-block:: python

    recent_executions = remote.recent_executions(project="flytesnacks", domain="development", limit=10)

The ``limit`` parameter is optional and defaults to 100.

To list tasks by version, use the :meth:`~flytekit.remote.remote.FlyteRemote.list_tasks_by_version` method.

.. code-block:: python

    tasks = remote.list_tasks_by_version(project="flytesnacks", domain="development", version="v1")

************************
Terminating an Execution
************************

To terminate an execution, use the :meth:`~flytekit.remote.remote.FlyteRemote.terminate` method.

.. code-block:: python

    execution = remote.fetch_execution(name="fb22e306a0d91e1c6000", project="flytesnacks", domain="development")
    remote.terminate(execution, cause="Code needs to be updated")
