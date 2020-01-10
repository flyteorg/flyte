.. _install-authentication

#################
Authentication
#################

Flyte Admin ships with a canonical implementation of OAuth2, integrating seamlessly into an organization's existing identity provider.  At Lyft, we use Okta as the IDP, but if you have issues integrating with another implementation of the OAuth server, please open an issue.

***********************
Components
***********************

While the most obvious interaction with the Flyte control plane is through the web based UI, there are other critical components of Flyte that also need to be considered. These components should be thought of as third-party services even though the Flyte codebase provides them.

Flyte CLI
=========
Principal amongst these is the Flyte CLI. This is the command-line entrypoint to Flyte Admin and is used by both administrators and users more comfortable in the command line, or are running in a headless OS.

The IDP application corresponding to the CLI will need to support PKCE.


IDL Client
==========
The gRPC client provided by the Flyte IDL, or direct calls to the HTTP endpoints on Admin from within a running script, comprise the primary client calls that Flyte Admin handles. These calls are often direct

In both cases they are generally made from within a running workflow itself. For instance, a Flyte task can fetch the definition for a launch plan associated with a completely different workflow, and then wait for its execution 


******
Flow
******

Flyte Admin authentication is implemented using the authorization code flow.

Flyte UI Flow
https://swimlanes.io/d/OmV4ybCkx
https://static.swimlanes.io/fd751baba1152f38f40744fc06a67b1f.png

Flyte CLI Flow
https://swimlanes.io/d/q64OxuoxT
https://static.swimlanes.io/3f0fa54e2cc52d4a07666be1cda3ccd3.png

*************
Configuration
*************

IDP Configuration
=================
Flyte Admin will require that the application in your identity provider be configured without PKCE and with a client secret. It should also be configured with a refresh token.

Flyte Admin Configuration
=========================
.. TODO : 
Please refer to the inline documentation on the ``Config`` object in the ``auth`` package for a discussion on the settings required.



CI Integration
==============
If your organization does any automated registration, then 

**********
References
**********

RFCs
This collection of RFCs may be helpful to those who wish to investigate the implementation in more depth.



