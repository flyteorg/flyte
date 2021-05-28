"""
Authentication in Flyte
-----------------------

Flyte ships with a canonical implementation of OpenIDConnect client and OAuth2 Server, integrating seamlessly into an organization's existing identity provider.

This section includes:

- :ref:`Overview <auth-overview>`
- :ref:`Authentication Setup <auth-setup>`
- :ref:`Migrating Your Authentication Config <migrating-auth-config>`
- :ref:`References <auth-references>`

.. _auth-overview:

########
Overview
########

Flyte system consists of multiple components. For the purposes of this document, let's categorize them into server-side and client-side components:

- **Admin**: A server-side control plane component accessible from Console, cli and other backends.
- **Catalog**: A server-side control plane component accessible from Console, cli and other backends.
- **Console**: A client-side single page react app.
- **flyte-cli**: A python-based client-side command line interface that interacts with Admin and Catalog.
- **flytectl**: A go-based client-side command line interface that interacts with Admin and Catalog.
- **Propeller**: A server-side data plane component that interacts with both Admin and Catalog services.

**************
OpenID Connect
**************

Flyte supports OpenID Connect. A defacto standard for user authentication. After configuring OpenID Connect, users accessing flyte console or flytectl 
(or other 3rd party apps) will be prompted to authenticate using the configured provider.

.. image:: https://mermaid.ink/img/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4lJXtjb25maWc6IHsgJ2ZvbnRGYW1pbHknOiAnTWVubG8nLCAnZm9udFNpemUnOiAxMCwgJ2ZvbnRXZWlnaHQnOiAxMDB9IH0lJVxuICAgIGF1dG9udW1iZXJcbiAgICBVc2VyLT4-K0Jyb3dzZXI6IC9ob21lXG4gICAgQnJvd3Nlci0-PitDb25zb2xlOiAvaG9tZVxuICAgIENvbnNvbGUtPj4tQnJvd3NlcjogMzAyIC9sb2dpblxuICAgIEJyb3dzZXItPj4rQWRtaW46IC9sb2dpblxuICAgIEFkbWluLT4-LUJyb3dzZXI6IElkcC5jb20vb2lkY1xuICAgIEJyb3dzZXItPj4rSWRwOiBJZHAuY29tL29pZGNcbiAgICBJZHAtPj4tQnJvd3NlcjogMzAyIC9sb2dpblxuICAgIEJyb3dzZXItPj4tVXNlcjogRW50ZXIgdXNlci9wYXNzXG4gICAgVXNlci0-PitCcm93c2VyOiBsb2dpblxuICAgIEJyb3dzZXItPj4rSWRwOiBTdWJtaXQgdXNlcm5hbWUvcGFzc1xuICAgIElkcC0-Pi1Ccm93c2VyOiBhZG1pbi8_YXV0aENvZGU9PGFiYz5cbiAgICBCcm93c2VyLT4-K0FkbWluOiBhZG1pbi9hdXRoQ29kZT08YWJjPlxuICAgIEFkbWluLT4-K0lkcDogRXhjaGFuZ2UgVG9rZW5zXG4gICAgSWRwLT4-LUFkbWluOiBpZHQsIGF0LCBydFxuICAgIEFkbWluLT4-K0Jyb3dzZXI6IFdyaXRlIENvb2tpZXMgJiBSZWRpcmVjdCB0byAvY29uc29sZVxuICAgIEJyb3dzZXItPj4rQ29uc29sZTogL2hvbWVcbiAgICBCcm93c2VyLT4-LVVzZXI6IFJlbmRlciAvaG9tZVxuIiwibWVybWFpZCI6eyJ0aGVtZSI6Im5ldXRyYWwifSwidXBkYXRlRWRpdG9yIjpmYWxzZX0
   :target: https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4lJXtjb25maWc6IHsgJ2ZvbnRGYW1pbHknOiAnTWVubG8nLCAnZm9udFNpemUnOiAxMCwgJ2ZvbnRXZWlnaHQnOiAxMDB9IH0lJVxuICAgIGF1dG9udW1iZXJcbiAgICBVc2VyLT4-K0Jyb3dzZXI6IC9ob21lXG4gICAgQnJvd3Nlci0-PitDb25zb2xlOiAvaG9tZVxuICAgIENvbnNvbGUtPj4tQnJvd3NlcjogMzAyIC9sb2dpblxuICAgIEJyb3dzZXItPj4rQWRtaW46IC9sb2dpblxuICAgIEFkbWluLT4-LUJyb3dzZXI6IElkcC5jb20vb2lkY1xuICAgIEJyb3dzZXItPj4rSWRwOiBJZHAuY29tL29pZGNcbiAgICBJZHAtPj4tQnJvd3NlcjogMzAyIC9sb2dpblxuICAgIEJyb3dzZXItPj4tVXNlcjogRW50ZXIgdXNlci9wYXNzXG4gICAgVXNlci0-PitCcm93c2VyOiBsb2dpblxuICAgIEJyb3dzZXItPj4rSWRwOiBTdWJtaXQgdXNlcm5hbWUvcGFzc1xuICAgIElkcC0-Pi1Ccm93c2VyOiBhZG1pbi8_YXV0aENvZGU9PGFiYz5cbiAgICBCcm93c2VyLT4-K0FkbWluOiBhZG1pbi9hdXRoQ29kZT08YWJjPlxuICAgIEFkbWluLT4-K0lkcDogRXhjaGFuZ2UgVG9rZW5zXG4gICAgSWRwLT4-LUFkbWluOiBpZHQsIGF0LCBydFxuICAgIEFkbWluLT4-K0Jyb3dzZXI6IFdyaXRlIENvb2tpZXMgJiBSZWRpcmVjdCB0byAvY29uc29sZVxuICAgIEJyb3dzZXItPj4rQ29uc29sZTogL2hvbWVcbiAgICBCcm93c2VyLT4-LVVzZXI6IFJlbmRlciAvaG9tZVxuIiwibWVybWFpZCI6eyJ0aGVtZSI6Im5ldXRyYWwifSwidXBkYXRlRWRpdG9yIjpmYWxzZX0
   :width: 600
   :alt: Flyte UI Swimlane

******
OAuth2
******

Flyte supports OAuth2 to control access to 3rd party and native apps. FlyteAdmin comes with a built in Authorization Server that can perform 3-legged
and 2-legged OAuth2 flows. It also supports delegating these responsibilities to an external Authorization Server.

Service Authentication using OAuth2
===================================

Propeller (and potentially other non-user facing services) can also authenticate using client_credentials to the Idp and be granted an
access_token valid to be used with admin and other backend services.

Using FlyteAdmin's builtin Authorization Server:

.. image:: https://mermaid.ink/img/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4gICAgUHJvcGVsbGVyLT4-K0FkbWluOiAvdG9rZW4_Y2xpZW50X2NyZWRzJnNjb3BlPWh0dHBzOi8vYWRtaW4vXG4gICAgQWRtaW4tPj4tUHJvcGVsbGVyOiBhY2Nlc3NfdG9rZW5cbiAgICBQcm9wZWxsZXItPj4rQWRtaW46IC9saXN0X3Byb2plY3RzP3Rva2VuPWFjY2Vzc190b2tlbiIsIm1lcm1haWQiOnsidGhlbWUiOiJuZXV0cmFsIn0sInVwZGF0ZUVkaXRvciI6ZmFsc2V9
   :target: https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4gICAgUHJvcGVsbGVyLT4-K0FkbWluOiAvdG9rZW4_Y2xpZW50X2NyZWRzJnNjb3BlPWh0dHBzOi8vYWRtaW4vXG4gICAgQWRtaW4tPj4tUHJvcGVsbGVyOiBhY2Nlc3NfdG9rZW5cbiAgICBQcm9wZWxsZXItPj4rQWRtaW46IC9saXN0X3Byb2plY3RzP3Rva2VuPWFjY2Vzc190b2tlbiIsIm1lcm1haWQiOnsidGhlbWUiOiJuZXV0cmFsIn0sInVwZGF0ZUVkaXRvciI6ZmFsc2V9
   :width: 600
   :alt: Service Authentication Swimlane

Using an External Authorization Server:

.. image:: https://mermaid.ink/img/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4gICAgUHJvcGVsbGVyLT4-K0V4dGVybmFsIEF1dGhvcml6YXRpb24gU2VydmVyOiAvdG9rZW4_Y2xpZW50X2NyZWRzJnNjb3BlPWh0dHBzOi8vYWRtaW4vXG4gICAgRXh0ZXJuYWwgQXV0aG9yaXphdGlvbiBTZXJ2ZXItPj4tUHJvcGVsbGVyOiBhY2Nlc3NfdG9rZW5cbiAgICBQcm9wZWxsZXItPj4rQWRtaW46IC9saXN0X3Byb2plY3RzP3Rva2VuPWFjY2Vzc190b2tlbiIsIm1lcm1haWQiOnsidGhlbWUiOiJuZXV0cmFsIn0sInVwZGF0ZUVkaXRvciI6ZmFsc2V9
   :target: https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4gICAgUHJvcGVsbGVyLT4-K0V4dGVybmFsIEF1dGhvcml6YXRpb24gU2VydmVyOiAvdG9rZW4_Y2xpZW50X2NyZWRzJnNjb3BlPWh0dHBzOi8vYWRtaW4vXG4gICAgRXh0ZXJuYWwgQXV0aG9yaXphdGlvbiBTZXJ2ZXItPj4tUHJvcGVsbGVyOiBhY2Nlc3NfdG9rZW5cbiAgICBQcm9wZWxsZXItPj4rQWRtaW46IC9saXN0X3Byb2plY3RzP3Rva2VuPWFjY2Vzc190b2tlbiIsIm1lcm1haWQiOnsidGhlbWUiOiJuZXV0cmFsIn0sInVwZGF0ZUVkaXRvciI6ZmFsc2V9
   :width: 600
   :alt: Service Authentication Swimlane

User Authentication in other clients (e.g. Cli) using OAuth2-Pkce
==================================================================

Users accessing backend services through Cli should be able to use OAuth2-Pkce flow to authenticate (in a browser) to the Idp and be issued
an access_token valid to communicate with the intended backend service on behalf of the user.

Using FlyteAdmin's builtin Authorization Server:

.. image:: https://mermaid.ink/img/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4lJXtjb25maWc6IHsgJ2ZvbnRGYW1pbHknOiAnTWVubG8nLCAnZm9udFNpemUnOiAxMCwgJ2ZvbnRXZWlnaHQnOiAxMDB9IH0lJVxuICAgIGF1dG9udW1iZXJcbiAgICBVc2VyLT4-K0NsaTogZmx5dGVjdGwgbGlzdC1wcm9qZWN0c1xuICAgIENsaS0-PitBZG1pbjogYWRtaW4vY2xpZW50LWNvbmZpZ1xuICAgIEFkbWluLT4-LUNsaTogQ2xpZW50X2lkPTxhYmM-LCAuLi5cbiAgICBDbGktPj4rQnJvd3NlcjogL29hdXRoMi9hdXRob3JpemU_cGtjZSZjb2RlX2NoYWxsZW5nZSxjbGllbnRfaWQsc2NvcGVcbiAgICBCcm93c2VyLT4-K0FkbWluOiAvb2F1dGgyL2F1dGhvcml6ZT9wa2NlLi4uXG4gICAgQWRtaW4tPj4tQnJvd3NlcjogMzAyIGlkcC5jb20vbG9naW5cbiAgICBOb3RlIG92ZXIgQnJvd3NlcixBZG1pbjogVGhlIHByaW9yIE9wZW5JRCBDb25uZWN0IGZsb3dcbiAgICBCcm93c2VyLT4-K0FkbWluOiBhZG1pbi9sb2dnZWRfaW5cbiAgICBOb3RlIG92ZXIgQnJvd3NlcixBZG1pbjogUG90ZW50aWFsbHkgc2hvdyBjdXN0b20gY29uc2VudCBzY3JlZW5cbiAgICBBZG1pbi0-Pi1Ccm93c2VyOiBsb2NhbGhvc3QvP2F1dGhDb2RlPTxhYmM-XG4gICAgQnJvd3Nlci0-PitDbGk6IGxvY2FsaG9zdC9hdXRoQ29kZT08YWJjPlxuICAgIENsaS0-PitBZG1pbjogL3Rva2VuP2NvZGUsY29kZV92ZXJpZmllclxuICAgIEFkbWluLT4-LUNsaTogYWNjZXNzX3Rva2VuXG4gICAgQ2xpLT4-K0FkbWluOiAvcHJvamVjdHMvICsgYWNjZXNzX3Rva2VuXG4gICAgQWRtaW4tPj4tQ2xpOiBwcm9qZWN0MSwgcHJvamVjdDJcbiIsIm1lcm1haWQiOnsidGhlbWUiOiJuZXV0cmFsIn0sInVwZGF0ZUVkaXRvciI6ZmFsc2V9
   :target: https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4lJXtjb25maWc6IHsgJ2ZvbnRGYW1pbHknOiAnTWVubG8nLCAnZm9udFNpemUnOiAxMCwgJ2ZvbnRXZWlnaHQnOiAxMDB9IH0lJVxuICAgIGF1dG9udW1iZXJcbiAgICBVc2VyLT4-K0NsaTogZmx5dGVjdGwgbGlzdC1wcm9qZWN0c1xuICAgIENsaS0-PitBZG1pbjogYWRtaW4vY2xpZW50LWNvbmZpZ1xuICAgIEFkbWluLT4-LUNsaTogQ2xpZW50X2lkPTxhYmM-LCAuLi5cbiAgICBDbGktPj4rQnJvd3NlcjogL29hdXRoMi9hdXRob3JpemU_cGtjZSZjb2RlX2NoYWxsZW5nZSxjbGllbnRfaWQsc2NvcGVcbiAgICBCcm93c2VyLT4-K0FkbWluOiAvb2F1dGgyL2F1dGhvcml6ZT9wa2NlLi4uXG4gICAgQWRtaW4tPj4tQnJvd3NlcjogMzAyIGlkcC5jb20vbG9naW5cbiAgICBOb3RlIG92ZXIgQnJvd3NlcixBZG1pbjogVGhlIHByaW9yIE9wZW5JRCBDb25uZWN0IGZsb3dcbiAgICBCcm93c2VyLT4-K0FkbWluOiBhZG1pbi9sb2dnZWRfaW5cbiAgICBOb3RlIG92ZXIgQnJvd3NlcixBZG1pbjogUG90ZW50aWFsbHkgc2hvdyBjdXN0b20gY29uc2VudCBzY3JlZW5cbiAgICBBZG1pbi0-Pi1Ccm93c2VyOiBsb2NhbGhvc3QvP2F1dGhDb2RlPTxhYmM-XG4gICAgQnJvd3Nlci0-PitDbGk6IGxvY2FsaG9zdC9hdXRoQ29kZT08YWJjPlxuICAgIENsaS0-PitBZG1pbjogL3Rva2VuP2NvZGUsY29kZV92ZXJpZmllclxuICAgIEFkbWluLT4-LUNsaTogYWNjZXNzX3Rva2VuXG4gICAgQ2xpLT4-K0FkbWluOiAvcHJvamVjdHMvICsgYWNjZXNzX3Rva2VuXG4gICAgQWRtaW4tPj4tQ2xpOiBwcm9qZWN0MSwgcHJvamVjdDJcbiIsIm1lcm1haWQiOnsidGhlbWUiOiJuZXV0cmFsIn0sInVwZGF0ZUVkaXRvciI6ZmFsc2V9
   :width: 600
   :alt: CLI Authentication with Admin's own Authorization Server

Using an External Authorization Server:

.. image:: https://mermaid.ink/img/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4lJXtjb25maWc6IHsgJ2ZvbnRGYW1pbHknOiAnTWVubG8nLCAnZm9udFNpemUnOiAxMCwgJ2ZvbnRXZWlnaHQnOiAxMDB9IH0lJVxuICAgIGF1dG9udW1iZXJcbiAgICBVc2VyLT4-K0NsaTogZmx5dGVjdGwgbGlzdC1wcm9qZWN0c1xuICAgIENsaS0-PitBZG1pbjogYWRtaW4vY2xpZW50LWNvbmZpZ1xuICAgIEFkbWluLT4-LUNsaTogQ2xpZW50X2lkPTxhYmM-LCAuLi5cbiAgICBDbGktPj4rQnJvd3NlcjogL29hdXRoMi9hdXRob3JpemU_cGtjZSZjb2RlX2NoYWxsZW5nZSxjbGllbnRfaWQsc2NvcGVcbiAgICBCcm93c2VyLT4-K0V4dGVybmFsSWRwOiAvb2F1dGgyL2F1dGhvcml6ZT9wa2NlLi4uXG4gICAgRXh0ZXJuYWxJZHAtPj4tQnJvd3NlcjogMzAyIGlkcC5jb20vbG9naW5cbiAgICBOb3RlIG92ZXIgQnJvd3NlcixFeHRlcm5hbElkcDogVGhlIHByaW9yIE9wZW5JRCBDb25uZWN0IGZsb3dcbiAgICBCcm93c2VyLT4-K0V4dGVybmFsSWRwOiAvbG9nZ2VkX2luXG4gICAgTm90ZSBvdmVyIEJyb3dzZXIsRXh0ZXJuYWxJZHA6IFBvdGVudGlhbGx5IHNob3cgY3VzdG9tIGNvbnNlbnQgc2NyZWVuXG4gICAgRXh0ZXJuYWxJZHAtPj4tQnJvd3NlcjogbG9jYWxob3N0Lz9hdXRoQ29kZT08YWJjPlxuICAgIEJyb3dzZXItPj4rQ2xpOiBsb2NhbGhvc3QvYXV0aENvZGU9PGFiYz5cbiAgICBDbGktPj4rRXh0ZXJuYWxJZHA6IC90b2tlbj9jb2RlLGNvZGVfdmVyaWZpZXJcbiAgICBFeHRlcm5hbElkcC0-Pi1DbGk6IGFjY2Vzc190b2tlblxuICAgIENsaS0-PitBZG1pbjogL3Byb2plY3RzLyArIGFjY2Vzc190b2tlblxuICAgIEFkbWluLT4-LUNsaTogcHJvamVjdDEsIHByb2plY3QyXG4iLCJtZXJtYWlkIjp7InRoZW1lIjoibmV1dHJhbCJ9LCJ1cGRhdGVFZGl0b3IiOmZhbHNlfQ
   :target: https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4lJXtjb25maWc6IHsgJ2ZvbnRGYW1pbHknOiAnTWVubG8nLCAnZm9udFNpemUnOiAxMCwgJ2ZvbnRXZWlnaHQnOiAxMDB9IH0lJVxuICAgIGF1dG9udW1iZXJcbiAgICBVc2VyLT4-K0NsaTogZmx5dGVjdGwgbGlzdC1wcm9qZWN0c1xuICAgIENsaS0-PitBZG1pbjogYWRtaW4vY2xpZW50LWNvbmZpZ1xuICAgIEFkbWluLT4-LUNsaTogQ2xpZW50X2lkPTxhYmM-LCAuLi5cbiAgICBDbGktPj4rQnJvd3NlcjogL29hdXRoMi9hdXRob3JpemU_cGtjZSZjb2RlX2NoYWxsZW5nZSxjbGllbnRfaWQsc2NvcGVcbiAgICBCcm93c2VyLT4-K0V4dGVybmFsSWRwOiAvb2F1dGgyL2F1dGhvcml6ZT9wa2NlLi4uXG4gICAgRXh0ZXJuYWxJZHAtPj4tQnJvd3NlcjogMzAyIGlkcC5jb20vbG9naW5cbiAgICBOb3RlIG92ZXIgQnJvd3NlcixFeHRlcm5hbElkcDogVGhlIHByaW9yIE9wZW5JRCBDb25uZWN0IGZsb3dcbiAgICBCcm93c2VyLT4-K0V4dGVybmFsSWRwOiAvbG9nZ2VkX2luXG4gICAgTm90ZSBvdmVyIEJyb3dzZXIsRXh0ZXJuYWxJZHA6IFBvdGVudGlhbGx5IHNob3cgY3VzdG9tIGNvbnNlbnQgc2NyZWVuXG4gICAgRXh0ZXJuYWxJZHAtPj4tQnJvd3NlcjogbG9jYWxob3N0Lz9hdXRoQ29kZT08YWJjPlxuICAgIEJyb3dzZXItPj4rQ2xpOiBsb2NhbGhvc3QvYXV0aENvZGU9PGFiYz5cbiAgICBDbGktPj4rRXh0ZXJuYWxJZHA6IC90b2tlbj9jb2RlLGNvZGVfdmVyaWZpZXJcbiAgICBFeHRlcm5hbElkcC0-Pi1DbGk6IGFjY2Vzc190b2tlblxuICAgIENsaS0-PitBZG1pbjogL3Byb2plY3RzLyArIGFjY2Vzc190b2tlblxuICAgIEFkbWluLT4-LUNsaTogcHJvamVjdDEsIHByb2plY3QyXG4iLCJtZXJtYWlkIjp7InRoZW1lIjoibmV1dHJhbCJ9LCJ1cGRhdGVFZGl0b3IiOmZhbHNlfQ
   :width: 600
   :alt: CLI Authentication with an external Authorization Server

Identity Providers Support
==========================

+-----------------+--------+-------------+---------------------+----------+-------+----------+--------+
| Feature         | Okta   | Google free | GC Identity Service | Azure AD | Auth0 | KeyCloak | Github |
+=================+========+=============+=====================+==========+=======+==========+========+
| OpenIdConnect   |   Yes  |     Yes     |          Yes        |    Yes   |  Yes  |    Yes   |   No   |
+-----------------+--------+-------------+---------------------+----------+-------+----------+--------+
| Custom RP       |   Yes  |      No     |          Yes        |    Yes   |   ?   |    Yes   |   No   |
+-----------------+--------+-------------+---------------------+----------+-------+----------+--------+

.. _auth-setup:

####################
Authentication Setup
####################

*****************
IdP Configuration
*****************
Flyte Admin requires that the application in your identity provider be configured as a web client (i.e. with a client secret). We recommend allowing the application to be issued a refresh token to avoid interrupting the user's flow by frequently redirecting to the IdP.

*************************
Flyte Admin Configuration
*************************
Please refer to the `inline documentation <https://github.com/flyteorg/flyteadmin/blob/eaca2fb0e6018a2e261e9e2da8998906477cadb5/pkg/auth/config/config.go>`_ on the ``Config`` object in the ``auth`` package for a discussion on the settings required.

**********************
Example Configurations
**********************

Below are some canonical examples of how to set up some of the common IdPs to secure your Fyte services. OpenID Connect enables users to authenticate, in the
browser, with an existing IdP. Flyte also allows connecting to an external OAuth2 Authorization Server to allow centrally managed third party app access.

OpenID Connect
===============

OpenID Connect allows users to authenticate to Flyte in their browser using a familiar authentication provider (perhaps an organization-wide configured IdP).
Flyte supports connecting with external OIdC providers. Here are some examples for how to set these up:

Google OpenID Connect
=====================

Follow `Google Docs <https://developers.google.com/identity/protocols/oauth2/openid-connect>`__ on how to configure the IdP for OpenIDConnect.

.. note::

  Make sure to create an OAuth2 Client Credential. The `client_id` and `client_secret` will be needed in the following
  steps.

Okta OpenID Connect
===================

Okta supports OpenID Connect protocol and the creation of custom OAuth2 Authorization Servers, allowing it to act as both the user and apps IdP.
It offers more detailed control on access policies, user consent, and app management.

1. If you don't already have an Okta account, sign up for one `here <https://developer.okta.com/signup/>`__.
2. Create an app (choose Web for the platform) and OpenID Connect for the sign-on method.
3. Add Login redirect URIs (e.g. http://localhost:30081/callback for sandbox or ``https://<your deployment url>/callback``)
4. *Optional*: Add logout redirect URIs (e.g. http://localhost:30081/logout for sandbox)
5. Write down the Client ID and Client Secret

KeyCloak OpenID Connect
=======================

`KeyCloak <https://www.keycloak.org/>`__ is an open source solution for authentication, it supports both OpenID Connect and OAuth2 protocols (among others). 
KeyCloak can be configured to be both the OpenID Connect and OAuth2 Authorization Server provider for Flyte.

1. Store the `client_secret` in a k8s secret as follows:

.. prompt:: bash

  kubectl edit secret -n flyte flyte-admin-auth

Add a new key under `stringData`:

.. code-block:: yaml

  stringData:
    oidc_client_secret: <client_secret> from the previous step
  data:
    ...

Save and close your editor.

2. Edit FlyteAdmin config to add `client_id` and configure auth as follows:

.. prompt:: bash

  kubectl get deploy -n flyte flyteadmin -o yaml | grep "name: flyte-admin-config"

This will output the name of the config map where the `client_id` needs to go.

.. prompt:: bash

  kubectl edit configmap -n flyte <the name of the config map from previous command>

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

3. Restart `flyteadmin` for the changes to take effect:

.. prompt:: bash

  kubectl rollout restart deployment/flyteadmin -n flyte

***************************
OAuth2 Authorization Server
***************************

An OAuth2 Authorization Server allows external clients to request to authenticate and act on behalf of users (or as their own identities). Having 
an OAuth2 Authorization Server enables Flyte administrators control over which apps can be installed and what scopes they are allowed to request or be granted (i.e. what privileges can they assume).

Flyte comes with a built-in authorization server that can be statically configured with a set of clients to request and act on behalf of the user.
The default clients are defined `here <https://github.com/flyteorg/flyteadmin/pull/168/files#diff-1267ff8bd9146e1c0ff22a9e9d53cfc56d71c1d47fed9905f95ed4bddf930f8eR74-R100>`__ 
and the corresponding section can be modified through configs.

To set up an external OAuth2 Authorization Server, please follow the instructions below:

Okta IdP
========

1. Under security -> API, click `Add Authorization Server`. Set the audience to the public URL of flyte admin (e.g. https://flyte.mycompany.io/).
2. Under `Access Policies`, click `Add New Access Policy` and walk through the wizard to allow access to the authorization server.
3. Under `Scopes`, click `Add Scope`. Set the name to `all` (required) and check `Require user consent for this scope` (recommended).
4. Create 2 apps (for fltyectl and flytepropeller) to enable these clients to communicate with the service.
   Flytectl should be created as a `native client`.
   FlytePropeller should be created as an `OAuth Service` and note the client ID and client Secrets provided.

KeyCloak IdP
============

`KeyCloak <https://www.keycloak.org/>`__ is an open source solution for authentication, it supports both OpenID Connect and OAuth2 protocols (among others). 
KeyCloak can be configured to be both the OpenID Connect and OAuth2 Authorization Server provider for flyte.

Apply Configuration
===================

1. It is possible to direct Flyte admin to use an external authorization server. To do so, edit the same config map once more and follow these changes:

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

1. Store flyte propeller's `client_secret` in a k8s secret as follows:

.. prompt:: bash

  kubectl edit secret -n flyte flyte-propeller-auth

Add a new key under `stringData`:

.. code-block:: yaml

  stringData:
    client_secret: <client_secret> from the previous step
  data:
    ...

Save and close your editor.

2. Edit FlytePropeller config to add `client_id` and configure auth as follows:

.. prompt:: bash

  kubectl get deploy -n flyte flytepropeller -o yaml | grep "name: flyte-propeller-config"

This will output the name of the config map where the `client_id` needs to go.

.. prompt:: bash

  kubectl edit configmap -n flyte <the name of the config map from previous command>

Follow the inline comments to make the necessary changes:

.. code-block:: yaml

    admin:
        # 1. Replace with the client_id provided by the OAuth2 Authorization Server above.
        clientId: flytepropeller

Close the editor

3. Restart `flytepropeller` for the changes to take effect:

.. prompt:: bash

  kubectl rollout restart deployment/flytepropeller -n flyte

***************************
Continuous Integration - CI
***************************

If your organization does any automated registration, then you'll need to authenticate with the `basic authentication <https://tools.ietf.org/html/rfc2617>`_ flow (username and password effectively). After retrieving an access token from the IDP, you can send it along to Flyte Admin as usual.

Flytekit configuration variables are automatically designed to look up values from relevant environment variables. However, to aid with continuous integration use-cases, Flytekit configuration can also reference other environment variables. 

For instance, if your CI system is not capable of setting custom environment variables like ``FLYTE_CREDENTIALS_CLIENT_SECRET`` but does set the necessary settings under a different variable, you may use ``export FLYTE_CREDENTIALS_CLIENT_SECRET_FROM_ENV_VAR=OTHER_ENV_VARIABLE`` to redirect the lookup. A ``FLYTE_CREDENTIALS_CLIENT_SECRET_FROM_FILE`` redirect is available as well, where the value should be the full path to the file containing the value for the configuration setting, in this case, the client secret. We found this redirect behavior necessary when setting up registration within our own CI pipelines.

The following is a listing of the Flytekit configuration values we set in CI, along with a brief explanation.

* ``FLYTE_CREDENTIALS_CLIENT_ID`` and ``FLYTE_CREDENTIALS_CLIENT_SECRET``
  When using basic authentication, this is the username and password.
* ``export FLYTE_CREDENTIALS_AUTH_MODE=basic``
  This tells the SDK to use basic authentication. If not set, Flytekit will assume you want to use the standard OAuth based three-legged flow.
* ``export FLYTE_CREDENTIALS_AUTHORIZATION_METADATA_KEY=text``
  At Lyft, the value is set to conform to this `header config <https://github.com/flyteorg/flyteadmin/blob/eaca2fb0e6018a2e261e9e2da8998906477cadb5/pkg/auth/config/config.go#L53>`_ on the Admin side.
* ``export FLYTE_CREDENTIALS_SCOPE=text``
  When using basic authentication, you'll need to specify a scope to the IDP (instead of ``openid``, which is only for OAuth). Set that here.
* ``export FLYTE_PLATFORM_AUTH=True``
  Set this to force Flytekit to use authentication, even if not required by Admin. This is useful as you're rolling out the requirement.


.. _migrating-auth-config:

####################################
Migrating Your Authentication Config
####################################

Using Okta as an example, you would have previously seen something like the following:

On the Okta side:
=================

* An Application (OpenID Connect Web) for Flyte Admin itself (e.g. **0oal5rch46pVhCGF45d6**).
* An Application (OpenID Native app) for Flyte-cli/flytectl (e.g. **0oal62nxuD6OSFSRq5d6**).
  These two applications would be assigned to the relevant users.
* An Application (Web) for Flyte Propeller (e.g. **0abc5rch46pVhCGF9876**).
  This application would either use the default Authorization server, or you would create a new one.

On the Admin side:
==================

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

From the Flyte-cli side, these two settings were needed:

.. code-block:: bash

    FLYTE_PLATFORM_HTTP_URL=http://localhost:8088 FLYTE_CREDENTIALS_CLIENT_ID=0oal62nxuD6OSFSRq5d6 flyte-cli ...

**FLYTE_PLATFORM_HTTP_URL** is used because **flyte-cli** uses only gRPC to communicate with Admin. It needs to know the HTTP port (which Admin hosts on a different port because of limitations of the 
grpc-gateway library). **flyte-cli** uses this setting to talk to **/.well-known/oauth-authorization-server** to retrieve information regarding the auth endpoints.  Previously this redirected to the
Okta Authorization Server's metadata endpoint. With this change, Admin now hosts its own (even if still using the external Authorization Service).

After version `0.13.0 <https://github.com/flyteorg/flyte/tree/v0.13.0>`__ of the platform, you can still use the IdP as the Authorization Server if you so choose. That configuration would now become:

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

.. code-block:: bash

    flyteadmin secrets init -p /etc/secrets

This will generate the new cookie hash/block keys, as well as other secrets Admin needs to run the Authorization server.

* The **clientSecretFile** has been moved to **/etc/secrets/oidc_client_secret** so move that there.
* **claims** has been removed, just delete that.
* **authorizeUrl** and **tokenUrl** are no longer necessary.
* The **baseUrl** for the external Authorization Server is now in the **appAuth** section.
* The **thirdPartyConfig** has been moved to **appAuth** as well.
* **redirectUrl** has been defaulted to **/console**. If that's the value you want, then you no longer need this setting.

From Propeller side, you might have a configuration section that looks like this:

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
