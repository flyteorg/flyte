.. _howto_authentication_migrate:

######################################################
How to migrate your authentication config (pre 0.13.0)
######################################################

Using Okta as an example, previously, you would've had something like the following.

On the Okta side, you'd have:

* An Application (OpenID Connect Web) for Flyte Admin itself (e.g. **0oal5rch46pVhCGF45d6**).
* An Application (OpenID Native app) for flyte-cli/flytectl (e.g. **0oal62nxuD6OSFSRq5d6**).
* These two applications would be assigned to the relevant users.
* An Application (Web) for Flyte Propeller (e.g. **0abc5rch46pVhCGF9876**).
* These applications would either use the default Authorization server, or you'd create a new one.

On the Admin side, you'd have the following configuration

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

From the flyte-cli side, you needed these two settings

.. code-block:: bash

    FLYTE_PLATFORM_HTTP_URL=http://localhost:8088 FLYTE_CREDENTIALS_CLIENT_ID=0oal62nxuD6OSFSRq5d6 flyte-cli ...

**FLYTE_PLATFORM_HTTP_URL** is used because **flyte-cli** uses only gRPC to communicate with Admin. It needs to know the HTTP port (which Admin hosts on a different port because of limitations of the 
grpc-gateway library). **flyte-cli** uses this setting to talk to **/.well-known/oauth-authorization-server** to retrieve information regarding the auth endpoints.  Previously this redirected to the
Okta Authorization Server's metadata endpoint. With this change, Admin now hosts its own (even if still using the external Authorization Service).

After version `0.13.0 <https://github.com/flyteorg/flyte/tree/v0.13.0>`__ of the platfform, you can still use the IDP as the Authorization Server if you so choose. That configuration would become:

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
        # This should point at your public http Uri.
        httpPublicUri: http://localhost:8088/
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

* The original **oauth** section has been moved two levels higher into its own section and renamed **auth** but enabling/disabling of authentication remains in the old place.
* Secrets by default will now be looked up in **/etc/secrets**. Use the following command to generate them.

.. code-block:: bash

    flyteadmin secrets init -p /etc/secrets

This will generate the new cookie hash/block keys, as well as other secrets Admin needs to run the Authorization server.

* The **clientSecretFile** has been moved to **/etc/secrets/oidc_client_secret** so move that there.
* **claims** has been removed, just delete that.
* **authorizeUrl** and **tokenUrl** are no longer necessary.
* The **baseUrl** for the external Authorization Server is now in the **appAuth** section.
* The **thirdPartyConfig** has been moved to **appAuth** as well.
* **redirectUrl** has been defaulted to **/console**. If that's the value you want, then you no longer need this setting.

