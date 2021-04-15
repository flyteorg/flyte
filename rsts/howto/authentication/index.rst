.. _howto_authentication:

#######################
Authentication in Flyte
#######################

Flyte ships with a canonical implementation of OpenIDConnect client and OAuth2 Server, integrating seamlessly into an organization's existing identity provider. 

.. toctree::
   :maxdepth: 1
   :caption: Setting up Flyte Authentication
   :name: howtoprovguidestoc

   setup
   migration

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


