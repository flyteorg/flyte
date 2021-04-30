.. _api_file_flyteidl/service/auth.proto:

auth.proto
===========================

.. _api_msg_flyteidl.service.OAuth2MetadataRequest:

flyteidl.service.OAuth2MetadataRequest
--------------------------------------

`[flyteidl.service.OAuth2MetadataRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/service/auth.proto#L21>`_


.. code-block:: json

  {}




.. _api_msg_flyteidl.service.OAuth2MetadataResponse:

flyteidl.service.OAuth2MetadataResponse
---------------------------------------

`[flyteidl.service.OAuth2MetadataResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/service/auth.proto#L25>`_

OAuth2MetadataResponse defines an RFC-Compliant response for /.well-known/oauth-authorization-server metadata
as defined in https://tools.ietf.org/html/rfc8414

.. code-block:: json

  {
    "issuer": "...",
    "authorization_endpoint": "...",
    "token_endpoint": "...",
    "response_types_supported": [],
    "scopes_supported": [],
    "token_endpoint_auth_methods_supported": [],
    "jwks_uri": "...",
    "code_challenge_methods_supported": [],
    "grant_types_supported": []
  }

.. _api_field_flyteidl.service.OAuth2MetadataResponse.issuer:

issuer
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Defines the issuer string in all JWT tokens this server issues. The issuer can be admin itself or an external
  issuer.
  
  
.. _api_field_flyteidl.service.OAuth2MetadataResponse.authorization_endpoint:

authorization_endpoint
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) URL of the authorization server's authorization endpoint [RFC6749]. This is REQUIRED unless no grant types are
  supported that use the authorization endpoint.
  
  
.. _api_field_flyteidl.service.OAuth2MetadataResponse.token_endpoint:

token_endpoint
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) URL of the authorization server's token endpoint [RFC6749].
  
  
.. _api_field_flyteidl.service.OAuth2MetadataResponse.response_types_supported:

response_types_supported
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Array containing a list of the OAuth 2.0 "response_type" values that this authorization server supports.
  
  
.. _api_field_flyteidl.service.OAuth2MetadataResponse.scopes_supported:

scopes_supported
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) JSON array containing a list of the OAuth 2.0 [RFC6749] "scope" values that this authorization server supports.
  
  
.. _api_field_flyteidl.service.OAuth2MetadataResponse.token_endpoint_auth_methods_supported:

token_endpoint_auth_methods_supported
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) JSON array containing a list of client authentication methods supported by this token endpoint.
  
  
.. _api_field_flyteidl.service.OAuth2MetadataResponse.jwks_uri:

jwks_uri
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) URL of the authorization server's JWK Set [JWK] document. The referenced document contains the signing key(s) the
  client uses to validate signatures from the authorization server.
  
  
.. _api_field_flyteidl.service.OAuth2MetadataResponse.code_challenge_methods_supported:

code_challenge_methods_supported
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) JSON array containing a list of Proof Key for Code Exchange (PKCE) [RFC7636] code challenge methods supported by
  this authorization server.
  
  
.. _api_field_flyteidl.service.OAuth2MetadataResponse.grant_types_supported:

grant_types_supported
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) JSON array containing a list of the OAuth 2.0 grant type values that this authorization server supports.
  
  


.. _api_msg_flyteidl.service.PublicClientAuthConfigRequest:

flyteidl.service.PublicClientAuthConfigRequest
----------------------------------------------

`[flyteidl.service.PublicClientAuthConfigRequest proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/service/auth.proto#L58>`_


.. code-block:: json

  {}




.. _api_msg_flyteidl.service.PublicClientAuthConfigResponse:

flyteidl.service.PublicClientAuthConfigResponse
-----------------------------------------------

`[flyteidl.service.PublicClientAuthConfigResponse proto] <https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/service/auth.proto#L61>`_

FlyteClientResponse encapsulates public information that flyte clients (CLIs... etc.) can use to authenticate users.

.. code-block:: json

  {
    "client_id": "...",
    "redirect_uri": "...",
    "scopes": [],
    "authorization_metadata_key": "..."
  }

.. _api_field_flyteidl.service.PublicClientAuthConfigResponse.client_id:

client_id
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) client_id to use when initiating OAuth2 authorization requests.
  
  
.. _api_field_flyteidl.service.PublicClientAuthConfigResponse.redirect_uri:

redirect_uri
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) redirect uri to use when initiating OAuth2 authorization requests.
  
  
.. _api_field_flyteidl.service.PublicClientAuthConfigResponse.scopes:

scopes
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) scopes to request when initiating OAuth2 authorization requests.
  
  
.. _api_field_flyteidl.service.PublicClientAuthConfigResponse.authorization_metadata_key:

authorization_metadata_key
  (`string <https://developers.google.com/protocol-buffers/docs/proto#scalar>`_) Authorization Header to use when passing Access Tokens to the server. If not provided, the client should use the
  default http `Authorization` header.
  
  

