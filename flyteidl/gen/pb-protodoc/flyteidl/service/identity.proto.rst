.. _api_file_flyteidl/service/identity.proto:

identity.proto
===============================

.. _api_msg_flyteidl.service.UserInfoRequest:

flyteidl.service.UserInfoRequest
--------------------------------

`[flyteidl.service.UserInfoRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/service/identity.proto#L8>`_


.. code-block:: json

  {}




.. _api_msg_flyteidl.service.UserInfoResponse:

flyteidl.service.UserInfoResponse
---------------------------------

`[flyteidl.service.UserInfoResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/service/identity.proto#L11>`_

See the OpenID Connect spec at https://openid.net/specs/openid-connect-core-1_0.html#UserInfoResponse for more information.

.. code-block:: json

  {
    "subject": "...",
    "name": "...",
    "preferred_username": "...",
    "given_name": "...",
    "family_name": "...",
    "email": "...",
    "picture": "..."
  }

.. _api_field_flyteidl.service.UserInfoResponse.subject:

subject
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Locally unique and never reassigned identifier within the Issuer for the End-User, which is intended to be consumed
  by the Client.
  
  
.. _api_field_flyteidl.service.UserInfoResponse.name:

name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Full name
  
  
.. _api_field_flyteidl.service.UserInfoResponse.preferred_username:

preferred_username
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Shorthand name by which the End-User wishes to be referred to
  
  
.. _api_field_flyteidl.service.UserInfoResponse.given_name:

given_name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Given name(s) or first name(s)
  
  
.. _api_field_flyteidl.service.UserInfoResponse.family_name:

family_name
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Surname(s) or last name(s)
  
  
.. _api_field_flyteidl.service.UserInfoResponse.email:

email
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Preferred e-mail address
  
  
.. _api_field_flyteidl.service.UserInfoResponse.picture:

picture
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Profile picture URL
  
  

