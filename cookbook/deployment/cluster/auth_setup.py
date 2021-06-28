"""
Authentication in Flyte
-----------------------

Flyte ships with a canonical implementation of OpenIDConnect client and OAuth2 Server, integrating seamlessly into an
organization's existing identity provider.

.. _auth-overview:

########
Overview
########

The Flyte system consists of multiple components. Securing communication between each components is crucial to ensure
the security of the overall system.

In abstract, Flyte supports OAuth2 and OpenId Connect (built on top of OAuth2) to secure the various connections:

* :ref:`OpenId Connect <auth-openid-appendix>`: Used to secure user's authentication to flyteadmin service.
* :ref:`OAuth2 <auth-oauth2-appendix>`: Used to secure communication between clients (i.e. flyte-cli, flytectl and
  flytepropeller) and flyteadmin service.

Identity Providers Support
==========================

Support for these protocols varies per IdP, checkout the following table to understand the available support level for
your IdP.

+----------------------+--------+-------------+---------------------+----------+-------+----------+--------+
| Feature              | Okta   | Google free | GC Identity Service | Azure AD | Auth0 | KeyCloak | Github |
+======================+========+=============+=====================+==========+=======+==========+========+
| OpenIdConnect        |   Yes  |     Yes     |          Yes        |    Yes   |  Yes  |    Yes   |   No   |
+----------------------+--------+-------------+---------------------+----------+-------+----------+--------+
| Custom Relying Party |   Yes  |      No     |          Yes        |    Yes   |   ?   |    Yes   |   No   |
+----------------------+--------+-------------+---------------------+----------+-------+----------+--------+

.. _auth-setup:

####################
Authentication Setup
####################

*****************
IdP Configuration
*****************
Flyte Admin requires that the application in your identity provider be configured as a web client (i.e. with a client secret). We recommend allowing the application to be issued a refresh token to avoid interrupting the user's flow by frequently redirecting to the IdP.

****************************
Example Flyte Configurations
****************************

Below are some canonical examples of how to set up some of the common IdPs to secure your Fyte services. OpenID Connect enables users to authenticate, in the
browser, with an existing IdP. Flyte also allows connecting to an external OAuth2 Authorization Server to allow centrally managed third party app access.

OpenID Connect
===============

OpenID Connect allows users to authenticate to Flyte in their browser using a familiar authentication provider (perhaps an organization-wide configured IdP).
Flyte supports connecting with external OIdC providers. Here are some examples for how to set these up:

.. tabs::

    .. tab:: Google

        Follow `Google Docs <https://developers.google.com/identity/protocols/oauth2/openid-connect>`__ on how to configure the IdP for OpenIDConnect.

        .. note::

          Make sure to create an OAuth2 Client Credential. The `client_id` and `client_secret` will be needed in the following
          steps.

    .. tab:: Okta

        Okta supports OpenID Connect protocol and the creation of custom OAuth2 Authorization Servers, allowing it to act as both the user and apps IdP.
        It offers more detailed control on access policies, user consent, and app management.

        1. If you don't already have an Okta account, sign up for one `here <https://developer.okta.com/signup/>`__.
        2. Create an app (choose Web for the platform) and OpenID Connect for the sign-on method.
        3. Add Login redirect URIs (e.g. http://localhost:30081/callback for sandbox or ``https://<your deployment url>/callback``)
        4. *Optional*: Add logout redirect URIs (e.g. http://localhost:30081/logout for sandbox)
        5. Write down the Client ID and Client Secret

    .. tab:: KeyCloak

        `KeyCloak <https://www.keycloak.org/>`__ is an open source solution for authentication, it supports both OpenID Connect and OAuth2 protocols (among others).
        KeyCloak can be configured to be both the OpenID Connect and OAuth2 Authorization Server provider for Flyte.

Apply Configuration
^^^^^^^^^^^^^^^^^^^

#. Store the `client_secret` in a k8s secret as follows:

   .. prompt:: bash

     kubectl edit secret -n flyte flyte-admin-auth

   Add a new key under `stringData`:

   .. code-block:: yaml

     stringData:
       oidc_client_secret: <client_secret from the previous step>
     data:
       ...

   Save and close your editor.

#. Edit FlyteAdmin config to add `client_id` and configure auth as follows:

   .. prompt:: bash

      kubectl edit configmap -n flyte flyte-admin-config

   Follow the inline comments to make the necessary changes:

   .. code-block:: yaml

      server:
        ...
        security:
          secure: false
          # 1. Enable Auth by turning useAuth to true
          useAuth: true
          ...
      auth:
        userAuth:
          openId:
            # 2. Put the URL of the OpenID Connect provider.
            #    baseUrl: https://accounts.google.com # Uncomment for Google
            baseUrl: https://dev-14186422.okta.com/oauth2/default # Okta with a custom Authorization Server
            scopes:
              - profile
              - openid
              # - offline_access # Uncomment if OIdC supports issuing refresh tokens.
            # 3. Replace with the client ID created for Flyte.
            clientId: 0oakkheteNjCMERst5d6

   Save and exit your editor.

#. Restart `flyteadmin` for the changes to take effect:

   .. prompt:: bash

      kubectl rollout restart deployment/flyteadmin -n flyte

OAuth2 Authorization Server
===========================

An OAuth2 Authorization Server allows external clients to request to authenticate and act on behalf of users (or as their own identities). Having 
an OAuth2 Authorization Server enables Flyte administrators control over which apps can be installed and what scopes they are allowed to request or be granted (i.e. what privileges can they assume).

Flyte comes with a built-in authorization server that can be statically configured with a set of clients to request and act on behalf of the user.
The default clients are defined `here <https://github.com/flyteorg/flyteadmin/pull/168/files#diff-1267ff8bd9146e1c0ff22a9e9d53cfc56d71c1d47fed9905f95ed4bddf930f8eR74-R100>`__ 
and the corresponding section can be modified through configs.

To set up an external OAuth2 Authorization Server, please follow the instructions below:

.. tabs::

    .. tab:: Okta

        1. Under security -> API, click `Add Authorization Server`. Set the audience to the public URL of flyte admin (e.g. ``<https://flyte.mycompany.io/>``).
        2. Under `Access Policies`, click `Add New Access Policy` and walk through the wizard to allow access to the authorization server.
        3. Under `Scopes`, click `Add Scope`. Set the name to `all` (required) and check `Require user consent for this scope` (recommended).
        4. Create 2 apps (for fltyectl and flytepropeller) to enable these clients to communicate with the service.
           Flytectl should be created as a `native client`.
           FlytePropeller should be created as an `OAuth Service` and note the client ID and client Secrets provided.

    .. tab:: KeyCloak

        `KeyCloak <https://www.keycloak.org/>`__ is an open source solution for authentication, it supports both OpenID Connect and OAuth2 protocols (among others).
        KeyCloak can be configured to be both the OpenID Connect and OAuth2 Authorization Server provider for flyte.

Apply Configuration
^^^^^^^^^^^^^^^^^^^

#. It is possible to direct Flyte admin to use an external authorization server. To do so, edit the same config map once
   more and follow these changes:

   .. code-block:: yaml

        auth:
            appAuth:
                # 1. Choose External if you will use an external Authorization Server (e.g. a Custom Authorization server in Okta)
                #    Choose Self (or omit the value) to use Flyte Admin's internal (albeit limited) Authorization Server.
                authServerType: External

                # 2. Optional: Set external auth server baseUrl if different from OpenId baseUrl.
                externalAuthServer:
                    baseUrl: https://dev-14186422.okta.com/oauth2/auskngnn7uBViQq6b5d6
            thirdPartyConfig:
                flyteClient:
                    # 3. Replace with a new Native Client ID provisioned in the custom authorization server
                    clientId: flytectl

                    redirectUri: https://localhost:53593/callback

                    # 4. "all" is a required scope and must be configured in the custom authorization server
                    scopes:
                    - offline
                    - all
            userAuth:
                openId:
                    baseUrl: https://dev-14186422.okta.com/oauth2/auskngnn7uBViQq6b5d6 # Okta with a custom Authorization Server
                    scopes:
                    - profile
                    - openid
                    # - offline_access # Uncomment if OIdC supports issuing refresh tokens.
                    clientId: 0oakkheteNjCMERst5d6

#. Store flyte propeller's `client_secret` in a k8s secret as follows:

   .. prompt:: bash

      kubectl edit secret -n flyte flyte-propeller-auth

   Add a new key under `stringData`:

   .. code-block:: yaml

      stringData:
        client_secret: <client_secret> from the previous step
      data:
        ...

   Save and close your editor.

#. Edit FlytePropeller config to add `client_id` and configure auth as follows:

   .. prompt:: bash

      kubectl edit configmap -n flyte flyte-propeller-config

   Follow the inline comments to make the necessary changes:

   .. code-block:: yaml

      admin:
          # 1. Replace with the client_id provided by the OAuth2 Authorization Server above.
          clientId: flytepropeller

   Close the editor

#. Restart `flytepropeller` for the changes to take effect:

   .. prompt:: bash

      kubectl rollout restart deployment/flytepropeller -n flyte

Continuous Integration - CI
===========================

If your organization does any automated registration, then you'll need to authenticate with the `client credentials <https://datatracker.ietf.org/doc/html/rfc6749#section-4.4>`_ flow. After retrieving an access token from the IDP, you can send it along to Flyte Admin as usual.

.. tabs::

    .. tab:: Flytectl

        Flytectl's `config.yaml <https://docs.flyte.org/projects/flytectl/en/stable/#configure>`_ can be
        configured to use either PKCE (`Proof key for code exchange <https://datatracker.ietf.org/doc/html/rfc7636>`_)
        or Client Credentials (`Client Credentials <https://datatracker.ietf.org/doc/html/rfc6749#section-4.4>`_) flows.

        Update ``config.yaml`` as follows:

        .. code-block:: yaml

            admin:
                # Update with the admin's endpoint. You must keep the 3 forward-slashe after dns:
                endpoint:: dns:///<admin's url>

                # Update auth type to `Pkce` or `ClientSecret`
                authType: Pkce

                # Set to the clientId (will be used for both Pkce and ClientSecret flows)
                # Leave empty to use the value discovered through flyteadmins Auth discovery endpoint.
                clientId: <Id>

                # Set to the location where the client secret is mounted.
                # Only needed/used for `ClientSecret` flow.
                clientSecretLocation: </some/path/to/key>

                # If required, set the scopes needed here. Otherwise, flytectl will discover scopes required for OpenID
                # Connect through Admin's Auth discovery endpoint.
                # scopes: [ "scope1", "scope2" ]

        To read further about the available config options, please
        `visit here <https://github.com/flyteorg/flyteidl/blob/master/clients/go/admin/config.go#L37-L64>`_

    .. tab:: Flytekit / Flyte-cli

        Flytekit configuration variables are automatically designed to look up values from relevant environment variables.
        However, to aid with continuous integration use-cases, Flytekit configuration can also reference other environment
        variables.

        For instance, if your CI system is not capable of setting custom environment variables like
        ``FLYTE_CREDENTIALS_CLIENT_SECRET`` but does set the necessary settings under a different variable, you may use
        ``export FLYTE_CREDENTIALS_CLIENT_SECRET_FROM_ENV_VAR=OTHER_ENV_VARIABLE`` to redirect the lookup. A
        ``FLYTE_CREDENTIALS_CLIENT_SECRET_FROM_FILE`` redirect is available as well, where the value should be the full
        path to the file containing the value for the configuration setting, in this case, the client secret. We found
        this redirect behavior necessary when setting up registration within our own CI pipelines.

        The following is a listing of the Flytekit configuration values we set in CI, along with a brief explanation.

        .. code:: bash

            # When using OAuth2 service auth, this is the username and password.
            export FLYTE_CREDENTIALS_CLIENT_ID=<client_id>
            export FLYTE_CREDENTIALS_CLIENT_SECRET=<client_secret>

            # This tells the SDK to use basic authentication. If not set, Flytekit will assume you want to use the
            # standard OAuth based three-legged flow.
            export FLYTE_CREDENTIALS_AUTH_MODE=basic

            # This value should be set to conform to this
            # `header config <https://github.com/flyteorg/flyteadmin/blob/12d6aa0a419ccec81b4c8289fd172e70a2ded525/auth/config/config.go#L124-L128>`_
            # on the Admin side.
            export FLYTE_CREDENTIALS_AUTHORIZATION_METADATA_KEY=<header name>

            # When using basic authentication, you'll need to specify a scope to the IDP (instead of ``openid``, which is
            # only for OAuth). Set that here.
            export FLYTE_CREDENTIALS_OAUTH_SCOPES=<idp defined scopes>

            # Set this to force Flytekit to use authentication, even if not required by Admin. This is useful as you're
            # rolling out the requirement.
            export FLYTE_PLATFORM_AUTH=True

.. _auth-references:

##########
References
##########

This collection of RFCs may be helpful to those who wish to investigate the implementation in more depth.

* `OAuth2 RFC 6749 <https://tools.ietf.org/html/rfc6749>`_
* `OAuth Discovery RFC 8414 <https://tools.ietf.org/html/rfc8414>`_
* `PKCE RFC 7636 <https://tools.ietf.org/html/rfc7636>`_
* `JWT RFC 7519 <https://tools.ietf.org/html/rfc7519>`_

"""
