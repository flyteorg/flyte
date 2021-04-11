.. _howto_authentication:

#######################################
How to setup Authentication?
#######################################

Flyte Admin ships with a canonical implementation of OAuth2, integrating seamlessly into an organization's existing identity provider.  At Lyft, we use Okta as the IDP, but if you have issues integrating with another implementation of the OAuth server, please open an issue.

***********************
Components
***********************

While the most obvious interaction with the Flyte control plane is through the web based UI, there are other critical components of Flyte that also need to be considered. These components should be thought of as third-party services even though the Flyte codebase provides them.

Flyte CLI
=========
Principal amongst these is the Flyte CLI. This is the command-line entrypoint to Flyte Admin and is used by both administrators and users more comfortable in the command line, or are running in a headless OS.

The IDP application corresponding to the CLI will need to support PKCE.


Direct Client Access
====================
The gRPC client provided by the Flyte IDL, or direct calls to the HTTP endpoints on Admin from within a running script are ways that we have seen users hit the control plane directly.  We generally discourage this behavior as it leads to a possible self-imposed DOS vector, as they are generally made from within a running workflow itself. For instance, a Flyte task can fetch the definition for a launch plan associated with a completely different workflow, and then launch an execution of it. This is not the correct way to launch one workflow from another but for the time being remains possible.


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

Example Configurations
======================

Google IdP
##########

1. Follow `Google Docs <https://developers.google.com/identity/protocols/oauth2/openid-connect>`__ on how to configure the IdP for OpenIDConnect.

.. note::

  Make sure to create an OAuth2 Client Credential. The `client_id` and `client_secret` will be needed in the following
  steps.

2. Store the `client_secret` in a k8s secrt as follows:

.. prompt:: bash

  kubectl edit secret -n flyte flyte-admin-auth

Add a new key under `data`:

.. code-block:: yaml

  stringData:
    oidc_client_secret: <client_secret> from the previous step

Save and close your editor.

3. Edit FlyteAdmin config to add `client_id` as follows:

.. prompt:: bash

  kubectl deploy -n flyte flyteadmin -o yaml | grep "name: flyte-admin-config"

This will output the name of the config map where the `client_id` need to go.

.. prompt:: bash

  kubectl edit configmap -n flyte <the name of the config map from previous command>

Find `client_id` and replace with the copied `client_id`

.. code-block:: yaml

  clientId: 657465813211-6eog7ek7li5k7i7fvgv2921075063hpe.apps.googleusercontent.com

Find `useAuth` and enable Auth enforcement:

.. code-block:: yaml

  useAuth: true

Save and exit your editor.

4. Restart `flyteadmin` for the changes to take effect:

.. prompt:: bash

  kubectl rollout restart deployment/flyteadmin -n flyte

******
CI
******

If your organization does any automated registration, then you'll need to authenticate with the `basic authentication <https://tools.ietf.org/html/rfc2617>`_ flow (username and password effectively) as CI systems are generally not suitable OAuth resource owners. After retrieving an access token from the IDP, you can send it along to Flyte Admin as usual.

Flytekit configuration variables are automatically designed to look up values from relevant environment variables. To aid with continuous integration use-cases however, Flytekit configuration can also reference other environment variables.  For instance, if your CI system is not capable of setting custom environment variables like ``FLYTE_CREDENTIALS_CLIENT_SECRET`` but does set the necessary settings under a different variable, you may use ``export FLYTE_CREDENTIALS_CLIENT_SECRET_FROM_ENV_VAR=OTHER_ENV_VARIABLE`` to redirect the lookup. A ``FLYTE_CREDENTIALS_CLIENT_SECRET_FROM_FILE`` redirect is available as well, where the value should be the full path to the file containing the value for the configuration setting, in this case, the client secret. We found this redirect behavior necessary when setting up registration within our own CI pipelines.

The following is a listing of the Flytekit configuration values we set in CI, along with a brief explanation where relevant.

* ``FLYTE_CREDENTIALS_CLIENT_ID`` and ``FLYTE_CREDENTIALS_CLIENT_SECRET``
  When using basic authentication, this is the username and password
* ``export FLYTE_CREDENTIALS_AUTH_MODE=basic``
  This tells the SDK to use basic authentication. If not set, Flytekit will assume you want to use the standard OAuth based three-legged flow.
* ``export FLYTE_CREDENTIALS_AUTHORIZATION_METADATA_KEY=text``
  At Lyft, we set this to conform to this `header config <https://github.com/flyteorg/flyteadmin/blob/eaca2fb0e6018a2e261e9e2da8998906477cadb5/pkg/auth/config/config.go#L53>`_ on the Admin side.
* ``export FLYTE_CREDENTIALS_SCOPE=text``
  When using basic authentication, you'll need to specify a scope to the IDP (instead of ``openid``, as that's only for OAuth). Set that here.
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


