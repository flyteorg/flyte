.. _flytescheduler-config-specification:

#########################################
Flyte Scheduler Configuration
#########################################

- `admin <#section-admin>`_

- `auth <#section-auth>`_

- `cloudevents <#section-cloudevents>`_

- `cluster_resources <#section-cluster_resources>`_

- `clusterpools <#section-clusterpools>`_

- `clusters <#section-clusters>`_

- `database <#section-database>`_

- `domains <#section-domains>`_

- `externalevents <#section-externalevents>`_

- `flyteadmin <#section-flyteadmin>`_

- `logger <#section-logger>`_

- `namespace_mapping <#section-namespace_mapping>`_

- `notifications <#section-notifications>`_

- `otel <#section-otel>`_

- `plugins <#section-plugins>`_

- `propeller <#section-propeller>`_

- `qualityofservice <#section-qualityofservice>`_

- `queues <#section-queues>`_

- `registration <#section-registration>`_

- `remotedata <#section-remotedata>`_

- `scheduler <#section-scheduler>`_

- `secrets <#section-secrets>`_

- `server <#section-server>`_

- `storage <#section-storage>`_

- `task_resources <#section-task_resources>`_

- `task_type_whitelist <#section-task_type_whitelist>`_

Section: admin
========================================================================================================================

endpoint (`config.URL`_)
------------------------------------------------------------------------------------------------------------------------

For admin types, specify where the uri of the service is located.

**Default Value**: 

.. code-block:: yaml

  ""
  

insecure (bool)
------------------------------------------------------------------------------------------------------------------------

Use insecure connection.

**Default Value**: 

.. code-block:: yaml

  "false"
  

insecureSkipVerify (bool)
------------------------------------------------------------------------------------------------------------------------

InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name. Caution : shouldn't be use for production usecases'

**Default Value**: 

.. code-block:: yaml

  "false"
  

caCertFilePath (string)
------------------------------------------------------------------------------------------------------------------------

Use specified certificate file to verify the admin server peer.

**Default Value**: 

.. code-block:: yaml

  ""
  

maxBackoffDelay (`config.Duration`_)
------------------------------------------------------------------------------------------------------------------------

Max delay for grpc backoff

**Default Value**: 

.. code-block:: yaml

  8s
  

perRetryTimeout (`config.Duration`_)
------------------------------------------------------------------------------------------------------------------------

gRPC per retry timeout

**Default Value**: 

.. code-block:: yaml

  15s
  

maxRetries (int)
------------------------------------------------------------------------------------------------------------------------

Max number of gRPC retries

**Default Value**: 

.. code-block:: yaml

  "4"
  

authType (uint8)
------------------------------------------------------------------------------------------------------------------------

Type of OAuth2 flow used for communicating with admin.ClientSecret,Pkce,ExternalCommand are valid values

**Default Value**: 

.. code-block:: yaml

  ClientSecret
  

tokenRefreshWindow (`config.Duration`_)
------------------------------------------------------------------------------------------------------------------------

Max duration between token refresh attempt and token expiry.

**Default Value**: 

.. code-block:: yaml

  0s
  

useAuth (bool)
------------------------------------------------------------------------------------------------------------------------

Deprecated: Auth will be enabled/disabled based on admin's dynamically discovered information.

**Default Value**: 

.. code-block:: yaml

  "false"
  

clientId (string)
------------------------------------------------------------------------------------------------------------------------

Client ID

**Default Value**: 

.. code-block:: yaml

  flytepropeller
  

clientSecretLocation (string)
------------------------------------------------------------------------------------------------------------------------

File containing the client secret

**Default Value**: 

.. code-block:: yaml

  /etc/secrets/client_secret
  

clientSecretEnvVar (string)
------------------------------------------------------------------------------------------------------------------------

Environment variable containing the client secret

**Default Value**: 

.. code-block:: yaml

  ""
  

scopes ([]string)
------------------------------------------------------------------------------------------------------------------------

List of scopes to request

**Default Value**: 

.. code-block:: yaml

  []
  

useAudienceFromAdmin (bool)
------------------------------------------------------------------------------------------------------------------------

Use Audience configured from admins public endpoint config.

**Default Value**: 

.. code-block:: yaml

  "false"
  

audience (string)
------------------------------------------------------------------------------------------------------------------------

Audience to use when initiating OAuth2 authorization requests.

**Default Value**: 

.. code-block:: yaml

  ""
  

authorizationServerUrl (string)
------------------------------------------------------------------------------------------------------------------------

This is the URL to your IdP's authorization server. It'll default to Endpoint

**Default Value**: 

.. code-block:: yaml

  ""
  

tokenUrl (string)
------------------------------------------------------------------------------------------------------------------------

OPTIONAL: Your IdP's token endpoint. It'll be discovered from flyte admin's OAuth Metadata endpoint if not provided.

**Default Value**: 

.. code-block:: yaml

  ""
  

authorizationHeader (string)
------------------------------------------------------------------------------------------------------------------------

Custom metadata header to pass JWT

**Default Value**: 

.. code-block:: yaml

  ""
  

pkceConfig (`pkce.Config`_)
------------------------------------------------------------------------------------------------------------------------

Config for Pkce authentication flow.

**Default Value**: 

.. code-block:: yaml

  refreshTime: 5m0s
  timeout: 2m0s
  

deviceFlowConfig (`deviceflow.Config`_)
------------------------------------------------------------------------------------------------------------------------

Config for Device authentication flow.

**Default Value**: 

.. code-block:: yaml

  pollInterval: 5s
  refreshTime: 5m0s
  timeout: 10m0s
  

command ([]string)
------------------------------------------------------------------------------------------------------------------------

Command for external authentication token generation

**Default Value**: 

.. code-block:: yaml

  []
  

proxyCommand ([]string)
------------------------------------------------------------------------------------------------------------------------

Command for external proxy-authorization token generation

**Default Value**: 

.. code-block:: yaml

  []
  

defaultServiceConfig (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

httpProxyURL (`config.URL`_)
------------------------------------------------------------------------------------------------------------------------

OPTIONAL: HTTP Proxy to be used for OAuth requests.

**Default Value**: 

.. code-block:: yaml

  ""
  

config.Duration
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Duration (int64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  8s
  

config.URL
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

URL (`url.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ForceQuery: false
  Fragment: ""
  Host: ""
  OmitHost: false
  Opaque: ""
  Path: ""
  RawFragment: ""
  RawPath: ""
  RawQuery: ""
  Scheme: ""
  User: null
  

url.URL
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Scheme (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Opaque (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

User (url.Userinfo)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

Host (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Path (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

RawPath (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

OmitHost (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "false"
  

ForceQuery (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "false"
  

RawQuery (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Fragment (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

RawFragment (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

deviceflow.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

refreshTime (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

grace period from the token expiry after which it would refresh the token.

**Default Value**: 

.. code-block:: yaml

  5m0s
  

timeout (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

amount of time the device flow should complete or else it will be cancelled.

**Default Value**: 

.. code-block:: yaml

  10m0s
  

pollInterval (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

amount of time the device flow would poll the token endpoint if auth server doesn't return a polling interval. Okta and google IDP do return an interval'

**Default Value**: 

.. code-block:: yaml

  5s
  

pkce.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

timeout (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Amount of time the browser session would be active for authentication from client app.

**Default Value**: 

.. code-block:: yaml

  2m0s
  

refreshTime (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

grace period from the token expiry after which it would refresh the token.

**Default Value**: 

.. code-block:: yaml

  5m0s
  

Section: auth
========================================================================================================================

httpAuthorizationHeader (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  flyte-authorization
  

grpcAuthorizationHeader (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  flyte-authorization
  

disableForHttp (bool)
------------------------------------------------------------------------------------------------------------------------

Disables auth enforcement on HTTP Endpoints.

**Default Value**: 

.. code-block:: yaml

  "false"
  

disableForGrpc (bool)
------------------------------------------------------------------------------------------------------------------------

Disables auth enforcement on Grpc Endpoints.

**Default Value**: 

.. code-block:: yaml

  "false"
  

authorizedUris ([]config.URL)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  null
  

httpProxyURL (`config.URL`_)
------------------------------------------------------------------------------------------------------------------------

OPTIONAL: HTTP Proxy to be used for OAuth requests.

**Default Value**: 

.. code-block:: yaml

  ""
  

userAuth (`config.UserAuthConfig`_)
------------------------------------------------------------------------------------------------------------------------

Defines Auth options for users.

**Default Value**: 

.. code-block:: yaml

  cookieBlockKeySecretName: cookie_block_key
  cookieHashKeySecretName: cookie_hash_key
  cookieSetting:
    domain: ""
    sameSitePolicy: DefaultMode
  httpProxyURL: ""
  openId:
    baseUrl: ""
    clientId: ""
    clientSecretFile: ""
    clientSecretName: oidc_client_secret
    scopes:
    - openid
    - profile
  redirectUrl: /console
  

appAuth (`config.OAuth2Options`_)
------------------------------------------------------------------------------------------------------------------------

Defines Auth options for apps. UserAuth must be enabled for AppAuth to work.

**Default Value**: 

.. code-block:: yaml

  authServerType: Self
  externalAuthServer:
    allowedAudience: []
    baseUrl: ""
    httpProxyURL: ""
    metadataUrl: ""
    retryAttempts: 5
    retryDelay: 1s
  selfAuthServer:
    accessTokenLifespan: 30m0s
    authorizationCodeLifespan: 5m0s
    claimSymmetricEncryptionKeySecretName: claim_symmetric_key
    issuer: ""
    oldTokenSigningRSAKeySecretName: token_rsa_key_old.pem
    refreshTokenLifespan: 1h0m0s
    staticClients:
      flyte-cli:
        audience: null
        grant_types:
        - refresh_token
        - authorization_code
        id: flyte-cli
        public: true
        redirect_uris:
        - http://localhost:53593/callback
        - http://localhost:12345/callback
        response_types:
        - code
        - token
        scopes:
        - all
        - offline
        - access_token
      flytectl:
        audience: null
        grant_types:
        - refresh_token
        - authorization_code
        id: flytectl
        public: true
        redirect_uris:
        - http://localhost:53593/callback
        - http://localhost:12345/callback
        response_types:
        - code
        - token
        scopes:
        - all
        - offline
        - access_token
      flytepropeller:
        audience: null
        client_secret: JDJhJDA2JGQ2UFFuMlFBRlUzY0w1VjhNRGtldXVrNjN4dWJxVXhOeGp0ZlB3LkZjOU1nVjZ2cG15T0l5
        grant_types:
        - refresh_token
        - client_credentials
        id: flytepropeller
        public: false
        redirect_uris:
        - http://localhost:3846/callback
        response_types:
        - token
        scopes:
        - all
        - offline
        - access_token
    tokenSigningRSAKeySecretName: token_rsa_key.pem
  thirdPartyConfig:
    flyteClient:
      audience: ""
      clientId: flytectl
      redirectUri: http://localhost:53593/callback
      scopes:
      - all
      - offline
  

config.OAuth2Options
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

authServerType (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  Self
  

selfAuthServer (`config.AuthorizationServer`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Authorization Server config to run as a service. Use this when using an IdP that does not offer a custom OAuth2 Authorization Server.

**Default Value**: 

.. code-block:: yaml

  accessTokenLifespan: 30m0s
  authorizationCodeLifespan: 5m0s
  claimSymmetricEncryptionKeySecretName: claim_symmetric_key
  issuer: ""
  oldTokenSigningRSAKeySecretName: token_rsa_key_old.pem
  refreshTokenLifespan: 1h0m0s
  staticClients:
    flyte-cli:
      audience: null
      grant_types:
      - refresh_token
      - authorization_code
      id: flyte-cli
      public: true
      redirect_uris:
      - http://localhost:53593/callback
      - http://localhost:12345/callback
      response_types:
      - code
      - token
      scopes:
      - all
      - offline
      - access_token
    flytectl:
      audience: null
      grant_types:
      - refresh_token
      - authorization_code
      id: flytectl
      public: true
      redirect_uris:
      - http://localhost:53593/callback
      - http://localhost:12345/callback
      response_types:
      - code
      - token
      scopes:
      - all
      - offline
      - access_token
    flytepropeller:
      audience: null
      client_secret: JDJhJDA2JGQ2UFFuMlFBRlUzY0w1VjhNRGtldXVrNjN4dWJxVXhOeGp0ZlB3LkZjOU1nVjZ2cG15T0l5
      grant_types:
      - refresh_token
      - client_credentials
      id: flytepropeller
      public: false
      redirect_uris:
      - http://localhost:3846/callback
      response_types:
      - token
      scopes:
      - all
      - offline
      - access_token
  tokenSigningRSAKeySecretName: token_rsa_key.pem
  

externalAuthServer (`config.ExternalAuthorizationServer`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

External Authorization Server config.

**Default Value**: 

.. code-block:: yaml

  allowedAudience: []
  baseUrl: ""
  httpProxyURL: ""
  metadataUrl: ""
  retryAttempts: 5
  retryDelay: 1s
  

thirdPartyConfig (`config.ThirdPartyConfigOptions`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines settings to instruct flyte cli tools (and optionally others) on what config to use to setup their client.

**Default Value**: 

.. code-block:: yaml

  flyteClient:
    audience: ""
    clientId: flytectl
    redirectUri: http://localhost:53593/callback
    scopes:
    - all
    - offline
  

config.AuthorizationServer
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

issuer (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the issuer to use when issuing and validating tokens. The default value is https://<requestUri.HostAndPort>/

**Default Value**: 

.. code-block:: yaml

  ""
  

accessTokenLifespan (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the lifespan of issued access tokens.

**Default Value**: 

.. code-block:: yaml

  30m0s
  

refreshTokenLifespan (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the lifespan of issued access tokens.

**Default Value**: 

.. code-block:: yaml

  1h0m0s
  

authorizationCodeLifespan (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the lifespan of issued access tokens.

**Default Value**: 

.. code-block:: yaml

  5m0s
  

claimSymmetricEncryptionKeySecretName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

OPTIONAL: Secret name to use to encrypt claims in authcode token.

**Default Value**: 

.. code-block:: yaml

  claim_symmetric_key
  

tokenSigningRSAKeySecretName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

OPTIONAL: Secret name to use to retrieve RSA Signing Key.

**Default Value**: 

.. code-block:: yaml

  token_rsa_key.pem
  

oldTokenSigningRSAKeySecretName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

OPTIONAL: Secret name to use to retrieve Old RSA Signing Key. This can be useful during key rotation to continue to accept older tokens.

**Default Value**: 

.. code-block:: yaml

  token_rsa_key_old.pem
  

staticClients (map[string]*fosite.DefaultClient)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  flyte-cli:
    audience: null
    grant_types:
    - refresh_token
    - authorization_code
    id: flyte-cli
    public: true
    redirect_uris:
    - http://localhost:53593/callback
    - http://localhost:12345/callback
    response_types:
    - code
    - token
    scopes:
    - all
    - offline
    - access_token
  flytectl:
    audience: null
    grant_types:
    - refresh_token
    - authorization_code
    id: flytectl
    public: true
    redirect_uris:
    - http://localhost:53593/callback
    - http://localhost:12345/callback
    response_types:
    - code
    - token
    scopes:
    - all
    - offline
    - access_token
  flytepropeller:
    audience: null
    client_secret: JDJhJDA2JGQ2UFFuMlFBRlUzY0w1VjhNRGtldXVrNjN4dWJxVXhOeGp0ZlB3LkZjOU1nVjZ2cG15T0l5
    grant_types:
    - refresh_token
    - client_credentials
    id: flytepropeller
    public: false
    redirect_uris:
    - http://localhost:3846/callback
    response_types:
    - token
    scopes:
    - all
    - offline
    - access_token
  

config.ExternalAuthorizationServer
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

baseUrl (`config.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

This should be the base url of the authorization server that you are trying to hit. With Okta for instance, it will look something like https://company.okta.com/oauth2/abcdef123456789/

**Default Value**: 

.. code-block:: yaml

  ""
  

allowedAudience ([]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Optional: A list of allowed audiences. If not provided, the audience is expected to be the public Uri of the service.

**Default Value**: 

.. code-block:: yaml

  []
  

metadataUrl (`config.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Optional: If the server doesn't support /.well-known/oauth-authorization-server, you can set a custom metadata url here.'

**Default Value**: 

.. code-block:: yaml

  ""
  

httpProxyURL (`config.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

OPTIONAL: HTTP Proxy to be used for OAuth requests.

**Default Value**: 

.. code-block:: yaml

  ""
  

retryAttempts (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Optional: The number of attempted retries on a transient failure to get the OAuth metadata

**Default Value**: 

.. code-block:: yaml

  "5"
  

retryDelay (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Optional, Duration to wait between retries

**Default Value**: 

.. code-block:: yaml

  1s
  

config.ThirdPartyConfigOptions
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

flyteClient (`config.FlyteClientConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  audience: ""
  clientId: flytectl
  redirectUri: http://localhost:53593/callback
  scopes:
  - all
  - offline
  

config.FlyteClientConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

clientId (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

public identifier for the app which handles authorization for a Flyte deployment

**Default Value**: 

.. code-block:: yaml

  flytectl
  

redirectUri (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

This is the callback uri registered with the app which handles authorization for a Flyte deployment

**Default Value**: 

.. code-block:: yaml

  http://localhost:53593/callback
  

scopes ([]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Recommended scopes for the client to request.

**Default Value**: 

.. code-block:: yaml

  - all
  - offline
  

audience (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Audience to use when initiating OAuth2 authorization requests.

**Default Value**: 

.. code-block:: yaml

  ""
  

config.UserAuthConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

redirectUrl (`config.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  /console
  

openId (`config.OpenIDOptions`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

OpenID Configuration for User Auth

**Default Value**: 

.. code-block:: yaml

  baseUrl: ""
  clientId: ""
  clientSecretFile: ""
  clientSecretName: oidc_client_secret
  scopes:
  - openid
  - profile
  

httpProxyURL (`config.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

OPTIONAL: HTTP Proxy to be used for OAuth requests.

**Default Value**: 

.. code-block:: yaml

  ""
  

cookieHashKeySecretName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

OPTIONAL: Secret name to use for cookie hash key.

**Default Value**: 

.. code-block:: yaml

  cookie_hash_key
  

cookieBlockKeySecretName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

OPTIONAL: Secret name to use for cookie block key.

**Default Value**: 

.. code-block:: yaml

  cookie_block_key
  

cookieSetting (`config.CookieSettings`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

settings used by cookies created for user auth

**Default Value**: 

.. code-block:: yaml

  domain: ""
  sameSitePolicy: DefaultMode
  

config.CookieSettings
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

sameSitePolicy (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

OPTIONAL: Allows you to declare if your cookie should be restricted to a first-party or same-site context.Wrapper around http.SameSite.

**Default Value**: 

.. code-block:: yaml

  DefaultMode
  

domain (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

OPTIONAL: Allows you to set the domain attribute on the auth cookies.

**Default Value**: 

.. code-block:: yaml

  ""
  

config.OpenIDOptions
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

clientId (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

clientSecretName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  oidc_client_secret
  

clientSecretFile (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

baseUrl (`config.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

scopes ([]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  - openid
  - profile
  

Section: cloudevents
========================================================================================================================

enable (bool)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "false"
  

type (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  local
  

aws (`interfaces.AWSConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  region: ""
  

gcp (`interfaces.GCPConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  projectId: ""
  

kafka (`interfaces.KafkaConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  brokers: null
  version: ""
  

eventsPublisher (`interfaces.EventsPublisherConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  eventTypes: null
  topicName: ""
  

reconnectAttempts (int)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

reconnectDelaySeconds (int)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

cloudEventVersion (uint8)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  v1
  

interfaces.AWSConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

region (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

interfaces.EventsPublisherConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

topicName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

eventTypes ([]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

interfaces.GCPConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

projectId (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

interfaces.KafkaConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

version (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

brokers ([]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

Section: cluster_resources
========================================================================================================================

templatePath (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

templateData (map[string]interfaces.DataSource)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  {}
  

refreshInterval (`config.Duration`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  1m0s
  

customData (map[string]map[string]interfaces.DataSource)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  {}
  

standaloneDeployment (bool)
------------------------------------------------------------------------------------------------------------------------

Whether the cluster resource sync is running in a standalone deployment and should call flyteadmin service endpoints

**Default Value**: 

.. code-block:: yaml

  "false"
  

Section: clusterpools
========================================================================================================================

clusterPoolAssignments (map[string]interfaces.ClusterPoolAssignment)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  {}
  

Section: clusters
========================================================================================================================

clusterConfigs ([]interfaces.ClusterConfig)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  null
  

labelClusterMap (map[string][]interfaces.ClusterEntity)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  null
  

defaultExecutionLabel (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

Section: database
========================================================================================================================

host (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

port (int)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

dbname (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

username (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

password (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

passwordPath (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

options (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

debug (bool)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "false"
  

enableForeignKeyConstraintWhenMigrating (bool)
------------------------------------------------------------------------------------------------------------------------

Whether to enable gorm foreign keys when migrating the db

**Default Value**: 

.. code-block:: yaml

  "false"
  

maxIdleConnections (int)
------------------------------------------------------------------------------------------------------------------------

maxIdleConnections sets the maximum number of connections in the idle connection pool.

**Default Value**: 

.. code-block:: yaml

  "10"
  

maxOpenConnections (int)
------------------------------------------------------------------------------------------------------------------------

maxOpenConnections sets the maximum number of open connections to the database.

**Default Value**: 

.. code-block:: yaml

  "100"
  

connMaxLifeTime (`config.Duration`_)
------------------------------------------------------------------------------------------------------------------------

sets the maximum amount of time a connection may be reused

**Default Value**: 

.. code-block:: yaml

  1h0m0s
  

postgres (`database.PostgresConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  dbname: postgres
  debug: false
  host: localhost
  options: sslmode=disable
  password: postgres
  passwordPath: ""
  port: 30001
  username: postgres
  

sqlite (`database.SQLiteConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  file: ""
  

database.PostgresConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

host (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

The host name of the database server

**Default Value**: 

.. code-block:: yaml

  localhost
  

port (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

The port name of the database server

**Default Value**: 

.. code-block:: yaml

  "30001"
  

dbname (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

The database name

**Default Value**: 

.. code-block:: yaml

  postgres
  

username (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

The database user who is connecting to the server.

**Default Value**: 

.. code-block:: yaml

  postgres
  

password (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

The database password.

**Default Value**: 

.. code-block:: yaml

  postgres
  

passwordPath (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Points to the file containing the database password.

**Default Value**: 

.. code-block:: yaml

  ""
  

options (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

See http://gorm.io/docs/connecting_to_the_database.html for available options passed, in addition to the above.

**Default Value**: 

.. code-block:: yaml

  sslmode=disable
  

debug (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Whether or not to start the database connection with debug mode enabled.

**Default Value**: 

.. code-block:: yaml

  "false"
  

database.SQLiteConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

file (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

The path to the file (existing or new) where the DB should be created / stored. If existing, then this will be re-used, else a new will be created

**Default Value**: 

.. code-block:: yaml

  ""
  

Section: domains
========================================================================================================================

id (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  development
  

name (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  development
  

Section: externalevents
========================================================================================================================

enable (bool)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "false"
  

type (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  local
  

aws (`interfaces.AWSConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  region: ""
  

gcp (`interfaces.GCPConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  projectId: ""
  

eventsPublisher (`interfaces.EventsPublisherConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  eventTypes: null
  topicName: ""
  

reconnectAttempts (int)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

reconnectDelaySeconds (int)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

Section: flyteadmin
========================================================================================================================

roleNameKey (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

metricsScope (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  'flyte:'
  

metricsKeys ([]string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  - project
  - domain
  - wf
  - task
  - phase
  - tasktype
  - runtime_type
  - runtime_version
  - app_name
  

profilerPort (int)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "10254"
  

metadataStoragePrefix ([]string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  - metadata
  - admin
  

eventVersion (int)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "2"
  

asyncEventsBufferSize (int)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "100"
  

maxParallelism (int32)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "25"
  

labels (map[string]string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  null
  

annotations (map[string]string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  null
  

interruptible (bool)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "false"
  

overwriteCache (bool)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "false"
  

assumableIamRole (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

k8sServiceAccount (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

outputLocationPrefix (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

useOffloadedWorkflowClosure (bool)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "false"
  

envs (map[string]string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  null
  

featureGates (`interfaces.FeatureGates`_)
------------------------------------------------------------------------------------------------------------------------

Enable experimental features.

**Default Value**: 

.. code-block:: yaml

  enableArtifacts: false
  

interfaces.FeatureGates
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

enableArtifacts (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Enable artifacts feature.

**Default Value**: 

.. code-block:: yaml

  "false"
  

Section: logger
========================================================================================================================

show-source (bool)
------------------------------------------------------------------------------------------------------------------------

Includes source code location in logs.

**Default Value**: 

.. code-block:: yaml

  "false"
  

mute (bool)
------------------------------------------------------------------------------------------------------------------------

Mutes all logs regardless of severity. Intended for benchmarks/tests only.

**Default Value**: 

.. code-block:: yaml

  "false"
  

level (int)
------------------------------------------------------------------------------------------------------------------------

Sets the minimum logging level.

**Default Value**: 

.. code-block:: yaml

  "3"
  

formatter (`logger.FormatterConfig`_)
------------------------------------------------------------------------------------------------------------------------

Sets logging format.

**Default Value**: 

.. code-block:: yaml

  type: json
  

logger.FormatterConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

type (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Sets logging format type.

**Default Value**: 

.. code-block:: yaml

  json
  

Section: namespace_mapping
========================================================================================================================

mapping (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

template (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  '{{ project }}-{{ domain }}'
  

templateData (map[string]interfaces.DataSource)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  null
  

Section: notifications
========================================================================================================================

type (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  local
  

region (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

aws (`interfaces.AWSConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  region: ""
  

gcp (`interfaces.GCPConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  projectId: ""
  

publisher (`interfaces.NotificationsPublisherConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  topicName: ""
  

processor (`interfaces.NotificationsProcessorConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  accountId: ""
  queueName: ""
  

emailer (`interfaces.NotificationsEmailerConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  body: ""
  emailServerConfig:
    apiKeyEnvVar: ""
    apiKeyFilePath: ""
    serviceName: ""
  sender: ""
  subject: ""
  

reconnectAttempts (int)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

reconnectDelaySeconds (int)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

interfaces.NotificationsEmailerConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

emailServerConfig (`interfaces.EmailServerConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  apiKeyEnvVar: ""
  apiKeyFilePath: ""
  serviceName: ""
  

subject (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

sender (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

body (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

interfaces.EmailServerConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

serviceName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

apiKeyEnvVar (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

apiKeyFilePath (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

interfaces.NotificationsProcessorConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

queueName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

accountId (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

interfaces.NotificationsPublisherConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

topicName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Section: otel
========================================================================================================================

type (string)
------------------------------------------------------------------------------------------------------------------------

Sets the type of exporter to configure [noop/file/jaeger].

**Default Value**: 

.. code-block:: yaml

  noop
  

file (`otelutils.FileConfig`_)
------------------------------------------------------------------------------------------------------------------------

Configuration for exporting telemetry traces to a file

**Default Value**: 

.. code-block:: yaml

  filename: /tmp/trace.txt
  

jaeger (`otelutils.JaegerConfig`_)
------------------------------------------------------------------------------------------------------------------------

Configuration for exporting telemetry traces to a jaeger

**Default Value**: 

.. code-block:: yaml

  endpoint: http://localhost:14268/api/traces
  

otelutils.FileConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

filename (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Filename to store exported telemetry traces

**Default Value**: 

.. code-block:: yaml

  /tmp/trace.txt
  

otelutils.JaegerConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

endpoint (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Endpoint for the jaeger telemtry trace ingestor

**Default Value**: 

.. code-block:: yaml

  http://localhost:14268/api/traces
  

Section: plugins
========================================================================================================================

catalogcache (`catalog.Config`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  reader:
    maxItems: 10000
    maxRetries: 3
    workers: 10
  writer:
    maxItems: 10000
    maxRetries: 3
    workers: 10
  

k8s (`config.K8sPluginConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  co-pilot:
    cpu: 500m
    default-input-path: /var/flyte/inputs
    default-output-path: /var/flyte/outputs
    image: cr.flyte.org/flyteorg/flytecopilot:v0.0.15
    input-vol-name: flyte-inputs
    memory: 128Mi
    name: flyte-copilot-
    output-vol-name: flyte-outputs
    start-timeout: 1m40s
    storage: ""
  create-container-config-error-grace-period: 0s
  create-container-error-grace-period: 3m0s
  default-annotations:
    cluster-autoscaler.kubernetes.io/safe-to-evict: "false"
  default-cpus: "1"
  default-env-vars: null
  default-env-vars-from-env: null
  default-labels: null
  default-memory: 1Gi
  default-node-selector: null
  default-pod-dns-config: null
  default-pod-security-context: null
  default-pod-template-name: ""
  default-pod-template-resync: 30s
  default-security-context: null
  default-tolerations: null
  delete-resource-on-finalize: false
  enable-host-networking-pod: null
  gpu-device-node-label: k8s.amazonaws.com/accelerator
  gpu-partition-size-node-label: k8s.amazonaws.com/gpu-partition-size
  gpu-resource-name: nvidia.com/gpu
  gpu-unpartitioned-node-selector-requirement: null
  gpu-unpartitioned-toleration: null
  image-pull-backoff-grace-period: 3m0s
  inject-finalizer: false
  interruptible-node-selector: null
  interruptible-node-selector-requirement: null
  interruptible-tolerations: null
  non-interruptible-node-selector-requirement: null
  pod-pending-timeout: 0s
  resource-tolerations: null
  scheduler-name: ""
  send-object-events: false
  

catalog.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

reader (`workqueue.Config`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Catalog reader workqueue config. Make sure the index cache must be big enough to accommodate the biggest array task allowed to run on the system.

**Default Value**: 

.. code-block:: yaml

  maxItems: 10000
  maxRetries: 3
  workers: 10
  

writer (`workqueue.Config`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Catalog writer workqueue config. Make sure the index cache must be big enough to accommodate the biggest array task allowed to run on the system.

**Default Value**: 

.. code-block:: yaml

  maxItems: 10000
  maxRetries: 3
  workers: 10
  

workqueue.Config
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

workers (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Number of concurrent workers to start processing the queue.

**Default Value**: 

.. code-block:: yaml

  "10"
  

maxRetries (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum number of retries per item.

**Default Value**: 

.. code-block:: yaml

  "3"
  

maxItems (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum number of entries to keep in the index.

**Default Value**: 

.. code-block:: yaml

  "10000"
  

config.K8sPluginConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

inject-finalizer (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Instructs the plugin to inject a finalizer on startTask and remove it on task termination.

**Default Value**: 

.. code-block:: yaml

  "false"
  

default-annotations (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  cluster-autoscaler.kubernetes.io/safe-to-evict: "false"
  

default-labels (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

default-env-vars (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

default-env-vars-from-env (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

default-cpus (`resource.Quantity`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines a default value for cpu for containers if not specified.

**Default Value**: 

.. code-block:: yaml

  "1"
  

default-memory (`resource.Quantity`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines a default value for memory for containers if not specified.

**Default Value**: 

.. code-block:: yaml

  1Gi
  

default-tolerations ([]v1.Toleration)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

default-node-selector (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

default-affinity (v1.Affinity)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

scheduler-name (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines scheduler name.

**Default Value**: 

.. code-block:: yaml

  ""
  

interruptible-tolerations ([]v1.Toleration)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

interruptible-node-selector (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

interruptible-node-selector-requirement (v1.NodeSelectorRequirement)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

non-interruptible-node-selector-requirement (v1.NodeSelectorRequirement)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

resource-tolerations (map[v1.ResourceName][]v1.Toleration)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

co-pilot (`config.FlyteCoPilotConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Co-Pilot Configuration

**Default Value**: 

.. code-block:: yaml

  cpu: 500m
  default-input-path: /var/flyte/inputs
  default-output-path: /var/flyte/outputs
  image: cr.flyte.org/flyteorg/flytecopilot:v0.0.15
  input-vol-name: flyte-inputs
  memory: 128Mi
  name: flyte-copilot-
  output-vol-name: flyte-outputs
  start-timeout: 1m40s
  storage: ""
  

delete-resource-on-finalize (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Instructs the system to delete the resource upon successful execution of a k8s pod rather than have the k8s garbage collector clean it up. This ensures that no resources are kept around (potentially consuming cluster resources). This, however, will cause k8s log links to expire as soon as the resource is finalized.

**Default Value**: 

.. code-block:: yaml

  "false"
  

create-container-error-grace-period (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  3m0s
  

create-container-config-error-grace-period (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  0s
  

image-pull-backoff-grace-period (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  3m0s
  

pod-pending-timeout (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  0s
  

gpu-device-node-label (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  k8s.amazonaws.com/accelerator
  

gpu-partition-size-node-label (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  k8s.amazonaws.com/gpu-partition-size
  

gpu-unpartitioned-node-selector-requirement (v1.NodeSelectorRequirement)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

gpu-unpartitioned-toleration (v1.Toleration)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

gpu-resource-name (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  nvidia.com/gpu
  

default-pod-security-context (v1.PodSecurityContext)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

default-security-context (v1.SecurityContext)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

enable-host-networking-pod (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  <invalid reflect.Value>
  

default-pod-dns-config (v1.PodDNSConfig)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

default-pod-template-name (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Name of the PodTemplate to use as the base for all k8s pods created by FlytePropeller.

**Default Value**: 

.. code-block:: yaml

  ""
  

default-pod-template-resync (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Frequency of resyncing default pod templates

**Default Value**: 

.. code-block:: yaml

  30s
  

send-object-events (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

If true, will send k8s object events in TaskExecutionEvent updates.

**Default Value**: 

.. code-block:: yaml

  "false"
  

config.FlyteCoPilotConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

name (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Flyte co-pilot sidecar container name prefix. (additional bits will be added after this)

**Default Value**: 

.. code-block:: yaml

  flyte-copilot-
  

image (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Flyte co-pilot Docker Image FQN

**Default Value**: 

.. code-block:: yaml

  cr.flyte.org/flyteorg/flytecopilot:v0.0.15
  

default-input-path (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Default path where the volume should be mounted

**Default Value**: 

.. code-block:: yaml

  /var/flyte/inputs
  

default-output-path (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Default path where the volume should be mounted

**Default Value**: 

.. code-block:: yaml

  /var/flyte/outputs
  

input-vol-name (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Name of the data volume that is created for storing inputs

**Default Value**: 

.. code-block:: yaml

  flyte-inputs
  

output-vol-name (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Name of the data volume that is created for storing outputs

**Default Value**: 

.. code-block:: yaml

  flyte-outputs
  

start-timeout (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  1m40s
  

cpu (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Used to set cpu for co-pilot containers

**Default Value**: 

.. code-block:: yaml

  500m
  

memory (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Used to set memory for co-pilot containers

**Default Value**: 

.. code-block:: yaml

  128Mi
  

storage (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Default storage limit for individual inputs / outputs

**Default Value**: 

.. code-block:: yaml

  ""
  

resource.Quantity
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

i (`resource.int64Amount`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  {}
  

d (`resource.infDecAmount`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  <nil>
  

s (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "1"
  

Format (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  DecimalSI
  

resource.infDecAmount
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Dec (inf.Dec)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

resource.int64Amount
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

value (int64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "1"
  

scale (int32)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

Section: propeller
========================================================================================================================

kube-config (string)
------------------------------------------------------------------------------------------------------------------------

Path to kubernetes client config file.

**Default Value**: 

.. code-block:: yaml

  ""
  

master (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

workers (int)
------------------------------------------------------------------------------------------------------------------------

Number of threads to process workflows

**Default Value**: 

.. code-block:: yaml

  "20"
  

workflow-reeval-duration (`config.Duration`_)
------------------------------------------------------------------------------------------------------------------------

Frequency of re-evaluating workflows

**Default Value**: 

.. code-block:: yaml

  10s
  

downstream-eval-duration (`config.Duration`_)
------------------------------------------------------------------------------------------------------------------------

Frequency of re-evaluating downstream tasks

**Default Value**: 

.. code-block:: yaml

  30s
  

limit-namespace (string)
------------------------------------------------------------------------------------------------------------------------

Namespaces to watch for this propeller

**Default Value**: 

.. code-block:: yaml

  all
  

prof-port (`config.Port`_)
------------------------------------------------------------------------------------------------------------------------

Profiler port

**Default Value**: 

.. code-block:: yaml

  10254
  

metadata-prefix (string)
------------------------------------------------------------------------------------------------------------------------

MetadataPrefix should be used if all the metadata for Flyte executions should be stored under a specific prefix in CloudStorage. If not specified, the data will be stored in the base container directly.

**Default Value**: 

.. code-block:: yaml

  metadata/propeller
  

rawoutput-prefix (string)
------------------------------------------------------------------------------------------------------------------------

a fully qualified storage path of the form s3://flyte/abc/..., where all data sandboxes should be stored.

**Default Value**: 

.. code-block:: yaml

  ""
  

queue (`config.CompositeQueueConfig`_)
------------------------------------------------------------------------------------------------------------------------

Workflow workqueue configuration, affects the way the work is consumed from the queue.

**Default Value**: 

.. code-block:: yaml

  batch-size: -1
  batching-interval: 1s
  queue:
    base-delay: 0s
    capacity: 10000
    max-delay: 1m0s
    rate: 1000
    type: maxof
  sub-queue:
    base-delay: 0s
    capacity: 10000
    max-delay: 0s
    rate: 1000
    type: bucket
  type: batch
  

metrics-prefix (string)
------------------------------------------------------------------------------------------------------------------------

An optional prefix for all published metrics.

**Default Value**: 

.. code-block:: yaml

  flyte
  

metrics-keys ([]string)
------------------------------------------------------------------------------------------------------------------------

Metrics labels applied to prometheus metrics emitted by the service.

**Default Value**: 

.. code-block:: yaml

  - project
  - domain
  - wf
  - task
  

enable-admin-launcher (bool)
------------------------------------------------------------------------------------------------------------------------

Enable remote Workflow launcher to Admin

**Default Value**: 

.. code-block:: yaml

  "true"
  

max-workflow-retries (int)
------------------------------------------------------------------------------------------------------------------------

Maximum number of retries per workflow

**Default Value**: 

.. code-block:: yaml

  "10"
  

max-ttl-hours (int)
------------------------------------------------------------------------------------------------------------------------

Maximum number of hours a completed workflow should be retained. Number between 1-23 hours

**Default Value**: 

.. code-block:: yaml

  "23"
  

gc-interval (`config.Duration`_)
------------------------------------------------------------------------------------------------------------------------

Run periodic GC every 30 minutes

**Default Value**: 

.. code-block:: yaml

  30m0s
  

leader-election (`config.LeaderElectionConfig`_)
------------------------------------------------------------------------------------------------------------------------

Config for leader election.

**Default Value**: 

.. code-block:: yaml

  enabled: false
  lease-duration: 15s
  lock-config-map:
    Name: ""
    Namespace: ""
  renew-deadline: 10s
  retry-period: 2s
  

publish-k8s-events (bool)
------------------------------------------------------------------------------------------------------------------------

Enable events publishing to K8s events API.

**Default Value**: 

.. code-block:: yaml

  "false"
  

max-output-size-bytes (int64)
------------------------------------------------------------------------------------------------------------------------

Maximum size of outputs per task

**Default Value**: 

.. code-block:: yaml

  "10485760"
  

enable-grpc-latency-metrics (bool)
------------------------------------------------------------------------------------------------------------------------

Enable grpc latency metrics. Note Histograms metrics can be expensive on Prometheus servers.

**Default Value**: 

.. code-block:: yaml

  "false"
  

kube-client-config (`config.KubeClientConfig`_)
------------------------------------------------------------------------------------------------------------------------

Configuration to control the Kubernetes client

**Default Value**: 

.. code-block:: yaml

  burst: 25
  qps: 100
  timeout: 30s
  

node-config (`config.NodeConfig`_)
------------------------------------------------------------------------------------------------------------------------

config for a workflow node

**Default Value**: 

.. code-block:: yaml

  default-deadlines:
    node-active-deadline: 0s
    node-execution-deadline: 0s
    workflow-active-deadline: 0s
  default-max-attempts: 1
  enable-cr-debug-metadata: false
  ignore-retry-cause: false
  interruptible-failure-threshold: -1
  max-node-retries-system-failures: 3
  

max-streak-length (int)
------------------------------------------------------------------------------------------------------------------------

Maximum number of consecutive rounds that one propeller worker can use for one workflow - >1 => turbo-mode is enabled.

**Default Value**: 

.. code-block:: yaml

  "8"
  

event-config (`config.EventConfig`_)
------------------------------------------------------------------------------------------------------------------------

Configures execution event behavior.

**Default Value**: 

.. code-block:: yaml

  fallback-to-output-reference: false
  raw-output-policy: reference
  

include-shard-key-label ([]string)
------------------------------------------------------------------------------------------------------------------------

Include the specified shard key label in the k8s FlyteWorkflow CRD label selector

**Default Value**: 

.. code-block:: yaml

  []
  

exclude-shard-key-label ([]string)
------------------------------------------------------------------------------------------------------------------------

Exclude the specified shard key label from the k8s FlyteWorkflow CRD label selector

**Default Value**: 

.. code-block:: yaml

  []
  

include-project-label ([]string)
------------------------------------------------------------------------------------------------------------------------

Include the specified project label in the k8s FlyteWorkflow CRD label selector

**Default Value**: 

.. code-block:: yaml

  []
  

exclude-project-label ([]string)
------------------------------------------------------------------------------------------------------------------------

Exclude the specified project label from the k8s FlyteWorkflow CRD label selector

**Default Value**: 

.. code-block:: yaml

  []
  

include-domain-label ([]string)
------------------------------------------------------------------------------------------------------------------------

Include the specified domain label in the k8s FlyteWorkflow CRD label selector

**Default Value**: 

.. code-block:: yaml

  []
  

exclude-domain-label ([]string)
------------------------------------------------------------------------------------------------------------------------

Exclude the specified domain label from the k8s FlyteWorkflow CRD label selector

**Default Value**: 

.. code-block:: yaml

  []
  

cluster-id (string)
------------------------------------------------------------------------------------------------------------------------

Unique cluster id running this flytepropeller instance with which to annotate execution events

**Default Value**: 

.. code-block:: yaml

  propeller
  

create-flyteworkflow-crd (bool)
------------------------------------------------------------------------------------------------------------------------

Enable creation of the FlyteWorkflow CRD on startup

**Default Value**: 

.. code-block:: yaml

  "false"
  

array-node-event-version (int)
------------------------------------------------------------------------------------------------------------------------

ArrayNode eventing version. 0 => legacy (drop-in replacement for maptask), 1 => new

**Default Value**: 

.. code-block:: yaml

  "0"
  

node-execution-worker-count (int)
------------------------------------------------------------------------------------------------------------------------

Number of workers to evaluate node executions, currently only used for array nodes

**Default Value**: 

.. code-block:: yaml

  "8"
  

config.CompositeQueueConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

type (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Type of composite queue to use for the WorkQueue

**Default Value**: 

.. code-block:: yaml

  batch
  

queue (`config.WorkqueueConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Workflow workqueue configuration, affects the way the work is consumed from the queue.

**Default Value**: 

.. code-block:: yaml

  base-delay: 0s
  capacity: 10000
  max-delay: 1m0s
  rate: 1000
  type: maxof
  

sub-queue (`config.WorkqueueConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

SubQueue configuration, affects the way the nodes cause the top-level Work to be re-evaluated.

**Default Value**: 

.. code-block:: yaml

  base-delay: 0s
  capacity: 10000
  max-delay: 0s
  rate: 1000
  type: bucket
  

batching-interval (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Duration for which downstream updates are buffered

**Default Value**: 

.. code-block:: yaml

  1s
  

batch-size (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "-1"
  

config.WorkqueueConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

type (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Type of RateLimiter to use for the WorkQueue

**Default Value**: 

.. code-block:: yaml

  maxof
  

base-delay (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

base backoff delay for failure

**Default Value**: 

.. code-block:: yaml

  0s
  

max-delay (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Max backoff delay for failure

**Default Value**: 

.. code-block:: yaml

  1m0s
  

rate (int64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Bucket Refill rate per second

**Default Value**: 

.. code-block:: yaml

  "1000"
  

capacity (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Bucket capacity as number of items

**Default Value**: 

.. code-block:: yaml

  "10000"
  

config.EventConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

raw-output-policy (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

How output data should be passed along in execution events.

**Default Value**: 

.. code-block:: yaml

  reference
  

fallback-to-output-reference (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Whether output data should be sent by reference when it is too large to be sent inline in execution events.

**Default Value**: 

.. code-block:: yaml

  "false"
  

config.KubeClientConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

qps (float32)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "100"
  

burst (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Max burst rate for throttle. 0 defaults to 10

**Default Value**: 

.. code-block:: yaml

  "25"
  

timeout (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Max duration allowed for every request to KubeAPI before giving up. 0 implies no timeout.

**Default Value**: 

.. code-block:: yaml

  30s
  

config.LeaderElectionConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

enabled (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Enables/Disables leader election.

**Default Value**: 

.. code-block:: yaml

  "false"
  

lock-config-map (`types.NamespacedName`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

ConfigMap namespace/name to use for resource lock.

**Default Value**: 

.. code-block:: yaml

  Name: ""
  Namespace: ""
  

lease-duration (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Duration that non-leader candidates will wait to force acquire leadership. This is measured against time of last observed ack.

**Default Value**: 

.. code-block:: yaml

  15s
  

renew-deadline (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Duration that the acting master will retry refreshing leadership before giving up.

**Default Value**: 

.. code-block:: yaml

  10s
  

retry-period (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Duration the LeaderElector clients should wait between tries of actions.

**Default Value**: 

.. code-block:: yaml

  2s
  

types.NamespacedName
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Namespace (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Name (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

config.NodeConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

default-deadlines (`config.DefaultDeadlines`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Default value for timeouts

**Default Value**: 

.. code-block:: yaml

  node-active-deadline: 0s
  node-execution-deadline: 0s
  workflow-active-deadline: 0s
  

max-node-retries-system-failures (int64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum number of retries per node for node failure due to infra issues

**Default Value**: 

.. code-block:: yaml

  "3"
  

interruptible-failure-threshold (int32)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

number of failures for a node to be still considered interruptible. Negative numbers are treated as complementary (ex. -1 means last attempt is non-interruptible).'

**Default Value**: 

.. code-block:: yaml

  "-1"
  

default-max-attempts (int32)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Default maximum number of attempts for a node

**Default Value**: 

.. code-block:: yaml

  "1"
  

ignore-retry-cause (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Ignore retry cause and count all attempts toward a node's max attempts

**Default Value**: 

.. code-block:: yaml

  "false"
  

enable-cr-debug-metadata (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Collapse node on any terminal state, not just successful terminations. This is useful to reduce the size of workflow state in etcd.

**Default Value**: 

.. code-block:: yaml

  "false"
  

config.DefaultDeadlines
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

node-execution-deadline (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Default value of node execution timeout that includes the time spent to run the node/workflow

**Default Value**: 

.. code-block:: yaml

  0s
  

node-active-deadline (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Default value of node timeout that includes the time spent queued.

**Default Value**: 

.. code-block:: yaml

  0s
  

workflow-active-deadline (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Default value of workflow timeout that includes the time spent queued.

**Default Value**: 

.. code-block:: yaml

  0s
  

config.Port
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

port (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "10254"
  

Section: qualityofservice
========================================================================================================================

tierExecutionValues (map[string]interfaces.QualityOfServiceSpec)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  {}
  

defaultTiers (map[string]string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  {}
  

Section: queues
========================================================================================================================

executionQueues (interfaces.ExecutionQueues)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  []
  

workflowConfigs (interfaces.WorkflowConfigs)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  []
  

Section: registration
========================================================================================================================

maxWorkflowNodes (int)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "100"
  

maxLabelEntries (int)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

maxAnnotationEntries (int)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

workflowSizeLimit (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

Section: remotedata
========================================================================================================================

scheme (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  none
  

region (string)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

signedUrls (`interfaces.SignedURL`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  durationMinutes: 0
  enabled: false
  signingPrincipal: ""
  

maxSizeInBytes (int64)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "2097152"
  

inlineEventDataPolicy (int)
------------------------------------------------------------------------------------------------------------------------

Specifies how inline execution event data should be saved in the backend

**Default Value**: 

.. code-block:: yaml

  Offload
  

interfaces.SignedURL
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

enabled (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Whether signed urls should even be returned with GetExecutionData, GetNodeExecutionData and GetTaskExecutionData response objects.

**Default Value**: 

.. code-block:: yaml

  "false"
  

durationMinutes (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

signingPrincipal (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Section: scheduler
========================================================================================================================

profilerPort (`config.Port`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  10254
  

eventScheduler (`interfaces.EventSchedulerConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  aws: null
  local: {}
  region: ""
  scheduleNamePrefix: ""
  scheduleRole: ""
  scheme: local
  targetName: ""
  

workflowExecutor (`interfaces.WorkflowExecutorConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  accountId: ""
  aws: null
  local:
    adminRateLimit:
      burst: 10
      tps: 100
    useUTCTz: false
  region: ""
  scheduleQueueName: ""
  scheme: local
  

reconnectAttempts (int)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

reconnectDelaySeconds (int)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

interfaces.EventSchedulerConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

scheme (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  local
  

region (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

scheduleRole (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

targetName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

scheduleNamePrefix (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

aws (interfaces.AWSSchedulerConfig)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

local (`interfaces.FlyteSchedulerConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  {}
  

interfaces.FlyteSchedulerConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

interfaces.WorkflowExecutorConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

scheme (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  local
  

region (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

scheduleQueueName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

accountId (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

aws (interfaces.AWSWorkflowExecutorConfig)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

local (`interfaces.FlyteWorkflowExecutorConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  adminRateLimit:
    burst: 10
    tps: 100
  useUTCTz: false
  

interfaces.FlyteWorkflowExecutorConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

adminRateLimit (`interfaces.AdminRateLimit`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  burst: 10
  tps: 100
  

useUTCTz (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "false"
  

interfaces.AdminRateLimit
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

tps (float64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "100"
  

burst (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "10"
  

Section: secrets
========================================================================================================================

secrets-prefix (string)
------------------------------------------------------------------------------------------------------------------------

Prefix where to look for secrets file

**Default Value**: 

.. code-block:: yaml

  /etc/secrets
  

env-prefix (string)
------------------------------------------------------------------------------------------------------------------------

Prefix for environment variables

**Default Value**: 

.. code-block:: yaml

  FLYTE_SECRET_
  

Section: server
========================================================================================================================

httpPort (int)
------------------------------------------------------------------------------------------------------------------------

On which http port to serve admin

**Default Value**: 

.. code-block:: yaml

  "8088"
  

grpcPort (int)
------------------------------------------------------------------------------------------------------------------------

deprecated

**Default Value**: 

.. code-block:: yaml

  "0"
  

grpcServerReflection (bool)
------------------------------------------------------------------------------------------------------------------------

deprecated

**Default Value**: 

.. code-block:: yaml

  "false"
  

kube-config (string)
------------------------------------------------------------------------------------------------------------------------

Path to kubernetes client config file, default is empty, useful for incluster config.

**Default Value**: 

.. code-block:: yaml

  ""
  

master (string)
------------------------------------------------------------------------------------------------------------------------

The address of the Kubernetes API server.

**Default Value**: 

.. code-block:: yaml

  ""
  

security (`config.ServerSecurityOptions`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  allowCors: true
  allowedHeaders:
  - Content-Type
  - flyte-authorization
  allowedOrigins:
  - '*'
  auditAccess: false
  secure: false
  ssl:
    certificateFile: ""
    keyFile: ""
  useAuth: false
  

grpc (`config.GrpcConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  enableGrpcLatencyMetrics: false
  maxMessageSizeBytes: 0
  port: 8089
  serverReflection: true
  

thirdPartyConfig (`config.ThirdPartyConfigOptions`_)
------------------------------------------------------------------------------------------------------------------------

Deprecated please use auth.appAuth.thirdPartyConfig instead.

**Default Value**: 

.. code-block:: yaml

  flyteClient:
    audience: ""
    clientId: ""
    redirectUri: ""
    scopes: []
  

dataProxy (`config.DataProxyConfig`_)
------------------------------------------------------------------------------------------------------------------------

Defines data proxy configuration.

**Default Value**: 

.. code-block:: yaml

  download:
    maxExpiresIn: 1h0m0s
  upload:
    defaultFileNameLength: 20
    maxExpiresIn: 1h0m0s
    maxSize: 6Mi
    storagePrefix: ""
  

readHeaderTimeoutSeconds (int)
------------------------------------------------------------------------------------------------------------------------

The amount of time allowed to read request headers.

**Default Value**: 

.. code-block:: yaml

  "32"
  

kubeClientConfig (`config.KubeClientConfig (kubeClientConfig)`_)
------------------------------------------------------------------------------------------------------------------------

Configuration to control the Kubernetes client

**Default Value**: 

.. code-block:: yaml

  burst: 25
  qps: 100
  timeout: 30s
  

config.DataProxyConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

upload (`config.DataProxyUploadConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines data proxy upload configuration.

**Default Value**: 

.. code-block:: yaml

  defaultFileNameLength: 20
  maxExpiresIn: 1h0m0s
  maxSize: 6Mi
  storagePrefix: ""
  

download (`config.DataProxyDownloadConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines data proxy download configuration.

**Default Value**: 

.. code-block:: yaml

  maxExpiresIn: 1h0m0s
  

config.DataProxyDownloadConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

maxExpiresIn (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum allowed expiration duration.

**Default Value**: 

.. code-block:: yaml

  1h0m0s
  

config.DataProxyUploadConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

maxSize (`resource.Quantity`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum allowed upload size.

**Default Value**: 

.. code-block:: yaml

  6Mi
  

maxExpiresIn (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum allowed expiration duration.

**Default Value**: 

.. code-block:: yaml

  1h0m0s
  

defaultFileNameLength (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Default length for the generated file name if not provided in the request.

**Default Value**: 

.. code-block:: yaml

  "20"
  

storagePrefix (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Storage prefix to use for all upload requests.

**Default Value**: 

.. code-block:: yaml

  ""
  

config.GrpcConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

port (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

On which grpc port to serve admin

**Default Value**: 

.. code-block:: yaml

  "8089"
  

serverReflection (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Enable GRPC Server Reflection

**Default Value**: 

.. code-block:: yaml

  "true"
  

maxMessageSizeBytes (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

The max size in bytes for incoming gRPC messages

**Default Value**: 

.. code-block:: yaml

  "0"
  

enableGrpcLatencyMetrics (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Enable grpc latency metrics. Note Histograms metrics can be expensive on Prometheus servers.

**Default Value**: 

.. code-block:: yaml

  "false"
  

config.KubeClientConfig (kubeClientConfig)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

qps (int32)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Max QPS to the master for requests to KubeAPI. 0 defaults to 5.

**Default Value**: 

.. code-block:: yaml

  "100"
  

burst (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Max burst rate for throttle. 0 defaults to 10

**Default Value**: 

.. code-block:: yaml

  "25"
  

timeout (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Max duration allowed for every request to KubeAPI before giving up. 0 implies no timeout.

**Default Value**: 

.. code-block:: yaml

  30s
  

config.ServerSecurityOptions
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

secure (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "false"
  

ssl (`config.SslOptions`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  certificateFile: ""
  keyFile: ""
  

useAuth (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "false"
  

auditAccess (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "false"
  

allowCors (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "true"
  

allowedOrigins ([]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  - '*'
  

allowedHeaders ([]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  - Content-Type
  - flyte-authorization
  

config.SslOptions
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

certificateFile (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

keyFile (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Section: storage
========================================================================================================================

type (string)
------------------------------------------------------------------------------------------------------------------------

Sets the type of storage to configure [s3/minio/local/mem/stow].

**Default Value**: 

.. code-block:: yaml

  s3
  

connection (`storage.ConnectionConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  access-key: ""
  auth-type: iam
  disable-ssl: false
  endpoint: ""
  region: us-east-1
  secret-key: ""
  

stow (`storage.StowConfig`_)
------------------------------------------------------------------------------------------------------------------------

Storage config for stow backend.

**Default Value**: 

.. code-block:: yaml

  {}
  

container (string)
------------------------------------------------------------------------------------------------------------------------

Initial container (in s3 a bucket) to create -if it doesn't exist-.'

**Default Value**: 

.. code-block:: yaml

  ""
  

enable-multicontainer (bool)
------------------------------------------------------------------------------------------------------------------------

If this is true, then the container argument is overlooked and redundant. This config will automatically open new connections to new containers/buckets as they are encountered

**Default Value**: 

.. code-block:: yaml

  "false"
  

cache (`storage.CachingConfig`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  max_size_mbs: 0
  target_gc_percent: 0
  

limits (`storage.LimitsConfig`_)
------------------------------------------------------------------------------------------------------------------------

Sets limits for stores.

**Default Value**: 

.. code-block:: yaml

  maxDownloadMBs: 2
  

defaultHttpClient (`storage.HTTPClientConfig`_)
------------------------------------------------------------------------------------------------------------------------

Sets the default http client config.

**Default Value**: 

.. code-block:: yaml

  headers: null
  timeout: 0s
  

signedUrl (`storage.SignedURLConfig`_)
------------------------------------------------------------------------------------------------------------------------

Sets config for SignedURL.

**Default Value**: 

.. code-block:: yaml

  {}
  

storage.CachingConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

max_size_mbs (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum size of the cache where the Blob store data is cached in-memory. If not specified or set to 0, cache is not used

**Default Value**: 

.. code-block:: yaml

  "0"
  

target_gc_percent (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Sets the garbage collection target percentage.

**Default Value**: 

.. code-block:: yaml

  "0"
  

storage.ConnectionConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

endpoint (`config.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

URL for storage client to connect to.

**Default Value**: 

.. code-block:: yaml

  ""
  

auth-type (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Auth Type to use [iam,accesskey].

**Default Value**: 

.. code-block:: yaml

  iam
  

access-key (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Access key to use. Only required when authtype is set to accesskey.

**Default Value**: 

.. code-block:: yaml

  ""
  

secret-key (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Secret to use when accesskey is set.

**Default Value**: 

.. code-block:: yaml

  ""
  

region (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Region to connect to.

**Default Value**: 

.. code-block:: yaml

  us-east-1
  

disable-ssl (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Disables SSL connection. Should only be used for development.

**Default Value**: 

.. code-block:: yaml

  "false"
  

storage.HTTPClientConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

headers (map[string][]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

timeout (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Sets time out on the http client.

**Default Value**: 

.. code-block:: yaml

  0s
  

storage.LimitsConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

maxDownloadMBs (int64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum allowed download size (in MBs) per call.

**Default Value**: 

.. code-block:: yaml

  "2"
  

storage.SignedURLConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

stowConfigOverride (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

storage.StowConfig
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

kind (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Kind of Stow backend to use. Refer to github/flyteorg/stow

**Default Value**: 

.. code-block:: yaml

  ""
  

config (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Configuration for stow backend. Refer to github/flyteorg/stow

**Default Value**: 

.. code-block:: yaml

  {}
  

Section: task_resources
========================================================================================================================

defaults (`interfaces.TaskResourceSet`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  cpu: "2"
  ephemeralStorage: "0"
  gpu: "0"
  memory: 200Mi
  storage: "0"
  

limits (`interfaces.TaskResourceSet`_)
------------------------------------------------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  cpu: "2"
  ephemeralStorage: "0"
  gpu: "1"
  memory: 1Gi
  storage: "0"
  

interfaces.TaskResourceSet
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

cpu (`resource.Quantity`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "2"
  

gpu (`resource.Quantity`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

memory (`resource.Quantity`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  200Mi
  

storage (`resource.Quantity`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

ephemeralStorage (`resource.Quantity`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

