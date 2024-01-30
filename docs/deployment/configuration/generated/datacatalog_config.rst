.. _flytedatacatalog-config-specification:

#########################################
Flyte Datacatalog Configuration
#########################################

- `application <#section-application>`_

- `database <#section-database>`_

- `datacatalog <#section-datacatalog>`_

- `logger <#section-logger>`_

- `otel <#section-otel>`_

- `storage <#section-storage>`_

Section: application
========================================================================================================================

grpcPort (int)
------------------------------------------------------------------------------------------------------------------------

On which grpc port to serve Catalog

**Default Value**: 

.. code-block:: yaml

  "8081"
  

grpcServerReflection (bool)
------------------------------------------------------------------------------------------------------------------------

Enable GRPC Server Reflection

**Default Value**: 

.. code-block:: yaml

  "true"
  

httpPort (int)
------------------------------------------------------------------------------------------------------------------------

On which http port to serve Catalog

**Default Value**: 

.. code-block:: yaml

  "8080"
  

secure (bool)
------------------------------------------------------------------------------------------------------------------------

Whether to run Catalog in secure mode or not

**Default Value**: 

.. code-block:: yaml

  "false"
  

readHeaderTimeoutSeconds (int)
------------------------------------------------------------------------------------------------------------------------

The amount of time allowed to read request headers.

**Default Value**: 

.. code-block:: yaml

  "32"
  

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
  

config.Duration
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Duration (int64)
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

**Default Value**: 

.. code-block:: yaml

  1h0m0s
  

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
  

Section: datacatalog
========================================================================================================================

storage-prefix (string)
------------------------------------------------------------------------------------------------------------------------

StoragePrefix specifies the prefix where DataCatalog stores offloaded ArtifactData in CloudStorage. If not specified, the data will be stored in the base container directly.

**Default Value**: 

.. code-block:: yaml

  metadata
  

metrics-scope (string)
------------------------------------------------------------------------------------------------------------------------

Scope that the metrics will record under.

**Default Value**: 

.. code-block:: yaml

  datacatalog
  

profiler-port (int)
------------------------------------------------------------------------------------------------------------------------

Port that the profiling service is listening on.

**Default Value**: 

.. code-block:: yaml

  "10254"
  

heartbeat-grace-period-multiplier (int)
------------------------------------------------------------------------------------------------------------------------

Number of heartbeats before a reservation expires without an extension.

**Default Value**: 

.. code-block:: yaml

  "3"
  

max-reservation-heartbeat (`config.Duration`_)
------------------------------------------------------------------------------------------------------------------------

The maximum available reservation extension heartbeat interval.

**Default Value**: 

.. code-block:: yaml

  10s
  

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
  

