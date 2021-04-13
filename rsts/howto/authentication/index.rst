.. _howto_authentication:

############################
How to Set Up Authentication
############################

Flyte Admin ships with a canonical implementation of OAuth2, integrating seamlessly into an organization's existing identity provider.  At Lyft, we use Okta as the IDP, but if you have issues integrating with another implementation of the OAuth server, please open an issue.

***********
Components
***********

While the most obvious interaction with the Flyte control plane is through the web based UI, there are other critical components of Flyte that also need to be considered. These components should be thought of as third-party services even though the Flyte codebase provides them.

Flyte CLI
=========
This is the primary component. Flyte CLI is the command-line entrypoint to Flyte Admin and is used by both administrators and users more comfortable in the command line, or are running in a headless OS.

The IDP application corresponding to the CLI will need to support PKCE.

Direct Client Access
====================
The gRPC client provided by the Flyte IDL, or direct calls to the HTTP endpoints on Admin from within a running script are ways that we have seen users hit the control plane directly.  We generally discourage this course of action as it leads to a possible self-imposed DOS vector, which is generally made from within a running workflow itself. 

For instance, a Flyte task can fetch the definition for a launch plan associated with a completely different workflow, and then launch an execution of it. This is not the correct way to launch one workflow from another but for the time being remains possible.

*****************
Swimlane Diagrams
*****************

Flyte Admin authentication is implemented using the authorization code flow.

Flyte UI Flow
=============
https://swimlanes.io/d/OmV4ybCkx

.. image:: flyte_ui_flow.png
   :width: 600
   :alt: Flyte UI Swimlane


Flyte CLI Flow
==============
https://swimlanes.io/d/q64OxuoxT

.. image:: flyte_cli_flow.png
  :width: 600
  :alt: Flyte CLI Swimlane

*************
Configuration
*************

IDP Configuration
=================
Flyte Admin will require that the application in your identity provider be configured without PKCE and with a client secret. It should also be configured with a refresh token.

Flyte Admin Configuration
=========================
Please refer to the `inline documentation <https://github.com/flyteorg/flyteadmin/blob/eaca2fb0e6018a2e261e9e2da8998906477cadb5/pkg/auth/config/config.go>`_ on the ``Config`` object in the ``auth`` package for a discussion on the settings required.

***********************
Example Configurations
***********************

OpenID Connect
===============

OpenID Connect allows users to authenticate to flyte in their browser using a familiar authentication provider (perhaps an organization-wide configured IdP).
Flyte supports connecting with external OIdC providers. Here are some examples for how to set these up:

Google OpenID Connect
----------------------

Follow `Google Docs <https://developers.google.com/identity/protocols/oauth2/openid-connect>`__ on how to configure the IdP for OpenIDConnect.

.. note::

  Make sure to create an OAuth2 Client Credential. The `client_id` and `client_secret` will be needed in the following
  steps.

Okta OpenID Connect
-------------------

Okta supports OpenIDConnect protocol and the creation of custom OAuth2 Authorization Servers, allowing it to act as both the user and apps IdP.
It offers more detailed control on access policies, users' consent, and app management.

1. If you don't already have an Okta account, sign up for one `here <https://developer.okta.com/signup/>`__.
2. Create an app (choose Web for the platform) and OpenID Connect for the sign on method.
3. Add Login redirect URIs (e.g. http://localhost:30081/callback for sandbox or https://<your deployment url>/callback)
4. OPTIONAL: Add logout redirect URIs (e.g. http://localhost:30081/logout for sandbox)
5. Note down the Client ID and Client Secret

Apply configuration
===================

1. Store the `client_secret` in a k8s secrt as follows:

.. prompt:: bash

  kubectl edit secret -n flyte flyte-admin-auth

Add a new key under `stringData`:

.. code-block:: yaml

  stringData:
    oidc_client_secret: <client_secret> from the previous step
  data:
    ...

Save and close your editor.

2. Edit FlyteAdmin config to add `client_id`, `client_secret` and configure auth as follows:

.. prompt:: bash

  kubectl deploy -n flyte flyteadmin -o yaml | grep "name: flyte-admin-config"

This will output the name of the config map where the `client_id` needs to go.

.. prompt:: bash

  kubectl edit configmap -n flyte <the name of the config map from previous command>

Follow the inline comments to make the necessary changes:

.. code-block:: yaml

  server:
    httpPort: 8088
    grpcPort: 8089
    grpcServerReflection: true
    kube-config: /Users/haythamabuelfutuh/.kube/config
    security:
      secure: false
      # 1. Enable Auth by turning this to true
      useAuth: true
      allowCors: true
      allowedOrigins:
        # Accepting all domains for Sandbox installation
        - "*"
      allowedHeaders:
        - "Content-Type"
  auth:
    # 2. Update with the public facing url of flyte admin (e.g. https://flyte.mycompany.io/)
    httpPublicUri: http://localhost:8088/
    userAuth:
      openId:
        # 3. Put the URL of the OpenID Connect provider.
        #    baseUrl: https://accounts.google.com/ # Uncomment for Google
        baseUrl: https://dev-14186422.okta.com/oauth2/default # Okta with a custom Authorization Server
        scopes:
          - profile
          - openid
          # - offline_access # Uncomment if OIdC supports issuing refresh tokens.
        # 4. Replace with the client ID created for Flyte.
        clientId: 0oakkheteNjCMERst5d6
        # 5. Replace with the public facing URL of flyte admin (e.g. https://flyte.mycompany.io/callback)
        callbackUrl: "http://localhost:8088/callback"
        # 6. Replace with the flyte console's URL (e.g. https://flyte.mycompany.io/console) 
        redirectUrl: "/api/v1/projects"

Save and exit your editor.

3. Restart `flyteadmin` for the changes to take effect:

.. prompt:: bash

  kubectl rollout restart deployment/flyteadmin -n flyte

OAuth2 Authorization Server
===========================

Flyte Admin comes with a built-in authorization server that can be statically configured with a set of clients to request and act on behalf of the user.

Okta IdP
--------

1. Under security -> API, click `Add Authorization Server`. Set the audience to the public URL of flyte admin (e.g. https://flyte.mycompany.io/).
2. Under `Access Policies`, click `Add New Access Policy` and walk through the wizard to allow access to the authorization server.
3. Under `Scopes`, click `Add Scope`. Set the name to `all` (required) and check `Require user consent for this scope` (recommended).

Apply Configurations
--------------------

It is possible to direct flyte admin to use an external authorization server. To do so, edit the same config map once more and follow these changes:

.. code-block:: yaml

  auth:
    # 1. Update with the public facing URL of flyte admin (e.g. https://flyte.mycompany.io/)
    httpPublicUri: http://localhost:8088/
    appAuth:
      # 1. Choose External if you will use an external Authorization Server (e.g. a Custom Authorization server in Okta)
      #    Choose Self (or omit the value) to use Flyte Admin's internal (albeit limited) Authorization Server.
      authServerType: External
      thirdPartyConfig:
        flyteClient:
          # 2. Replace with a new native client ID provisioned in the custom authorization server
          clientId: flytectl
          redirectUri: https://localhost:53593/callback
          # 3. "all" is a required scope and must be configured in the custom authorization server
          scopes:
            - offline
            - all
    userAuth:
      openId:
        # 4. Use the URL of your custom authorization server created above:
        baseUrl: https://dev-14186422.okta.com/oauth2/auskngnn7uBViQq6b5d6 # Okta with a custom Authorization Server
        scopes:
          - profile
          - openid
          # - offline_access # Uncomment if OIdC supports issuing refresh tokens.
        clientId: 0oakkheteNjCMERst5d6
        callbackUrl: "http://localhost:8088/callback"
        redirectUrl: "/api/v1/projects"

***************************
Continuous Integration - CI
***************************

If your organization does any automated registration, then you'll need to authenticate with the `basic authentication <https://tools.ietf.org/html/rfc2617>`_ flow (username and password effectively) as CI systems are generally not suitable OAuth resource owners. After retrieving an access token from the IDP, you can send it along to Flyte Admin as usual.

Flytekit configuration variables are automatically designed to look up values from relevant environment variables. However, to aid with continuous integration use-cases, Flytekit configuration can also reference other environment variables. 

For instance, if your CI system is not capable of setting custom environment variables like ``FLYTE_CREDENTIALS_CLIENT_SECRET`` but does set the necessary settings under a different variable, you may use ``export FLYTE_CREDENTIALS_CLIENT_SECRET_FROM_ENV_VAR=OTHER_ENV_VARIABLE`` to redirect the lookup. A ``FLYTE_CREDENTIALS_CLIENT_SECRET_FROM_FILE`` redirect is available as well, where the value should be the full path to the file containing the value for the configuration setting, in this case, the client secret. We found this redirect behavior necessary when setting up registration within our own CI pipelines.

The following is a listing of the Flytekit configuration values we set in CI, along with a brief explanation.

* ``FLYTE_CREDENTIALS_CLIENT_ID`` and ``FLYTE_CREDENTIALS_CLIENT_SECRET``
  When using basic authentication, this is the username and password.
* ``export FLYTE_CREDENTIALS_AUTH_MODE=basic``
  This tells the SDK to use basic authentication. If not set, Flytekit will assume you want to use the standard OAuth based three-legged flow.
* ``export FLYTE_CREDENTIALS_AUTHORIZATION_METADATA_KEY=text``
  At Lyft, we set this to conform to this `header config <https://github.com/flyteorg/flyteadmin/blob/eaca2fb0e6018a2e261e9e2da8998906477cadb5/pkg/auth/config/config.go#L53>`_ on the Admin side.
* ``export FLYTE_CREDENTIALS_SCOPE=text``
  When using basic authentication, you'll need to specify a scope to the IDP (instead of ``openid``, which is only for OAuth). Set that here.
* ``export FLYTE_PLATFORM_AUTH=True``
  Set this to force Flytekit to use authentication, even if not required by Admin. This is useful as you're rolling out the requirement.


**********
References
**********

RFCs
======
This collection of RFCs may be helpful to those who wish to investigate the implementation in more depth.

* `OAuth2 RFC 6749 <https://tools.ietf.org/html/rfc6749>`_
* `OAuth Discovery RFC 8414 <https://tools.ietf.org/html/rfc8414>`_
* `PKCE RFC 7636 <https://tools.ietf.org/html/rfc7636>`_
* `JWT RFC 7519 <https://tools.ietf.org/html/rfc7519>`_


