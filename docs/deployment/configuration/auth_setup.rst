.. _deployment-configuration-auth-setup:

#######################
Authenticating in Flyte
#######################

.. tags:: Authentication, Infrastructure, Advanced

The Flyte platform consists of multiple components. Securing communication between each component is crucial to ensure
the integrity of the overall system.

The following diagram summarizes the components and their interactions as part of Flyte's auth implementation:


.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/deployment/auth/flyte-auth-arch-v2.png


In summary, there are two main resources required for a complete auth flow in Flyte:

**An identity layer**

Using an implementation of the `Open ID Connect (OIDC) specification <https://openid.net/specs/openid-connect-core-1_0.html>`__, it enables clients to verify the identity of the end user based on the authentication performed by an Authorization Server. For this flow to work, you must first register Flyte as a new client (app) to the Identity Provider (IdP).

**An authorization server**

As defined by IETF's `RFC #6749 <https://datatracker.ietf.org/doc/html/rfc6749>`__, the authorization server's role is to issue *access tokens to the client after successfully authenticating the resource owner and obtaining authorization*. In this context, the *resource owner* is the end user of Flyte; and the *client* is the tool or component that intends to interact with ``flyteadmin`` : ``flytepropeller``, ``flyteconsole`` or any of the CLI tools.

There are two supported options to use an authorization server in Flyte:
  
 * **Internal authorization server**: It comes pre-installed with Flyte and it is a suitable choice for quick start and testing purposes. 
 * **External (custom) authorization server**: This a service provided by one of the supported IdPs and is the recommended option if your organization needs to retain control over scope definitions and grants, token expiration policies and other advanced security controls. 


.. note::

   Regardless of the type of authorization server to use, you will still need an IdP to provide identity through OIDC.


.. _auth-setup:

********************
Authentication Setup
********************

Prerequisites
=============


The following is required for non-sandbox (non ``flytectl demo``) deployments:


* A public domain name (e.g. example.foobar.com)
* Routing of traffic from that domain name to the Kubernetes Flyte Ingress IP address

.. note::

   Checkout this `community-maintained guide <https://github.com/davidmirror-ops/flyte-the-hard-way/blob/main/docs/06-intro-to-ingress.md>`__ for more information about setting up Flyte in production, including Ingress.


Identity Providers Support
==========================

Flyte supports OAuth2 and OpenId Connect to secure the various connections:

* :ref:`OpenID Connect <deployment-auth-openid-appendix>`: used to secure user's authentication to ``flyteadmin`` through the UI.
* :ref:`OAuth2 <deployment-auth-oauth2-appendix>`: used to secure connections from clients (i.e. ``pyflyte``, ``flytectl`` and
  ``flytepropeller``) to the ``flyteadmin`` service. 

Support for these protocols varies per IdP. Checkout the following table to understand the available support level for
your IdP:

+----------------------+--------+-------------+---------------------+----------+-------+----------+--------+
| Feature              | Okta   | Google free | GCP Identity Service| Azure AD | Auth0 | KeyCloak | Github |
+======================+========+=============+=====================+==========+=======+==========+========+
| OpenID Connect (OIDC)|   Yes  |     Yes     |          Yes        |    Yes   |  Yes  |    Yes   |   No   |
+----------------------+--------+-------------+---------------------+----------+-------+----------+--------+
| Custom Auth Server   |   Yes  |      No     |          Yes        |    Yes   |   ?   |    Yes   |   No   |
+----------------------+--------+-------------+---------------------+----------+-------+----------+--------+


Identity Management layer : OIDC
===================================

In this section, you can find canonical examples of how to set up OIDC on some of the supported IdPs; enabling users to authenticate in the
browser. 

.. tabs::

   .. group-tab:: Google
   


          - Create an OAuth2 Client Credential following the `official documentation  <https://developers.google.com/identity/protocols/oauth2/openid-connect>`__ and take note of the ``client_id`` and ``client_secret``

          - In the **Authorized redirect URIs** field, add ``http://localhost:30081/callback`` for **sandbox** deployments, or ``https://<your-deployment-URL>/callback`` for other methods of deployment. 
          
   
   .. group-tab:: Okta
   
   
       1. If you don't already have an Okta account, sign up for one `here <https://developer.okta.com/signup/>`__.
       2. Create an app integration, with `OIDC - OpenID Connect` as the sign-on method and `Web Application` as the app type.
       3. Add sign-in redirect URIs: 
       
          - ``http://localhost:30081/callback`` for sandbox or ``https://<your deployment url>/callback`` for other Flyte deployment types.  
          
       4. *Optional* - Add logout redirect URIs:
       
          - ``http://localhost:30081/logout`` for sandbox, ``https://<your deployment url>/callback`` for other Flyte deployment types). 
          
       5. Take note of the Client ID and Client Secret
   
   .. group-tab:: Keycloak

   
       1. If you don't have a Keycloak installation, you can use `this <https://www.amazonaws.cn/en/solutions/keycloak-on-aws/>`__ which provides a quick way to deploy Keycloak cluster on AWS.
       2. Create a realm using the `admin console <https://wjw465150.gitbooks.io/keycloak-documentation/content/server_admin/topics/realms/create.html>`__
       3. Create an OIDC client with client secret and note them down. Use the following `instructions <https://wjw465150.gitbooks.io/keycloak-documentation/content/server_admin/topics/clients/client-oidc.html>`__
       4. Add Login redirect URIs:
       
          - ``http://localhost:30081/callback`` for sandbox or ``https://<your deployment url>/callback`` for other Flyte deployment types.
   
   .. group-tab:: Microsoft Entra ID (Azure AD)
   
       1. From the Azure homepage go to **Azure Active Directory**
       2. From the **Ovierview** page, take note of the **Tenant ID**
       3. Go to **App registrations**
       4. Create a **New registration**
       5. Give it a descriptive name
       6. For the **Supported account types** select the option that matches your organization's security policy
       7. In the **Redirect URI** section select:
       
          - **Web** platform
          - Add ``http://localhost:30081/callback`` for sandbox or ``https://<your deployment url>/callback`` for other Flyte deployment types
          
       9. Click on **Register**
       10. Once created, click on the registered app and go to the **Certificates and secrets** section
       11. Go to **Client secrets** and create a **New client secret**
       12. Enter a description and an expiration policy
       13. Take note of the secret **Value** as it will be used in the Helm chart
       
       For further reference, check out the official `Azure AD Docs <https://docs.microsoft.com/en-us/power-apps/maker/portals/configure/configure-openid-settings>`__ on how to configure the IdP for OpenIDConnect.
   
       .. note::
   
         Make sure the app is registered without `additional claims <https://docs.microsoft.com/en-us/power-apps/maker/portals/configure/configure-openid-settings#configure-additional-claims>`__.
         The OpenIDConnect authentication will not work otherwise, please refer to this `GitHub Issue <https://github.com/coreos/go-oidc/issues/215>`__ and `Azure AD Docs <https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-protocols-oidc#sample-response>`__ for more information.


Apply OIDC Configuration
===========================

.. tabs::

   .. group-tab:: flyte-binary


      1. Generate a random password to be used internally by ``flytepropeller``

      2. Use the following command to generate a bcrypt hash for that password:
   
      .. prompt:: bash $

         pip install bcrypt && python -c 'import bcrypt; import base64; print(base64.b64encode(bcrypt.hashpw("<your-random-password>".encode("utf-8"), bcrypt.gensalt(6))))'

      3. Go to your values file and locate the ``auth`` section and replace values accordingly:
      
      .. code-block:: yaml

         auth:
          enabled: true
          oidc:
            # baseUrl: https://accounts.google.com # Uncomment for Google
            # baseUrl: https://<keycloak-url>/auth/realms/<keycloak-realm> # Uncomment for Keycloak and update with your installation host and realm name
            # baseUrl: https://login.microsoftonline.com/<tenant-id>/oauth2/v2.0/authorize # Uncomment for Azure AD
            # For Okta use the Issuer URI from Okta's default auth server
            baseUrl: https://dev-<org-id>.okta.com/oauth2/default
            # Replace with the client ID and secret created for Flyte in your IdP
            clientId: <client_ID>
            clientSecret: <client_secret>
          internal:
            clientSecret: '<your-random-password>'
            # Use the output of step #2 (only the content inside of '')

            clientSecretHash: <your-hashed-password>

          authorizedUris:
          - https://<your-flyte-deployment-URL>

      4. Save your changes
      5. Upgrade your Helm release with the new values:

      .. prompt:: bash $
      
         helm upgrade <release-name> flyteorg/flyte-binary -n <your-namespace> --values <your-values-file>.yaml     

      Where:
      
      * ``<release-name>`` is the name of your Helm release, typically ``flyte-backend``. You can find it using ``helm ls -n <your-namespace>``
        

      6. Verify that your Flyte deployment now requires successful login to your IdP to access the UI (``https://<your domain>/console``)

      7. For ``flytectl`` / ``pyflyte``, make sure that your local config file (``$HOME/.flyte/config.yaml``) includes the following option:

      .. code-block:: yaml

         admin:
           ...
           authType: Pkce #change from the default `clientCred` to enable client auth without using shared secrets
           ... 


   .. group-tab:: flyte-core

      1. Generate a random password to be used internally by flytepropeller
      2. Use the following command to generate a bcrypt hash for that password:
   
      .. prompt:: bash $
         
         pip install bcrypt && python -c 'import bcrypt; import base64; print(base64.b64encode(bcrypt.hashpw("<your-random-password>".encode("utf-8"), bcrypt.gensalt(6))))'
      
      Take note of the output (only the contents inside `''`)

      3. Store the ``client_secret`` provided by your IdP in a Kubernetes secret as follows:

      .. prompt:: bash $

         kubectl edit secret -n <flyte-namespace> flyte-admin-secrets

      Where ``flyte-namespace`` is the Kubernetes namespace where you have installed Flyte.

      4. Add a new key under ``stringData``:

      .. code-block:: yaml

         apiVersion: v1
         # Add from here
         stringData:
           oidc_client_secret: <client_secret from the previous step>
         # End here
         data:
         ...

      5. Save and close your editor.

      6. Go to your Helm values file and verify that the ``configmap`` section include the following, replacing the content where indicated:

      .. code-block:: yaml

         configmap:
           adminServer:
             server:
               httpPort: 8088
               grpcPort: 8089
             security:
               secure: false
               useAuth: true
               allowCors: true
               allowedOrigins:
          # Accepting all domains for Sandbox installation
                 - "*"
               allowedHeaders:
                 - "Content-Type"
             auth:
               appAuth:
                 thirdPartyConfig:
                   flyteClient:
                     clientId: flytectl
                     redirectUri: http://localhost:53593/callback
                     scopes:
                       - offline
                       - all
                 selfAuthServer:
                   staticClients:
                     flyte-cli:
                       id: flyte-cli
                       redirect_uris:
                       - http://localhost:53593/callback
                       - http://localhost:12345/callback 
                       grant_types:
                         - refresh_token
                         - authorization_code
                       response_types:
                         - code
                         - token
                       scopes:
                         - all
                         - offline
                         - access_token
                       public: true
                     flytectl:
                       id: flytectl
                       redirect_uris:
                         - http://localhost:53593/callback
                         - http://localhost:12345/callback
                       grant_types:
                         - refresh_token
                         - authorization_code
                       response_types:
                         - code
                         - token
                       scopes:
                         - all
                         - offline
                         - access_token
                       public: true
                     flytepropeller:
                       id: flytepropeller
              # Use the bcrypt hash generated for your random password
                       client_secret: "<bcrypt-hash>" 
                       redirect_uris:
                         - http://localhost:3846/callback
                       grant_types:
                         - refresh_token
                         - client_credentials
                       response_types:
                         - token
                       scopes:
                         - all
                         - offline
                         - access_token
                       public: false
              
               authorizedUris:
               # Use the public URL of flyteadmin (a DNS record pointing to your Ingress resource)
                 - https://<your-flyte-deployment-URL>
                 - http://flyteadmin:80
                 - http://flyteadmin.flyte.svc.cluster.local:80
               userAuth:
                 openId:
                # baseUrl: https://accounts.google.com # Uncomment for Google
                # baseUrl: https://login.microsoftonline.com/<tenant-id>/v2.0 # Uncomment for Azure AD
                  # For Okta, use the Issuer URI of the default auth server
                  baseUrl: https://dev-<org-id>.okta.com/oauth2/default
                  # Use the client ID generated by your IdP
                  clientId: <client_ID>
                  scopes:
                    - profile
                    - openid

      7. Additionally, outside the ``configmap`` section, add the following block and replace the necessary information:
      
      .. code-block:: yaml
         
         secrets:
           adminOauthClientCredentials:
          # -- If enabled is true, helm will create and manage `flyte-secret-auth` and populate it with `clientSecret`.
          # If enabled is false, it's up to the user to create `flyte-secret-auth`
             enabled: true
           # Use the non-encoded version of the random password 
             clientSecret: "<your-random-password>" 
             clientId: flytepropeller

.. note::

   For `multi-cluster deployments <https://docs.flyte.org/en/latest/deployment/deployment/multicluster.html>`__, 
   you must add this Secret definition block to the `values-dataplane.yaml` file.
   If you are not running `flytepropeller` in the control plane cluster, you do not need to create this secret there.

      8. Save and exit your editor.

      9. Upgrade your Helm release with the new configuration:

      .. prompt:: bash $

         helm upgrade <release-name> flyteorg/flyte-binary -n <your-namespace> --values <your-values-file>.yaml

      10. Verify that the `flytepropeller`, `flytescheduler` and `flyteadmin` Pods are restarted and running: 

      .. prompt:: bash $

          kubectl get pods -n flyte

      11. For flytectl/pyflyte, make sure that your local config file (``$HOME/.flyte/config.yaml``) includes the following option:

      .. code-block:: yaml

         admin:
           ...
           authType: Pkce #change from the default `clientCred` to enable client auth without using shared secrets
           ...    

.. note::

   **Congratulations!**

   It should now be possible to go to Flyte UI and be prompted for authentication. Flytectl should automatically pickup the change and start prompting for authentication as well.
   If you want to use an external OAuth2 provider for App authentication, please continue reading into the next section.

***************************
Custom Authorization Server
***************************


As mentioned previously, Flyte ships with an internal authorization server; hence setting up an external Authorization Server is optional and dependent on your organization's security requirements. 

In this section, you will find instructions on how to setup an OAuth2 Authorization Server in the different IdPs supported by Flyte:

.. note::

   **Google IdP**

   Google IdP does not offer an OAuth2 Authorization Server that could be used to protect external services (For example Flyte). In this case, Google offers a separate Cloud Product called Google Cloud Identity.
   Configuration for Cloud Identity is not included in this guide. If unavailable, setup can stop here and FlyteAdmin BuiltIn OAuth2 Authorization Server can be used instead.

.. tabs::

   .. group-tab:: Okta
   
       Okta's custom authorization servers are available through an add-on license. The free developer accounts do include access, which you can use to test before rolling out the configuration more broadly.
   
       1. From the left-hand menu, go to **Security** > **API**
       2. Click on **Add Authorization Server**. 
       3. Assign an informative name and set the audience to the public URL of FlyteAdmin (e.g. https://example.foobar.com).
   
       .. note::
   
          The audience must exactly match one of the URIs in the ``authorizedUris`` section above
   
       4. Note down the **Issuer URI**; this will be used for all the ``baseUrl`` settings in the Flyte config.  
       5. Go to **Scopes** and click **Add Scope**. 
       6. Set the name to ``all`` (required) and check ``Required`` under the **User consent** option.   
       7. Uncheck the **Block services from requesting this scope** option and save your changes.    
       8. Add another scope, named ``offline``. Check both the **Required** and **Include in public metadata** options.
       9. Uncheck the **Block services from requesting this scope** option. 
       10. Click **Save**. 
       11. Go to  **Access Policies**, click **Add New Access Policy**. Enter a name and description and enable **Assign to** -  ``All clients``.  
       12. Add a rule to the policy with the default settings (you can fine-tune these later).
       13. Navigate back to the **Applications** section.
       14. Create an integration for ``flytectl``; it should be created with the **OIDC - OpenID Connect** sign-on method, and the **Native Application** type.
       15. Add ``http://localhost:53593/callback`` to the sign-in redirect URIs. The other options can remain as default.
       16. Assign this integration to any Okta users or groups who should be able to use the ``flytectl`` tool.
       17. Note down the **Client ID**; there will not be a secret.
       18. Create an integration for ``flytepropeller``; it should be created with the **OIDC - OpenID Connect** sign-on method and **Web Application** type.
       19. Check the ``Client Credentials`` option under **Client acting on behalf of itself**.
       20. This app does not need a specific redirect URI; nor does it need to be assigned to any users.
       21. Note down the **Client ID** and **Client secret**; you will need these later.
       22. Take note of the **Issuer URI** for your Authorization Server. It will be used as the baseURL parameter in the Helm chart
   
       You should have three integrations total - one for the web interface (``flyteconsole``), one for ``flytectl``, and one for ``flytepropeller``.
   
   .. group-tab:: Keycloak
   
   
       1. If you don't have a Keycloak installation, you can use `this <https://www.amazonaws.cn/en/solutions/keycloak-on-aws/>`__ which provides quick way to deploy Keycloak cluster on AWS.
       2. Create a realm in keycloak installation using its `admin console <https://wjw465150.gitbooks.io/keycloak-documentation/content/server_admin/topics/realms/create.html>`__
       3. Under `Client Scopes`, click `Add Create` inside the admin console.
       4. Create two clients (for `flytectl` and `flytepropeller`) to enable these clients to communicate with the service.
       5. `flytectl` should be created with `Access Type Public` and standard flow enabled.
       6. `flytePropeller` should be created as an `Access Type Confidential`, enabling the standard flow
       7. Take note of the client ID and client Secrets provided.

   .. group-tab:: Azure AD
   
       1. Navigate to tab **Overview**, obtain ``<client id>`` and ``<tenant id>``
       2. Navigate to tab **Authentication**, click ``+Add a platform``
       3. Add **Web** for flyteconsole and flytepropeller, **Mobile and desktop applications** for flytectl.
       4. Add URL ``https://<console-url>/callback`` as the callback for Web
       5. Add URL ``http://localhost:53593/callback`` as the callback for flytectl
       6. In **Advanced settings**, set ``Enable the following mobile and desktop flows`` to **Yes** to enable deviceflow
       7. Navigate to tab **Certificates & secrets**, click ``+New client secret`` to create ``<client secret>``
       8. Navigate to tab **Token configuration**, click ``+Add optional claim`` and create email claims for both ID and Access Token
       9.  Navigate to tab **API permissions**, add ``email``, ``offline_access``, ``openid``, ``profile``, ``User.Read``
       10. Navigate to tab **Expose an API**, Click ``+Add a scope`` and ``+Add a client application`` to create ``<custom scope>``


Apply external auth server configuration
========================================

Follow the steps in this section to configure `flyteadmin` to use an external auth server. This section assumes that you have already completed and applied the configuration for the OIDC Identity Layer.

.. tabs::

   .. group-tab:: flyte-binary
      
      1. Go to the values YAML file you used to install Flyte using a Helm chart
      2. Find the ``auth`` section and follow the inline comments to insert your configuration:
      
      .. code-block:: yaml

         auth:
           enabled: true
           oidc:
           # baseUrl: https://<keycloak-url>/auth/realms/<keycloak-realm> # Uncomment for Keycloak and update with your installation host and realm name
           # baseUrl: https://login.microsoftonline.com/<tenant-id>/oauth2/v2.0/authorize # Uncomment for Azure AD
           # For Okta, use the Issuer URI of the custom auth server:
             baseUrl: https://dev-<org-id>.okta.com/oauth2/<auth-server-id>            
           # Use the client ID and secret generated by your IdP for the first OIDC registration in the "Identity Management layer : OIDC" section of this guide
             clientId: <oidc-clientId>
             clientSecret: <oidc-clientSecret>
           internal:
           # Use the clientID generated by your IdP for the flytepropeller app registration
             clientId: <flytepropeller-client-id>
           #Use the secret generated by your IdP for flytepropeller
             clientSecret: '<flytepropeller-client-secret-non-encoded>'
           # Use the bcrypt hash for the clientSecret
             clientSecretHash: <-flytepropeller-secret-bcrypt-hash>
           authorizedUris:
           # Use here the exact same value used for 'audience' when the Authorization server was configured
           - https://<your-flyte-deployment-URL>     
        
        
      3. Find the ``inline`` section of the values file and add the following content, replacing where needed:
      
      .. code-block:: yaml

         inline:
           auth:
             appAuth:
               authServerType: External
               externalAuthServer:
               # baseUrl: https://<keycloak-url>/auth/realms/<keycloak-realm> # Uncomment for Keycloak and update with your installation host and realm name
               # baseUrl: https://login.microsoftonline.com/<tenant-id>/oauth2/v2.0/authorize # Uncomment for Azure AD
               # For Okta, use the Issuer URI of the custom auth server:
                 baseUrl: https://dev-<org-id>.okta.com/oauth2/<auth-server-id>  
                 metadataUrl: .well-known/oauth-authorization-server 
               thirdPartyConfig:
                 flyteClient:
                   # Use the clientID generated by your IdP for the `flytectl` app registration
                   clientId: <flytectl-client-id>
                   redirectUri: http://localhost:53593/callback
                   scopes:
                   - offline
                   - all
             userAuth:
               openId:
               # baseUrl: https://<keycloak-url>/auth/realms/<keycloak-realm> # Uncomment for Keycloak and update with your installation host and realm name
               # baseUrl: https://login.microsoftonline.com/<tenant-id>/oauth2/v2.0/authorize # Uncomment for Azure AD
               # For Okta, use the Issuer URI of the custom auth server:  
                 baseUrl: https://dev-<org-id>.okta.com/oauth2/<auth-server-id>
                 scopes:  
                 - profile  
                 - openid 
               # - offline_access # Uncomment if your IdP supports issuing refresh tokens (optional) 
               # Use the client ID and secret generated by your IdP for the first OIDC registration in the "Identity Management layer : OIDC" section of this guide  
                 clientId: <oidc-clientId>
      
      
      4. Save your changes
      5. Upgrade your Helm release with the new configuration:

      .. prompt:: bash $

         helm upgrade  <release-name> flyteorg/flyte-core -n <your-namespace> --values <your-updated-values-filel>.yaml

        
   .. group-tab:: flyte-core

       
      1. Find the ``auth`` section in your Helm values file, and replace the necessary data:

      .. note:: 

         If you were previously using the internal auth server, make sure to delete all the ``selfAuthServer`` section from your values file

      .. code-block:: yaml
          
         configmap:
           auth:
             appAuth:

               authServerType: External

               # 2. Optional: Set external auth server baseUrl if different from OpenId baseUrl.
               externalAuthServer:
               # baseUrl: https://<keycloak-url>/auth/realms/<keycloak-realm> # Uncomment for Keycloak and update with your installation host and realm name
               # baseUrl: https://login.microsoftonline.com/<tenant-id>/oauth2/v2.0/authorize # Uncomment for Azure AD
               # For Okta, use the Issuer URI of the custom auth server:  
                 baseUrl: https://dev-<org-id>.okta.com/oauth2/<auth-server-id>
               
                 metadataUrl: .well-known/openid-configuration

               thirdPartyConfig:
                  flyteClient:
                     # 3. Replace with a new Native/Public Client ID provisioned in the custom authorization server.
                     clientId: flytectl
                     # This should not change
                     redirectUri: http://localhost:53593/callback
                     # 4. "all" is a required scope and must be configured in the custom authorization server.
                     scopes:
                     - offline
                     - all

             userAuth:
               openId:
               # baseUrl: https://<keycloak-url>/auth/realms/<keycloak-realm> # Uncomment for Keycloak and update with your installation host and realm name
               # baseUrl: https://login.microsoftonline.com/<tenant-id>/oauth2/v2.0/authorize # Uncomment for Azure AD
               # For Okta, use the Issuer URI of the custom auth server:  
                 baseUrl: https://dev-<org-id>.okta.com/oauth2/<auth-server-id>
                 scopes:
                 - profile
                 - openid
                 # - offline_access # Uncomment if OIdC supports issuing refresh tokens.
                 clientId: <client id>
                  
   
         secrets:
           adminOauthClientCredentials:
             enabled: true # see the section "Disable Helm secret management" if you require to do so
             # Replace with the client_secret provided by your IdP for flytepropeller.
             clientSecret: <client_secret>
             # Replace with the client_id provided by provided by your IdP for flytepropeller.
             clientId: <client_id>
      
      2. Save your changes
      3. Upgrade your Helm release with the new configuration:

      .. prompt:: bash $

         helm upgrade  <release-name> flyteorg/flyte-core -n <your-namespace> --values <your-updated-values-file>.yaml

   .. group-tab:: flyte-core with Azure AD

      .. code-block:: yaml

         secrets:
         adminOauthClientCredentials:
            enabled: true
            clientSecret: <client secret>
            clientId: <client id>
         ---
         configmap:
         admin:
            admin:
               endpoint: <admin endpoint>
               insecure: true
               clientId: <client id>
               clientSecretLocation: /etc/secrets/client_secret
               scopes:
               - api://<client id>/.default
               useAudienceFromAdmin: true
         ---
         auth:
            appAuth:
               authServerType: External
               externalAuthServer:
                  baseUrl: https://login.microsoftonline.com/<tenant id>/v2.0/
                  metadataUrl: .well-known/openid-configuration
                  AllowedAudience:
                     - api://<client id>
               thirdPartyConfig:
                  flyteClient:
                     clientId: <client id>
                     redirectUri: http://localhost:53593/callback
                     scopes:
                     - api://<client id>/<custom-scope>

            userAuth:
               openId:
                  baseUrl: https://login.microsoftonline.com/<tenant id>/v2.0
                  scopes:
                     - openid
                     - profile
                  clientId: <client id>

.. note::

   **Congratulations**

   At this point, every interaction with Flyte components -be it in the UI or CLI- should require a successful login to your IdP, where your security policies are maintained and enforced.


Disable Helm secret management
------------------------------

Alternatively, you can instruct Helm not to create and manage the secret for ``flytepropeller``. In that case, you'll have to create it following these steps:


1. Disable Helm secrets management in your values file

.. code-block:: yaml

   secrets:
     adminOauthClientCredentials:
       enabled: false #set to false
       # Replace with the client_id provided by provided by your IdP for flytepropeller.
       clientId: <client_id> 

2. Create a secret declaratively:

.. code-block:: yaml

   apiVersion: v1
   kind: Secret
   metadata:
    name: flyte-secret-auth
    namespace: flyte
   type: Opaque
   stringData:
  # Replace with the client_secret provided by your IdP for flytepropeller.
     client_secret: <client_secret>



Continuous Integration - CI
---------------------------

If your organization does any automated registration, then you'll need to authenticate with the `client credentials <https://datatracker.ietf.org/doc/html/rfc6749#section-4.4>`_ flow. After retrieving an access token from the IDP, you can send it along to `flyteadmin`` as usual.

.. tabs::

   .. group-tab:: flytectl
   
      Flytectl's `config.yaml <https://docs.flyte.org/en/latest/flytectl/overview.html#configuration>`_ can be
      configured to use either PKCE (`Proof key for code exchange <https://datatracker.ietf.org/doc/html/rfc7636>`_)
      or Client Credentials (`Client Credentials <https://datatracker.ietf.org/doc/html/rfc6749#section-4.4>`_) flows.
   
      1. Update ``config.yaml`` as follows:
   
      .. code-block:: yaml
   
          admin:
              # Update with the Flyte's ingress endpoint (e.g. flyteIngressIP for sandbox or example.foobar.com)
              # You must keep the 3 forward-slashes after dns:
              endpoint: dns:///<Flyte ingress url>
   
              # Update auth type to `Pkce` or `ClientSecret`
              authType: Pkce
   
              # Set to the clientId (will be used for both Pkce and ClientSecret flows)
              # Leave empty to use the value discovered through flyteAdmin's Auth discovery endpoint.
              clientId: <Id>
   
              # Set to the location where the client secret is mounted.
              # Only needed/used for `ClientSecret` flow.
              clientSecretLocation: </some/path/to/key>
   
              # If required, set the scopes needed here. Otherwise, flytectl will discover scopes required for OpenID
              # Connect through flyteAdmin's Auth discovery endpoint.
              # scopes: [ "scope1", "scope2" ]
   
      To read further about the available config options, please
      `visit here <https://github.com/flyteorg/flyteidl/blob/master/clients/go/admin/config.go#L37-L64>`_
   
   .. group-tab:: Flytekit / pyflyte
   
      Flytekit configuration variables are automatically designed to look up values from relevant environment variables.

      .. important::

         However, to aid with continuous integration use-cases, Flytekit configuration can also reference other environment
         variables.
   
         For instance, if your CI system is not capable of setting custom environment variables like
         ``FLYTE_CREDENTIALS_CLIENT_SECRET`` but does set the necessary settings under a different variable, you may use
         ``export FLYTE_CREDENTIALS_CLIENT_SECRET_FROM_ENV_VAR=OTHER_ENV_VARIABLE`` to redirect the lookup. A
         ``FLYTE_CREDENTIALS_CLIENT_SECRET_FROM_FILE`` redirect is available as well, where the value should be the full
         path to the file containing the value for the configuration setting, in this case, the client secret. We found
         this redirect behavior necessary when setting up registration within our own CI pipelines.
   
      The following is a listing of the Flytekit configuration values we set in CI, along with a brief explanation.
   
      .. code-block:: bash
   
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

**********
References
**********

This collection of RFCs may be helpful to those who wish to investigate the implementation in more depth.

* `OAuth2 RFC 6749 <https://tools.ietf.org/html/rfc6749>`_
* `OAuth Discovery RFC 8414 <https://tools.ietf.org/html/rfc8414>`_
* `PKCE RFC 7636 <https://tools.ietf.org/html/rfc7636>`_
* `JWT RFC 7519 <https://tools.ietf.org/html/rfc7519>`_

There's also more detailed information about the authentication flows in the :ref:`deployment-configuration-auth-appendix`.
