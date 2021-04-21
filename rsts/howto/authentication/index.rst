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

********
Overview
********

Flyte system consists of multiple components. For the purposes of this document, let's categorize them into server-side and client-side components:

- **Admin**: A server-side control plane component  accessible from console, cli and other backends.
- **Catalog**: A server-side control plane component accessible from console, cli and other backends.
- **Console**: A client-side single page react app.
- **flyte-cli**: A python-based client-side command line interface that interacts with Admin and Catalog.
- **flytectl**: A go-based client-side command line interface that interacts with Admin and Catalog.
- **Propeller**: A server-side data plane component that interacts with both admin and catalog services.

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

.. image:: https://mermaid.ink/img/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4lJXtjb25maWc6IHsgJ2ZvbnRGYW1pbHknOiAnTWVubG8nLCAnZm9udFNpemUnOiAxMCwgJ2ZvbnRXZWlnaHQnOiAxMDB9IH0lJVxuICAgIGF1dG9udW1iZXJcbiAgICBVc2VyLT4-K0NsaTogZmx5dGVjdGwgbGlzdC1wcm9qZWN0c1xuICAgIENsaS0-PitBZG1pbjogYWRtaW4vbGlzdC1wcm9qZWN0cy9cbiAgICBBZG1pbi0-Pi1DbGk6IDQwMSBVbmF1dGhvcml6ZWRcbiAgICBDbGktPj4rQWRtaW46IGFkbWluL2NsaWVudC1jb25maWdcbiAgICBBZG1pbi0-Pi1DbGk6IENsaWVudF9pZD08YWJjPiwgLi4uXG4gICAgQ2xpLT4-K0Jyb3dzZXI6IC9hdXRoP3BrY2UmY29kZV9jaGFsbGVuZ2UsY2xpZW50X2lkLHNjb3BlXG4gICAgQnJvd3Nlci0-PitBZG1pbjogYWRtaW4vb2F1dGgyP3BrY2UuLi5cbiAgICBBZG1pbi0-Pi1Ccm93c2VyOiAzMDIgaWRwLmNvbS9sb2dpblxuICAgIE5vdGUgb3ZlciBCcm93c2VyLEFkbWluOiBUaGUgcHJpb3IgT3BlbklEIENvbm5lY3QgZmxvd1xuICAgIEJyb3dzZXItPj4rQWRtaW46IGFkbWluL2xvZ2dlZF9pblxuICAgIE5vdGUgb3ZlciBCcm93c2VyLEFkbWluOiBQb3RlbnRpYWxseSBzaG93IGN1c3RvbSBjb25zZW50IHNjcmVlblxuICAgIEFkbWluLT4-LUJyb3dzZXI6IGxvY2FsaG9zdC8_YXV0aENvZGU9PGFiYz5cbiAgICBCcm93c2VyLT4-K0NsaTogbG9jYWxob3N0L2F1dGhDb2RlPTxhYmM-XG4gICAgQ2xpLT4-K0FkbWluOiAvdG9rZW4_Y29kZSxjb2RlX3ZlcmlmaWVyXG4gICAgQWRtaW4tPj4tQ2xpOiBhY2Nlc3NfdG9rZW5cbiAgICBDbGktPj4rQWRtaW46IC9wcm9qZWN0cy8gKyBhY2Nlc3NfdG9rZW5cbiAgICBBZG1pbi0-Pi1DbGk6IHByb2plY3QxLCBwcm9qZWN0MlxuIiwibWVybWFpZCI6eyJ0aGVtZSI6Im5ldXRyYWwifSwidXBkYXRlRWRpdG9yIjpmYWxzZX0
   :target: https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4lJXtjb25maWc6IHsgJ2ZvbnRGYW1pbHknOiAnTWVubG8nLCAnZm9udFNpemUnOiAxMCwgJ2ZvbnRXZWlnaHQnOiAxMDB9IH0lJVxuICAgIGF1dG9udW1iZXJcbiAgICBVc2VyLT4-K0NsaTogZmx5dGVjdGwgbGlzdC1wcm9qZWN0c1xuICAgIENsaS0-PitBZG1pbjogYWRtaW4vbGlzdC1wcm9qZWN0cy9cbiAgICBBZG1pbi0-Pi1DbGk6IDQwMSBVbmF1dGhvcml6ZWRcbiAgICBDbGktPj4rQWRtaW46IGFkbWluL2NsaWVudC1jb25maWdcbiAgICBBZG1pbi0-Pi1DbGk6IENsaWVudF9pZD08YWJjPiwgLi4uXG4gICAgQ2xpLT4-K0Jyb3dzZXI6IC9hdXRoP3BrY2UmY29kZV9jaGFsbGVuZ2UsY2xpZW50X2lkLHNjb3BlXG4gICAgQnJvd3Nlci0-PitBZG1pbjogYWRtaW4vb2F1dGgyP3BrY2UuLi5cbiAgICBBZG1pbi0-Pi1Ccm93c2VyOiAzMDIgaWRwLmNvbS9sb2dpblxuICAgIE5vdGUgb3ZlciBCcm93c2VyLEFkbWluOiBUaGUgcHJpb3IgT3BlbklEIENvbm5lY3QgZmxvd1xuICAgIEJyb3dzZXItPj4rQWRtaW46IGFkbWluL2xvZ2dlZF9pblxuICAgIE5vdGUgb3ZlciBCcm93c2VyLEFkbWluOiBQb3RlbnRpYWxseSBzaG93IGN1c3RvbSBjb25zZW50IHNjcmVlblxuICAgIEFkbWluLT4-LUJyb3dzZXI6IGxvY2FsaG9zdC8_YXV0aENvZGU9PGFiYz5cbiAgICBCcm93c2VyLT4-K0NsaTogbG9jYWxob3N0L2F1dGhDb2RlPTxhYmM-XG4gICAgQ2xpLT4-K0FkbWluOiAvdG9rZW4_Y29kZSxjb2RlX3ZlcmlmaWVyXG4gICAgQWRtaW4tPj4tQ2xpOiBhY2Nlc3NfdG9rZW5cbiAgICBDbGktPj4rQWRtaW46IC9wcm9qZWN0cy8gKyBhY2Nlc3NfdG9rZW5cbiAgICBBZG1pbi0-Pi1DbGk6IHByb2plY3QxLCBwcm9qZWN0MlxuIiwibWVybWFpZCI6eyJ0aGVtZSI6Im5ldXRyYWwifSwidXBkYXRlRWRpdG9yIjpmYWxzZX0
   :width: 600
   :alt: CLI Authentication with Admin's own Authorization Server

Using an External Authorization Server:

.. image:: https://mermaid.ink/img/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4lJXtjb25maWc6IHsgJ2ZvbnRGYW1pbHknOiAnTWVubG8nLCAnZm9udFNpemUnOiAxMCwgJ2ZvbnRXZWlnaHQnOiAxMDB9IH0lJVxuICAgIGF1dG9udW1iZXJcbiAgICBVc2VyLT4-K0NsaTogZmx5dGVjdGwgbGlzdC1wcm9qZWN0c1xuICAgIENsaS0-PitBZG1pbjogYWRtaW4vbGlzdC1wcm9qZWN0cy9cbiAgICBBZG1pbi0-Pi1DbGk6IDQwMSBVbmF1dGhvcml6ZWRcbiAgICBDbGktPj4rQWRtaW46IGFkbWluL2NsaWVudC1jb25maWdcbiAgICBBZG1pbi0-Pi1DbGk6IENsaWVudF9pZD08YWJjPiwgLi4uXG4gICAgQ2xpLT4-K0Jyb3dzZXI6IC9hdXRoP3BrY2UmY29kZV9jaGFsbGVuZ2UsY2xpZW50X2lkLHNjb3BlXG4gICAgQnJvd3Nlci0-PitFeHRlcm5hbElkcDogYWRtaW4vb2F1dGgyP3BrY2UuLi5cbiAgICBFeHRlcm5hbElkcC0-Pi1Ccm93c2VyOiAzMDIgaWRwLmNvbS9sb2dpblxuICAgIE5vdGUgb3ZlciBCcm93c2VyLEV4dGVybmFsSWRwOiBUaGUgcHJpb3IgT3BlbklEIENvbm5lY3QgZmxvd1xuICAgIEJyb3dzZXItPj4rRXh0ZXJuYWxJZHA6IC9sb2dnZWRfaW5cbiAgICBOb3RlIG92ZXIgQnJvd3NlcixFeHRlcm5hbElkcDogUG90ZW50aWFsbHkgc2hvdyBjdXN0b20gY29uc2VudCBzY3JlZW5cbiAgICBFeHRlcm5hbElkcC0-Pi1Ccm93c2VyOiBsb2NhbGhvc3QvP2F1dGhDb2RlPTxhYmM-XG4gICAgQnJvd3Nlci0-PitDbGk6IGxvY2FsaG9zdC9hdXRoQ29kZT08YWJjPlxuICAgIENsaS0-PitFeHRlcm5hbElkcDogL3Rva2VuP2NvZGUsY29kZV92ZXJpZmllclxuICAgIEV4dGVybmFsSWRwLT4-LUNsaTogYWNjZXNzX3Rva2VuXG4gICAgQ2xpLT4-K0FkbWluOiAvcHJvamVjdHMvICsgYWNjZXNzX3Rva2VuXG4gICAgQWRtaW4tPj4tQ2xpOiBwcm9qZWN0MSwgcHJvamVjdDJcbiIsIm1lcm1haWQiOnsidGhlbWUiOiJuZXV0cmFsIn0sInVwZGF0ZUVkaXRvciI6ZmFsc2V9
   :target: https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4lJXtjb25maWc6IHsgJ2ZvbnRGYW1pbHknOiAnTWVubG8nLCAnZm9udFNpemUnOiAxMCwgJ2ZvbnRXZWlnaHQnOiAxMDB9IH0lJVxuICAgIGF1dG9udW1iZXJcbiAgICBVc2VyLT4-K0NsaTogZmx5dGVjdGwgbGlzdC1wcm9qZWN0c1xuICAgIENsaS0-PitBZG1pbjogYWRtaW4vbGlzdC1wcm9qZWN0cy9cbiAgICBBZG1pbi0-Pi1DbGk6IDQwMSBVbmF1dGhvcml6ZWRcbiAgICBDbGktPj4rQWRtaW46IGFkbWluL2NsaWVudC1jb25maWdcbiAgICBBZG1pbi0-Pi1DbGk6IENsaWVudF9pZD08YWJjPiwgLi4uXG4gICAgQ2xpLT4-K0Jyb3dzZXI6IC9hdXRoP3BrY2UmY29kZV9jaGFsbGVuZ2UsY2xpZW50X2lkLHNjb3BlXG4gICAgQnJvd3Nlci0-PitFeHRlcm5hbElkcDogYWRtaW4vb2F1dGgyP3BrY2UuLi5cbiAgICBFeHRlcm5hbElkcC0-Pi1Ccm93c2VyOiAzMDIgaWRwLmNvbS9sb2dpblxuICAgIE5vdGUgb3ZlciBCcm93c2VyLEV4dGVybmFsSWRwOiBUaGUgcHJpb3IgT3BlbklEIENvbm5lY3QgZmxvd1xuICAgIEJyb3dzZXItPj4rRXh0ZXJuYWxJZHA6IC9sb2dnZWRfaW5cbiAgICBOb3RlIG92ZXIgQnJvd3NlcixFeHRlcm5hbElkcDogUG90ZW50aWFsbHkgc2hvdyBjdXN0b20gY29uc2VudCBzY3JlZW5cbiAgICBFeHRlcm5hbElkcC0-Pi1Ccm93c2VyOiBsb2NhbGhvc3QvP2F1dGhDb2RlPTxhYmM-XG4gICAgQnJvd3Nlci0-PitDbGk6IGxvY2FsaG9zdC9hdXRoQ29kZT08YWJjPlxuICAgIENsaS0-PitFeHRlcm5hbElkcDogL3Rva2VuP2NvZGUsY29kZV92ZXJpZmllclxuICAgIEV4dGVybmFsSWRwLT4-LUNsaTogYWNjZXNzX3Rva2VuXG4gICAgQ2xpLT4-K0FkbWluOiAvcHJvamVjdHMvICsgYWNjZXNzX3Rva2VuXG4gICAgQWRtaW4tPj4tQ2xpOiBwcm9qZWN0MSwgcHJvamVjdDJcbiIsIm1lcm1haWQiOnsidGhlbWUiOiJuZXV0cmFsIn0sInVwZGF0ZUVkaXRvciI6ZmFsc2V9
   :width: 600
   :alt: CLI Authentication with an external Authorization Server

Identity Providers Support
==========================

+-----------------+--------+-------------+---------------------+----------+-------+----------+
| Feature         | Okta   | Google free | GC Identity Service | Azure AD | Auth0 | KeyCloak |
+=================+========+=============+=====================+==========+=======+==========+
| OpenIdConnect   |   Yes  |     Yes     |          Yes        |    Yes   |  Yes  |    Yes   |
+-----------------+--------+-------------+---------------------+----------+-------+----------+
| Custom RP       |   Yes  |      No     |          Yes        |    Yes   |   ?   |    Yes   |
+-----------------+--------+-------------+---------------------+----------+-------+----------+

**********
References
**********

This collection of RFCs may be helpful to those who wish to investigate the implementation in more depth.

* `OAuth2 RFC 6749 <https://tools.ietf.org/html/rfc6749>`_
* `OAuth Discovery RFC 8414 <https://tools.ietf.org/html/rfc8414>`_
* `PKCE RFC 7636 <https://tools.ietf.org/html/rfc7636>`_
* `JWT RFC 7519 <https://tools.ietf.org/html/rfc7519>`_


