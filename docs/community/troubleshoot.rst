.. _troubleshoot:

=====================
Troubleshooting Guide
=====================

.. tags:: Troubleshoot, Basic

The content in this section will help Flyte users isolate the most probable causes for some of the common issues that could arise while getting started with the project.

Before getting started, collect the following information from the underlying infrastructure:

- Capture the ``Status`` column from the output of: 

.. prompt:: bash $

 kubectl describe pod <PodName> -n <namespace>

Where <PodName> will typically correspond to the node execution string that you can find in the UI.

- Pay close attention to the `Events` section in the output.
- Also, collect the logs from the Pod:

.. prompt:: bash $

 kubectl logs pods -n <namespace>

Where <namespace> will typically correspond to the Flyte <project>-<domain>, e.g. flytesnacks-development.

Depending on the contents of the logs or the `Events`, you can try different things:

Debugging common execution errors
----------------------------------

``message: '0/1 nodes are available: 1 Insufficient cpu. preemption: 0/1 nodes are available: 1 No preemption victims found for incoming pod.'``
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

This issue is more common on MacOS devices. Make sure that your Docker daemon has allocated a minimum of 4 CPU cores and 3GB of RAM

``terminated with exit code (137). Reason [OOMKilled]``
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- For single binary environment deployed with Helm chart, make sure you are using `the most recent charts <https://github.com/flyteorg/flyte/tree/master/charts>`_

- For EKS deployments, you cand adjust resource limits and requests in the `inline <https://github.com/flyteorg/flyte/blob/d60c9af85a59ebb4c2265f76cb082b992078a309/charts/flyte-binary/eks-production.yaml#L30>`_ section of the ``eks-production.yaml`` file. Example:

.. code-block:: yaml

  inline: 
    task_resources:
      defaults:
        cpu: 100m
        memory: 100Mi
        storage: 100Mi
      limits:
        memory: 1Gi

- Also, the default container resource limits are can be overridden from the task itself:

.. code-block:: python

      from flytekit import Resources, task
      @task(limits=Resources(mem="256Mi")    
      def your_task(...

``Error: ImagePullBackOff``
^^^^^^^^^^^^^^^^^^^^^^^^^^^

- If your environment requires the use of a network proxy use the ``--env`` option when starting the sandbox and pass the proxy configuration:

.. prompt:: bash $

 flytectl demo start --env HTTP_PROXY=<your-proxy-IP>

- If you're building a custom Docker image, make sure to use a tag other than ``latest``. Otherwise, the Kubernetes default pull policy will be changed from ``IfNotPresent`` to ``Always``, forcing an image pull with every Pod deployment.

Issues running workloads
-------------------------

``OPENSSL_internal:WRONG_VERSION_NUMBER`` 
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- For ``flyte-binary``: make sure that the endpoint name you have set in your ``config.yaml`` file, is included in the DNS names of the SSL certificate installed (be it self signed or issued by a Certificate Authority)
-  For ``sandbox``: verify the ``FLYTECTL_CONFIG`` environment variable has the correct value by running:

.. prompt:: bash $

 export FLYTECTL_CONFIG=~/.flyte/config-sandbox.yaml

``ModuleNotFoundError``
^^^^^^^^^^^^^^^^^^^^^^^

- If you're using a custom container image and using Docker, make sure your ``Dockerfile`` is located at the same level of the ``flyte`` directory and that there is an empty ``__init__.py`` file in your project's folder :

.. prompt::

 myflyteapp
 ├── Dockerfile
 ├── docker_build_and_tag.sh
 ├── flyte
 │         ├── __init__.py
 │         └── workflows
 │             ├── __init__.py
 │             └── example.py
 └── requirements.txt

``An error occurred (AccessDenied) when calling the PutObject operation`` in an EKS deployment
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- Make sure that the Kubernetes service account Flyte is using has the annotation that refers to the IAM Role is connected to:

.. prompt:: bash $

 kubectl describe sa <my-flyte-sa> -n <flyte-namespace>

Example output:

.. prompt::

 Name:                <my-flyte-sa>
 Namespace:           flyte
 Labels:              app.kubernetes.io/managed-by=eksctl
 Annotations:         eks.amazonaws.com/role-arn: arn:aws:iam::<aws-account-id>:role/flyte-system-role
 Image pull secrets:  <none>
 Mountable secrets:   <none>
 Tokens:              <none>
 Events:              <none>

- Otherwise, obtain your IAM role's ARN and manually annotate the service account:

.. prompt:: bash $

 kubectl annotate serviceaccount -n <flyte-namespace> <http://eks.amazonaws.com/role-arn=arn:aws:iam::xxxx:role/<flyte-iam-role>eks.amazonaws.com/role-arn=arn:aws:iam::xxxx:role/<flyte-iam-role>

- Refer to this community-maintained `guides <https://github.com/davidmirror-ops/flyte-the-hard-way/blob/main/docs/03-roles-service-accounts.md>`_ for further information about Flyte deployment on EKS

``FlyteScopedUserException: 'JavaPackage' object is not callable`` when running a Spark task
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Please add ``spark`` to the list of `enabled-plugins` in the config yaml file. For example,

.. code-block:: yaml

  tasks:
    task-plugins:
      enabled-plugins:
        - container
        - sidecar
        - K8S-ARRAY
        - spark
      default-for-task-types:
        - container: container
        - container_array: K8S-ARRAY
