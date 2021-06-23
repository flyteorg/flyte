"""
Authentication Appendix
------------------------

.. _auth-openid-appendix:

##############
OpenID Connect
##############

Flyte supports OpenID Connect. A defacto standard for user authentication. After configuring OpenID Connect, users
accessing flyte console or flytectl (or other 3rd party apps) will be prompted to authenticate using the configured
provider.

.. mermaid::
    :alt: Flyte UI Swimlane

    sequenceDiagram
    %%{config: { 'fontFamily': 'Menlo', 'fontSize': 10, 'fontWeight': 100} }%%
        autonumber
        User->>+Browser: /home
        Browser->>+Console: /home
        Console->>-Browser: 302 /login
        Browser->>+Admin: /login
        Admin->>-Browser: Idp.com/oidc
        Browser->>+Idp: Idp.com/oidc
        Idp->>-Browser: 302 /login
        Browser->>-User: Enter user/pass
        User->>+Browser: login
        Browser->>+Idp: Submit username/pass
        Idp->>-Browser: admin/?authCode=<abc>
        Browser->>+Admin: admin/authCode=<abc>
        Admin->>+Idp: Exchange Tokens
        Idp->>-Admin: idt, at, rt
        Admin->>+Browser: Write Cookies & Redirect to /console
        Browser->>+Console: /home
        Browser->>-User: Render /home

.. _auth-oauth2-appendix:

########
OAuth2
########

Flyte supports OAuth2 to control access to 3rd party and native apps. FlyteAdmin comes with a built in Authorization
Server that can perform 3-legged and 2-legged OAuth2 flows. It also supports delegating these responsibilities to an
external Authorization Server.

Service Authentication using OAuth2
===================================

Propeller (and potentially other non-user facing services) can also authenticate using client_credentials to the Idp and
be granted an access_token valid to be used with admin and other backend services.

.. tabs::

    .. tab:: FlyteAdmin's builtin Authorization Server

        .. mermaid::
            :alt: Service Authentication Swimlane

            sequenceDiagram
                Propeller->>+Admin: /token?client_creds&scope=https://admin/
                Admin->>-Propeller: access_token
                Propeller->>+Admin: /list_projects?token=access_token

    .. tab:: External Authorization Server

        .. mermaid::
            :alt: Service Authentication Swimlane

            sequenceDiagram
                Propeller->>+External Authorization Server: /token?client_creds&scope=https://admin/
                External Authorization Server->>-Propeller: access_token
                Propeller->>+Admin: /list_projects?token=access_token

User Authentication in other clients (e.g. Cli) using OAuth2-Pkce
==================================================================

Users accessing backend services through Cli should be able to use OAuth2-Pkce flow to authenticate (in a browser) to the Idp and be issued
an access_token valid to communicate with the intended backend service on behalf of the user.

.. tabs::

    .. tab:: FlyteAdmin's builtin Authorization Server

        .. mermaid::
            :alt: CLI Authentication with Admin's own Authorization Server

            sequenceDiagram
            %%{config: { 'fontFamily': 'Menlo', 'fontSize': 10, 'fontWeight': 100} }%%
                autonumber
                User->>+Cli: flytectl list-projects
                Cli->>+Admin: admin/client-config
                Admin->>-Cli: Client_id=<abc>, ...
                Cli->>+Browser: /oauth2/authorize?pkce&code_challenge,client_id,scope
                Browser->>+Admin: /oauth2/authorize?pkce...
                Admin->>-Browser: 302 idp.com/login
                Note over Browser,Admin: The prior OpenID Connect flow
                Browser->>+Admin: admin/logged_in
                Note over Browser,Admin: Potentially show custom consent screen
                Admin->>-Browser: localhost/?authCode=<abc>
                Browser->>+Cli: localhost/authCode=<abc>
                Cli->>+Admin: /token?code,code_verifier
                Admin->>-Cli: access_token
                Cli->>+Admin: /projects/ + access_token
                Admin->>-Cli: project1, project2

    .. tab:: External Authorization Server

        .. mermaid::
            :alt: CLI Authentication with an external Authorization Server

            sequenceDiagram
            %%{config: { 'fontFamily': 'Menlo', 'fontSize': 10, 'fontWeight': 100} }%%
                autonumber
                User->>+Cli: flytectl list-projects
                Cli->>+Admin: admin/client-config
                Admin->>-Cli: Client_id=<abc>, ...
                Cli->>+Browser: /oauth2/authorize?pkce&code_challenge,client_id,scope
                Browser->>+ExternalIdp: /oauth2/authorize?pkce...
                ExternalIdp->>-Browser: 302 idp.com/login
                Note over Browser,ExternalIdp: The prior OpenID Connect flow
                Browser->>+ExternalIdp: /logged_in
                Note over Browser,ExternalIdp: Potentially show custom consent screen
                ExternalIdp->>-Browser: localhost/?authCode=<abc>
                Browser->>+Cli: localhost/authCode=<abc>
                Cli->>+ExternalIdp: /token?code,code_verifier
                ExternalIdp->>-Cli: access_token
                Cli->>+Admin: /projects/ + access_token
                Admin->>-Cli: project1, project2
"""
