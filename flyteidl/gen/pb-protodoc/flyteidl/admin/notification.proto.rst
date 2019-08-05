.. _api_file_flyteidl/admin/notification.proto:

notification.proto
=================================

import "validate/validate.proto";

.. _api_msg_flyteidl.admin.EmailMessage:

flyteidl.admin.EmailMessage
---------------------------

`[flyteidl.admin.EmailMessage proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/admin/notification.proto#L11>`_

Represents the Email object that is sent to a publisher/subscriber
to forward the notification.
Note: This is internal to Admin and doesn't need to be exposed to other components.

.. code-block:: json

  {
    "recipients_email": [],
    "sender_email": "...",
    "subject_line": "...",
    "body": "..."
  }

.. _api_field_flyteidl.admin.EmailMessage.recipients_email:

recipients_email
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) The list of email addresses to receive an email with the content populated in the other fields.
  Currently, each email recipient will receive its own email.
  This populates the TO field.
  
  
.. _api_field_flyteidl.admin.EmailMessage.sender_email:

sender_email
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) The email of the sender.
  This populates the FROM field.
  
  
.. _api_field_flyteidl.admin.EmailMessage.subject_line:

subject_line
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) The content of the subject line.
  This populates the SUBJECT field.
  
  
.. _api_field_flyteidl.admin.EmailMessage.body:

body
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) The content of the email body.
  This populates the BODY field.
  
  

