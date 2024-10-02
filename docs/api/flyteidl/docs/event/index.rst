
##############################################
Flyte Internal and External Eventing interface
##############################################

This section contains all the protocol buffer definitions for Internal and
External Eventing system.

Flyte Internal Eventing
========================

This is the interface used by the dataplane (execution engine) to communicate with the control plane admin service about the workflow and task progress.

Flyte External Eventing - Event Egress
=======================================

This refers to the interface for all the event messages leaving the Flyte
**control plane** and reaching on the configured pubsub channel.

`Event raw proto <https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/event/event.proto>`__

.. toctree::
	:maxdepth: 1
	:caption: event
	:name: eventtoc

	event
