"""
.. _migrating-auth-config:

####################################
Migrating Your Authentication Config
####################################

Using Okta as an example, you would have previously seen something like the following:

Okta Config changes
===================

* An Application (OpenID Connect Web) for Flyte Admin itself (e.g. **0oal5rch46pVhCGF45d6**).
* An Application (OpenID Native app) for Flyte-cli/flytectl (e.g. **0oal62nxuD6OSFSRq5d6**).
  These two applications would be assigned to the relevant users.
* An Application (Web) for Flyte Propeller (e.g. **0abc5rch46pVhCGF9876**).
  This application would either use the default Authorization server, or you would create a new one.

Admin Config Changes
====================

.. code-block:: yaml

    server:
    # ... other settings
    security:
        secure: false
        useAuth: true
        allowCors: true
        allowedOrigins:
        - "*"
        allowedHeaders:
        - "Content-Type"
        oauth:
        baseUrl: https://dev-62129345.okta.com/oauth2/default/
        scopes:
            - profile
            - openid
            - email
        claims:
            iss: https://dev-62129345.okta.com/oauth2/default
            aud: 0oal5rch46pVhCGF45d6
        clientId: 0oal5rch46pVhCGF45d6
        clientSecretFile: "/Users/ytong/etc/secrets/oauth/secret"
        authorizeUrl: "https://dev-62129345.okta.com/oauth2/default/v1/authorize"
        tokenUrl: "https://dev-62129345.okta.com/oauth2/default/v1/token"
        callbackUrl: "http://localhost:8088/callback"
        cookieHashKeyFile: "/Users/ytong/etc/secrets/hashkey/hashkey"
        cookieBlockKeyFile: "/Users/ytong/etc/secrets/blockkey/blockkey"
        redirectUrl: "/api/v1/projects"
        thirdPartyConfig:
            flyteClient:
            clientId: 0oal62nxuD6OSFSRq5d6
            redirectUri: http://localhost:12345/callback

After version `0.13.0 <https://github.com/flyteorg/flyte/tree/v0.13.0>`__ of the platform, you can still use the IdP as
the Authorization Server if you wish. That configuration would now become:

.. code-block:: yaml

    server:
    # ... other settings
    security:
        secure: false
        useAuth: true
        allowCors: true
        allowedOrigins:
        - "*"
        allowedHeaders:
        - "Content-Type"
    auth:
        authorizedUris:
            # This should point at your public http Uri.
            - https://flyte.mycompany.com
            # This will be used by internal services in the same namespace as flyteadmin
            - http://flyteadmin:80
            # This will be used by internal services in the same cluster but different namespaces
            - http://flyteadmin.flyte.svc.cluster.local:80
        userAuth:
            openId:
                # Put the URL of the OpenID Connect provider.
                baseUrl: https://dev-62129345.okta.com/oauth2/default # Okta with a custom Authorization Server
                scopes:
                    - profile
                    - openid
                    - offline_access # Uncomment if OIdC supports issuing refresh tokens.
                # Replace with the client id created for Flyte.
                clientId: 0oal5rch46pVhCGF45d6
        appAuth:
            # External delegates app auth responsibilities to an external authorization server, Internal means Flyte Admin does it itself
            authServerType: External
            thirdPartyConfig:
                flyteClient:
                    clientId: 0oal62nxuD6OSFSRq5d6
                    redirectUri: http://localhost:12345/callback
                    scopes:
                    - all
                    - offline

Specifically,

* The original **oauth** section has been moved two levels higher into its own section and renamed **auth** but enabling/disabling of authentication remains in the old location.
* Secrets by default will now be looked up in **/etc/secrets**. Use the following command to generate them:

  .. prompt:: bash

     flyteadmin secrets init -p /etc/secrets

  This will generate the new cookie hash/block keys, as well as other secrets Admin needs to run the Authorization server.

* The **clientSecretFile** has been moved to **/etc/secrets/oidc_client_secret** so move that there.
* **claims** has been removed, just delete that.
* **authorizeUrl** and **tokenUrl** are no longer necessary.
* The **baseUrl** for the external Authorization Server is now in the **appAuth** section.
* The **thirdPartyConfig** has been moved to **appAuth** as well.
* **redirectUrl** has been defaulted to **/console**. If that's the value you want, then you no longer need this setting.

Propeller Config Changes
========================

You might have a configuration section that looks like this:

.. code-block:: yaml

    admin:
      endpoint: dns:///mycompany.domain.com
      useAuth: true
      clientId: flytepropeller
      clientSecretLocation: /etc/secrets/client_secret
      tokenUrl: https://demo.nuclyde.io/oauth2/token
      scopes:
      - all

This can now be simplified to:

.. code-block:: yaml

    admin:
      endpoint: dns:///mycompany.domain.com
      # If you are using the built-in authorization server, you can delete the following two lines:
      clientId: flytepropeller
      clientSecretLocation: /etc/secrets/client_secret

Specifically,

* **useAuth** is deprecated and will be removed in a future version. Auth requirement will be discovered through an anonymous admin discovery call.
* **tokenUrl** and **scopes** will also be discovered through a metadata call.
* **clientId** and **clientSecretLocation** have defaults that work out of the box with the built-in authorization server (e.g. if you setup Google OpenID Connect).

Flyte-cli Config Changes
========================

You need these two configs set:

.. prompt:: bash

    FLYTE_PLATFORM_HTTP_URL=http://localhost:8088 FLYTE_CREDENTIALS_CLIENT_ID=0oal62nxuD6OSFSRq5d6 flyte-cli ...

**FLYTE_PLATFORM_HTTP_URL** is used because **flyte-cli** uses only gRPC to communicate with Admin. It needs to know the
HTTP port (which Admin hosts on a different port because of limitations of the grpc-gateway library). **flyte-cli** uses
this setting to talk to **/.well-known/oauth-authorization-server** to retrieve information regarding the auth endpoints
. Previously this redirected to the Okta Authorization Server's metadata endpoint. With this change, Admin now hosts its
own (even if still using the external Authorization Service).
"""
