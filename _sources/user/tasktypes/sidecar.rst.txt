.. _sidecar-task-type:

#############
Sidecar Tasks
#############

Sidecar tasks can be used to bring up multiple containers within a single task. Sidecar tasks are defined using a Kubernetes `pod spec <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#podspec-v1-core>`_ but differ slightly in that the plugin task executor solely monitors the status of a user-specified primary container over the task lifecycle.

************
Installation
************
If you haven't already, install the sidecar extra from flytekit like so:

.. code-block:: text

   pip install flytekit[sidecar]


And assert that you have a dependency in your project on 

.. code-block:: text

   k8s-proto>=0.0.2


*****
Usage
*****

Sidecar tasks accept all arguments that ordinary :ref:`container tasks <container-task-type>` accept. However, sidecar tasks require two additional arguments, ``pod_spec`` and ``primary_container_name``

Pod Spec
========

Using the `generated python protobuf code in flyteproto <https://github.com/lyft/flyteproto>`_, a task can define a completely kubernetes-native pod spec that will be deployed as part of the sidecar task execution.


Primary container
================================================================================

This is a required name you use to distinguish your primary container. The code in the body of the task definition will be injected in the primary container. The pod spec you pass to the task definition does not necessarily need to include a container definition with the primary container, but if you'd like to modify the primary container by setting a shared volume mount for example, you can do so in the pod spec.

For primary containers defined in the pod spec, a few caveats apply. The following `container <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#container-v1-core>`_
fields will be overwritten at registration time and therefore are meaningless if set:

* image
* command
* args
* resource requirements
* env

Resource requirements and env will use the values set in the sidecar task definition and are therefore still customizable.

For example:

.. code-block:: python

   def get_pod_spec():
       my_pod_spec = generated_pb2.PodSpec()
       my_container = generated_pb2.Container(name="primary")
       # apply more customization to my_container as desired

       my_pod_spec.containers.extend([my_container])
       # apply more customization to my_pod_spec as desired
       
       return my_pod_spec
      

   @inputs(in1=Types.String)
   @outputs(out1=Types.String)
   @sidecar_task(
       cpu_request='10',
       cpu_limit='20',
       environment={"foo": "bar"},
       pod_spec=get_pod_spec(),
       primary_container_name="primary",
   )
   def simple_sidecar_task(wf_params, in1, out1):
       print("Hi, {} I'll be run in a sidecar task!".format(in1))
       ...


***************
Working Example
***************

For a trivial example of configuring multiple containers so that one writes to a shared volume mount and the second waits until it receives the message, see:

.. code-block:: python

   from __future__ import absolute_import
   from __future__ import print_function

   import time
   import os

   from flytekit.sdk.tasks import sidecar_task
   from k8s.io.api.core.v1 import generated_pb2


   # A simple pod spec in which a shared volume is mounted in both the primary and secondary containers. The secondary
   # writes a file that the primary waits on before completing.
   def generate_pod_spec_for_task():
       pod_spec = generated_pb2.PodSpec()

       primary_container = generated_pb2.Container(name="primary")

       secondary_container = generated_pb2.Container(
	         name="secondary",
	         image="alpine",
       )
       secondary_container.command.extend(["/bin/sh"])
       secondary_container.args.extend(["-c", "echo hi sidecar world > /data/message.txt"])
       shared_volume_mount = generated_pb2.VolumeMount(
	         name="shared-data",
	         mountPath="/data",
       )
       secondary_container.volumeMounts.extend([shared_volume_mount])
       primary_container.volumeMounts.extend([shared_volume_mount])

       pod_spec.volumes.extend([generated_pb2.Volume(
	          name="shared-data",
	          volumeSource=generated_pb2.VolumeSource(
	              emptyDir=generated_pb2.EmptyDirVolumeSource(
		                 medium="Memory",
	              )
	          )
       )])
       pod_spec.containers.extend([primary_container, secondary_container])
       return pod_spec


   @sidecar_task(
       pod_spec=generate_pod_spec_for_task(),
       primary_container_name="primary",
   )
   def my_sidecar_task(wfparams):
       # The code defined in this task will get injected into the primary container.
       while not os.path.isfile('/data/message.txt'):
           time.sleep(5)

