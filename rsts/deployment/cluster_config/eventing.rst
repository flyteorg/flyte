.. _deployment-cluster-config-eventing:

#################
Platform Events
#################

Progress of Flyte workflow and task execution is delimited by a series of events that are passed from the Flyte Propeller to Flyte Admin. Administrators can configure Flyte Admin to send these events onwards to a pub/sub system like SNS/SQS as well. Note that this configuration is distinct from the configuration for notifications :ref:`deployment-cluster-config-notifications`. They should use separate topics/queues. These events are meant for external consumption, outside the Flyte platform, whereas the notifications pub/sub setup is entirely for Admin itself to send email/pagerduty/etc notifications.

*************
Configuration
*************

To turn on, add the following to your Flyte Admin

.. code:: yaml

  external_events.yaml: |
    externalEvents:
      enable: true
      aws:
        region: us-east-2
      eventsPublisher:
        eventTypes:
        - all
        topicName: arn:aws:sns:us-east-2:123456:123-my-topic
      type: aws

Helm
======
There should already be a section for this in the ``values.yaml`` file. Update the settings under the ``external_events`` key and turn ``enable`` to ``true``. The same flag is used for Helm as for Admin itself.

*****
Usage
*****

The events emitted will be base64 encoded binary representation of the following IDL messages:

* ``admin_event_pb2.TaskExecutionEventRequest``
* ``admin_event_pb2.NodeExecutionEventRequest``
* ``admin_event_pb2.WorkflowExecutionEventRequest``

Which of these three events is being sent can be distinguished by the subject line of the message, which will be one of the three strings above.

Note that these message wrap the underlying event messages :std:doc:`found here <flyteidl:protos/docs/event/event>`.
