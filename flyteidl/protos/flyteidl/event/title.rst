##############################################
Flyte Internal and External Eventing interface
##############################################

This section contains all the protocol buffer definitions for Internal and
External Eventing system.

Flyte Internal Eventing
========================

This is the interface used by the dataplane (execution engine) to communicate
workflow and task progress back to the **control plane** admin service. 

Flyte External Eventing - Event Egress
=======================================

This refers to the interface for all the event messages exiting from the Flyte
**control plane** and delivered on the configured pubsub channel.


