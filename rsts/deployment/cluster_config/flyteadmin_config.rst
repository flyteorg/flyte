.. _flyteadmin-config-specification:

#########################################
Flyte Admin Configuration
#########################################

- `auth <#section-auth>`_

- `cluster_resources <#section-cluster_resources>`_

- `clusters <#section-clusters>`_

- `database <#section-database>`_

- `domains <#section-domains>`_

- `externalevents <#section-externalevents>`_

- `flyteadmin <#section-flyteadmin>`_

- `logger <#section-logger>`_

- `namespace_mapping <#section-namespace_mapping>`_

- `notifications <#section-notifications>`_

- `plugins <#section-plugins>`_

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

Section: auth
================================================================================

httpAuthorizationHeader (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  flyte-authorization
  

grpcAuthorizationHeader (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  flyte-authorization
  

disableForHttp (bool)
--------------------------------------------------------------------------------

Disables auth enforcement on HTTP Endpoints.

**Default Value**: 

.. code-block:: yaml

  "false"
  

disableForGrpc (bool)
--------------------------------------------------------------------------------

Disables auth enforcement on Grpc Endpoints.

**Default Value**: 

.. code-block:: yaml

  "false"
  

authorizedUris ([]config.URL)
--------------------------------------------------------------------------------

Optional: Defines the set of URIs that clients are allowed to visit the service on. If set, the system will attempt to match the incoming host to the first authorized URIs and use that (including the scheme) when generating metadata endpoints and when validating audience and issuer claims. If not provided, the urls will be deduced based on the request url and the 'secure' setting.

**Default Value**: 

.. code-block:: yaml

  []
  

userAuth (`config.UserAuthConfig`_)
--------------------------------------------------------------------------------

Defines Auth options for users.

**Default Value**: 

.. code-block:: yaml

  cookieBlockKeySecretName: cookie_block_key
  cookieHashKeySecretName: cookie_hash_key
  openId:
    baseUrl: ""
    clientId: ""
    clientSecretFile: ""
    clientSecretName: oidc_client_secret
    scopes: []
  redirectUrl: /console
  

appAuth (`config.OAuth2Options`_)
--------------------------------------------------------------------------------

Defines Auth options for apps. UserAuth must be enabled for AppAuth to work.

**Default Value**: 

.. code-block:: yaml

  authServerType: Self
  externalAuthServer:
    allowedAudience: []
    baseUrl: ""
    metadataUrl: ""
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
        client_secret: JDJhJDA2JHB4czFBa0c4MUt2cmhwbWwxUWlMU09RYVRrOWVlUHJVLzdZYWI5eTA3aDN4MFRnbGJhb1Q2
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
      clientId: flytectl
      redirectUri: http://localhost:53593/callback
      scopes: []
  

config.OAuth2Options
--------------------------------------------------------------------------------

authServerType (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  Self
  

selfAuthServer (`config.AuthorizationServer`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

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
      client_secret: JDJhJDA2JHB4czFBa0c4MUt2cmhwbWwxUWlMU09RYVRrOWVlUHJVLzdZYWI5eTA3aDN4MFRnbGJhb1Q2
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
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

External Authorization Server config.

**Default Value**: 

.. code-block:: yaml

  allowedAudience: []
  baseUrl: ""
  metadataUrl: ""
  

thirdPartyConfig (`config.ThirdPartyConfigOptions`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines settings to instruct flyte cli tools (and optionally others) on what config to use to setup their client.

**Default Value**: 

.. code-block:: yaml

  flyteClient:
    clientId: flytectl
    redirectUri: http://localhost:53593/callback
    scopes: []
  

config.AuthorizationServer
--------------------------------------------------------------------------------

issuer (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the issuer to use when issuing and validating tokens. The default value is https://<requestUri.HostAndPort>/

**Default Value**: 

.. code-block:: yaml

  ""
  

accessTokenLifespan (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the lifespan of issued access tokens.

**Default Value**: 

.. code-block:: yaml

  30m0s
  

refreshTokenLifespan (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the lifespan of issued access tokens.

**Default Value**: 

.. code-block:: yaml

  1h0m0s
  

authorizationCodeLifespan (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Defines the lifespan of issued access tokens.

**Default Value**: 

.. code-block:: yaml

  5m0s
  

claimSymmetricEncryptionKeySecretName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

OPTIONAL: Secret name to use to encrypt claims in authcode token.

**Default Value**: 

.. code-block:: yaml

  claim_symmetric_key
  

tokenSigningRSAKeySecretName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

OPTIONAL: Secret name to use to retrieve RSA Signing Key.

**Default Value**: 

.. code-block:: yaml

  token_rsa_key.pem
  

oldTokenSigningRSAKeySecretName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

OPTIONAL: Secret name to use to retrieve Old RSA Signing Key. This can be useful during key rotation to continue to accept older tokens.

**Default Value**: 

.. code-block:: yaml

  token_rsa_key_old.pem
  

staticClients (map[string]*fosite.DefaultClient)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

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
    client_secret: JDJhJDA2JHB4czFBa0c4MUt2cmhwbWwxUWlMU09RYVRrOWVlUHJVLzdZYWI5eTA3aDN4MFRnbGJhb1Q2
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
  

config.Duration
--------------------------------------------------------------------------------

Duration (int64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  30m0s
  

config.ExternalAuthorizationServer
--------------------------------------------------------------------------------

baseUrl (`config.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

This should be the base url of the authorization server that you are trying to hit. With Okta for instance, it will look something like https://company.okta.com/oauth2/abcdef123456789/

**Default Value**: 

.. code-block:: yaml

  ""
  

allowedAudience ([]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Optional: A list of allowed audiences. If not provided, the audience is expected to be the public Uri of the service.

**Default Value**: 

.. code-block:: yaml

  []
  

metadataUrl (`config.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Optional: If the server doesn't support /.well-known/oauth-authorization-server, you can set a custom metadata url here.'

**Default Value**: 

.. code-block:: yaml

  ""
  

config.URL
--------------------------------------------------------------------------------

URL (`url.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ForceQuery: false
  Fragment: ""
  Host: ""
  Opaque: ""
  Path: ""
  RawFragment: ""
  RawPath: ""
  RawQuery: ""
  Scheme: ""
  User: null
  

url.URL
--------------------------------------------------------------------------------

Scheme (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Opaque (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

User (url.Userinfo)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

Host (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Path (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

RawPath (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

ForceQuery (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "false"
  

RawQuery (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Fragment (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

RawFragment (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

config.ThirdPartyConfigOptions
--------------------------------------------------------------------------------

flyteClient (`config.FlyteClientConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  clientId: flytectl
  redirectUri: http://localhost:53593/callback
  scopes: []
  

config.FlyteClientConfig
--------------------------------------------------------------------------------

clientId (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

public identifier for the app which handles authorization for a Flyte deployment

**Default Value**: 

.. code-block:: yaml

  flytectl
  

redirectUri (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

This is the callback uri registered with the app which handles authorization for a Flyte deployment

**Default Value**: 

.. code-block:: yaml

  http://localhost:53593/callback
  

scopes ([]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Recommended scopes for the client to request.

**Default Value**: 

.. code-block:: yaml

  []
  

config.UserAuthConfig
--------------------------------------------------------------------------------

redirectUrl (`config.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  /console
  

openId (`config.OpenIDOptions`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

OpenID Configuration for User Auth

**Default Value**: 

.. code-block:: yaml

  baseUrl: ""
  clientId: ""
  clientSecretFile: ""
  clientSecretName: oidc_client_secret
  scopes: []
  

cookieHashKeySecretName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

OPTIONAL: Secret name to use for cookie hash key.

**Default Value**: 

.. code-block:: yaml

  cookie_hash_key
  

cookieBlockKeySecretName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

OPTIONAL: Secret name to use for cookie block key.

**Default Value**: 

.. code-block:: yaml

  cookie_block_key
  

config.OpenIDOptions
--------------------------------------------------------------------------------

clientId (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

clientSecretName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  oidc_client_secret
  

clientSecretFile (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

baseUrl (`config.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

scopes ([]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  []
  

Section: cluster_resources
================================================================================

templatePath (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

templateData (map[string]interfaces.DataSource)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  {}
  

refreshInterval (`config.Duration`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  1m0s
  

customData (map[string]map[string]interfaces.DataSource)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  {}
  

Section: clusters
================================================================================

clusterConfigs ([]interfaces.ClusterConfig)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  null
  

labelClusterMap (map[string][]interfaces.ClusterEntity)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  null
  

Section: database
================================================================================

host (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  postgres
  

port (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "5432"
  

dbname (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  postgres
  

username (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  postgres
  

password (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

passwordPath (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

options (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  sslmode=disable
  

debug (bool)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "false"
  

Section: domains
================================================================================

id (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  development
  

name (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  development
  

Section: externalevents
================================================================================

enable (bool)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "false"
  

type (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  local
  

aws (`interfaces.AWSConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  region: ""
  

gcp (`interfaces.GCPConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  projectId: ""
  

eventsPublisher (`interfaces.EventsPublisherConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  eventTypes: null
  topicName: ""
  

reconnectAttempts (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

reconnectDelaySeconds (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

interfaces.AWSConfig
--------------------------------------------------------------------------------

region (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

interfaces.EventsPublisherConfig
--------------------------------------------------------------------------------

topicName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

eventTypes ([]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

interfaces.GCPConfig
--------------------------------------------------------------------------------

projectId (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Section: flyteadmin
================================================================================

roleNameKey (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

metricsScope (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  'flyte:'
  

profilerPort (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "10254"
  

metadataStoragePrefix ([]string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  - metadata
  - admin
  

eventVersion (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "1"
  

asyncEventsBufferSize (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "100"
  

Section: logger
================================================================================

show-source (bool)
--------------------------------------------------------------------------------

Includes source code location in logs.

**Default Value**: 

.. code-block:: yaml

  "false"
  

mute (bool)
--------------------------------------------------------------------------------

Mutes all logs regardless of severity. Intended for benchmarks/tests only.

**Default Value**: 

.. code-block:: yaml

  "false"
  

level (int)
--------------------------------------------------------------------------------

Sets the minimum logging level.

**Default Value**: 

.. code-block:: yaml

  "4"
  

formatter (`logger.FormatterConfig`_)
--------------------------------------------------------------------------------

Sets logging format.

**Default Value**: 

.. code-block:: yaml

  type: json
  

logger.FormatterConfig
--------------------------------------------------------------------------------

type (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Sets logging format type.

**Default Value**: 

.. code-block:: yaml

  json
  

Section: namespace_mapping
================================================================================

mapping (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

template (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  '{{ project }}-{{ domain }}'
  

templateData (map[string]interfaces.DataSource)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  null
  

Section: notifications
================================================================================

type (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  local
  

region (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

aws (`interfaces.AWSConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  region: ""
  

gcp (`interfaces.GCPConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  projectId: ""
  

publisher (`interfaces.NotificationsPublisherConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  topicName: ""
  

processor (`interfaces.NotificationsProcessorConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  accountId: ""
  queueName: ""
  

emailer (`interfaces.NotificationsEmailerConfig`_)
--------------------------------------------------------------------------------

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
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

reconnectDelaySeconds (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

interfaces.NotificationsEmailerConfig
--------------------------------------------------------------------------------

emailServerConfig (`interfaces.EmailServerConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  apiKeyEnvVar: ""
  apiKeyFilePath: ""
  serviceName: ""
  

subject (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

sender (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

body (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

interfaces.EmailServerConfig
--------------------------------------------------------------------------------

serviceName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

apiKeyEnvVar (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

apiKeyFilePath (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

interfaces.NotificationsProcessorConfig
--------------------------------------------------------------------------------

queueName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

accountId (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

interfaces.NotificationsPublisherConfig
--------------------------------------------------------------------------------

topicName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Section: plugins
================================================================================

enabled-plugins ([]string)
--------------------------------------------------------------------------------

List of enabled plugins, default value is to enable all plugins.

**Default Value**: 

.. code-block:: yaml

  - '*'
  

catalogcache (`catalog.Config`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  reader:
    maxItems: 1000
    maxRetries: 3
    workers: 10
  writer:
    maxItems: 1000
    maxRetries: 3
    workers: 10
  

catalog.Config
--------------------------------------------------------------------------------

reader (`workqueue.Config`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Catalog reader workqueue config. Make sure the index cache must be big enough to accommodate the biggest array task allowed to run on the system.

**Default Value**: 

.. code-block:: yaml

  maxItems: 1000
  maxRetries: 3
  workers: 10
  

writer (`workqueue.Config`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Catalog writer workqueue config. Make sure the index cache must be big enough to accommodate the biggest array task allowed to run on the system.

**Default Value**: 

.. code-block:: yaml

  maxItems: 1000
  maxRetries: 3
  workers: 10
  

workqueue.Config
--------------------------------------------------------------------------------

workers (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Number of concurrent workers to start processing the queue.

**Default Value**: 

.. code-block:: yaml

  "10"
  

maxRetries (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum number of retries per item.

**Default Value**: 

.. code-block:: yaml

  "3"
  

maxItems (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum number of entries to keep in the index.

**Default Value**: 

.. code-block:: yaml

  "1000"
  

Section: qualityofservice
================================================================================

tierExecutionValues (map[string]interfaces.QualityOfServiceSpec)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  {}
  

defaultTiers (map[string]string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  {}
  

Section: queues
================================================================================

executionQueues (interfaces.ExecutionQueues)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  []
  

workflowConfigs (interfaces.WorkflowConfigs)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  []
  

Section: registration
================================================================================

maxWorkflowNodes (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "100"
  

maxLabelEntries (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

maxAnnotationEntries (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

workflowSizeLimit (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

Section: remotedata
================================================================================

scheme (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  none
  

region (string)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  ""
  

signedUrls (`interfaces.SignedURL`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  durationMinutes: 0
  signingPrincipal: ""
  

maxSizeInBytes (int64)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "2097152"
  

interfaces.SignedURL
--------------------------------------------------------------------------------

durationMinutes (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

signingPrincipal (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Section: scheduler
================================================================================

eventScheduler (`interfaces.EventSchedulerConfig`_)
--------------------------------------------------------------------------------

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
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  accountId: ""
  aws: null
  local:
    adminRateLimit:
      burst: 10
      tps: 100
  region: ""
  scheduleQueueName: ""
  scheme: local
  

reconnectAttempts (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

reconnectDelaySeconds (int)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  "0"
  

interfaces.EventSchedulerConfig
--------------------------------------------------------------------------------

scheme (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  local
  

region (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

scheduleRole (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

targetName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

scheduleNamePrefix (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

aws (interfaces.AWSSchedulerConfig)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

local (`interfaces.FlyteSchedulerConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  {}
  

interfaces.FlyteSchedulerConfig
--------------------------------------------------------------------------------

interfaces.WorkflowExecutorConfig
--------------------------------------------------------------------------------

scheme (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  local
  

region (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

scheduleQueueName (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

accountId (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

aws (interfaces.AWSWorkflowExecutorConfig)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

local (`interfaces.FlyteWorkflowExecutorConfig`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  adminRateLimit:
    burst: 10
    tps: 100
  

interfaces.FlyteWorkflowExecutorConfig
--------------------------------------------------------------------------------

adminRateLimit (`interfaces.AdminRateLimit`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  burst: 10
  tps: 100
  

interfaces.AdminRateLimit
--------------------------------------------------------------------------------

tps (float64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "100"
  

burst (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "10"
  

Section: secrets
================================================================================

secrets-prefix (string)
--------------------------------------------------------------------------------

Prefix where to look for secrets file

**Default Value**: 

.. code-block:: yaml

  /etc/secrets
  

env-prefix (string)
--------------------------------------------------------------------------------

Prefix for environment variables

**Default Value**: 

.. code-block:: yaml

  FLYTE_SECRET_
  

Section: server
================================================================================

httpPort (int)
--------------------------------------------------------------------------------

On which http port to serve admin

**Default Value**: 

.. code-block:: yaml

  "0"
  

grpcPort (int)
--------------------------------------------------------------------------------

On which grpc port to serve admin

**Default Value**: 

.. code-block:: yaml

  "0"
  

grpcServerReflection (bool)
--------------------------------------------------------------------------------

Enable GRPC Server Reflection

**Default Value**: 

.. code-block:: yaml

  "false"
  

kube-config (string)
--------------------------------------------------------------------------------

Path to kubernetes client config file.

**Default Value**: 

.. code-block:: yaml

  ""
  

master (string)
--------------------------------------------------------------------------------

The address of the Kubernetes API server.

**Default Value**: 

.. code-block:: yaml

  ""
  

security (`config.ServerSecurityOptions`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  allowCors: false
  allowedHeaders: []
  allowedOrigins: []
  auditAccess: false
  secure: false
  ssl:
    certificateFile: ""
    keyFile: ""
  useAuth: false
  

thirdPartyConfig (`config.ThirdPartyConfigOptions`_)
--------------------------------------------------------------------------------

Deprecated please use auth.appAuth.thirdPartyConfig instead.

**Default Value**: 

.. code-block:: yaml

  flyteClient:
    clientId: ""
    redirectUri: ""
    scopes: []
  

config.ServerSecurityOptions
--------------------------------------------------------------------------------

secure (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "false"
  

ssl (`config.SslOptions`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  certificateFile: ""
  keyFile: ""
  

useAuth (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "false"
  

auditAccess (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "false"
  

allowCors (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "false"
  

allowedOrigins ([]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  []
  

allowedHeaders ([]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  []
  

config.SslOptions
--------------------------------------------------------------------------------

certificateFile (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

keyFile (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Section: storage
================================================================================

type (string)
--------------------------------------------------------------------------------

Sets the type of storage to configure [s3/minio/local/mem/stow].

**Default Value**: 

.. code-block:: yaml

  s3
  

connection (`storage.ConnectionConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  access-key: ""
  auth-type: iam
  disable-ssl: false
  endpoint: ""
  region: us-east-1
  secret-key: ""
  

stow (`storage.StowConfig`_)
--------------------------------------------------------------------------------

Storage config for stow backend.

**Default Value**: 

.. code-block:: yaml

  {}
  

container (string)
--------------------------------------------------------------------------------

Initial container (in s3 a bucket) to create -if it doesn't exist-.'

**Default Value**: 

.. code-block:: yaml

  ""
  

enable-multicontainer (bool)
--------------------------------------------------------------------------------

If this is true, then the container argument is overlooked and redundant. This config will automatically open new connections to new containers/buckets as they are encountered

**Default Value**: 

.. code-block:: yaml

  "false"
  

cache (`storage.CachingConfig`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  max_size_mbs: 0
  target_gc_percent: 0
  

limits (`storage.LimitsConfig`_)
--------------------------------------------------------------------------------

Sets limits for stores.

**Default Value**: 

.. code-block:: yaml

  maxDownloadMBs: 2
  

defaultHttpClient (`storage.HTTPClientConfig`_)
--------------------------------------------------------------------------------

Sets the default http client config.

**Default Value**: 

.. code-block:: yaml

  headers: null
  timeout: 0s
  

storage.CachingConfig
--------------------------------------------------------------------------------

max_size_mbs (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum size of the cache where the Blob store data is cached in-memory. If not specified or set to 0, cache is not used

**Default Value**: 

.. code-block:: yaml

  "0"
  

target_gc_percent (int)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Sets the garbage collection target percentage.

**Default Value**: 

.. code-block:: yaml

  "0"
  

storage.ConnectionConfig
--------------------------------------------------------------------------------

endpoint (`config.URL`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

URL for storage client to connect to.

**Default Value**: 

.. code-block:: yaml

  ""
  

auth-type (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Auth Type to use [iam,accesskey].

**Default Value**: 

.. code-block:: yaml

  iam
  

access-key (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Access key to use. Only required when authtype is set to accesskey.

**Default Value**: 

.. code-block:: yaml

  ""
  

secret-key (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Secret to use when accesskey is set.

**Default Value**: 

.. code-block:: yaml

  ""
  

region (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Region to connect to.

**Default Value**: 

.. code-block:: yaml

  us-east-1
  

disable-ssl (bool)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Disables SSL connection. Should only be used for development.

**Default Value**: 

.. code-block:: yaml

  "false"
  

storage.HTTPClientConfig
--------------------------------------------------------------------------------

headers (map[string][]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

timeout (`config.Duration`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Sets time out on the http client.

**Default Value**: 

.. code-block:: yaml

  0s
  

storage.LimitsConfig
--------------------------------------------------------------------------------

maxDownloadMBs (int64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Maximum allowed download size (in MBs) per call.

**Default Value**: 

.. code-block:: yaml

  "2"
  

storage.StowConfig
--------------------------------------------------------------------------------

kind (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Kind of Stow backend to use. Refer to github/graymeta/stow

**Default Value**: 

.. code-block:: yaml

  ""
  

config (map[string]string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

Configuration for stow backend. Refer to github/graymeta/stow

**Default Value**: 

.. code-block:: yaml

  {}
  

Section: task_resources
================================================================================

defaults (`interfaces.TaskResourceSet`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  cpu: "0"
  ephemeralStorage: "0"
  gpu: "0"
  memory: "0"
  storage: "0"
  

limits (`interfaces.TaskResourceSet`_)
--------------------------------------------------------------------------------

**Default Value**: 

.. code-block:: yaml

  cpu: "0"
  ephemeralStorage: "0"
  gpu: "0"
  memory: "0"
  storage: "0"
  

interfaces.TaskResourceSet
--------------------------------------------------------------------------------

cpu (`resource.Quantity`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

gpu (`resource.Quantity`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

memory (`resource.Quantity`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

storage (`resource.Quantity`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

ephemeralStorage (`resource.Quantity`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

resource.Quantity
--------------------------------------------------------------------------------

i (`resource.int64Amount`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  {}
  

d (`resource.infDecAmount`_)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  <nil>
  

s (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

Format (string)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  ""
  

resource.infDecAmount
--------------------------------------------------------------------------------

Dec (inf.Dec)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  null
  

resource.int64Amount
--------------------------------------------------------------------------------

value (int64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

scale (int32)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  "0"
  

