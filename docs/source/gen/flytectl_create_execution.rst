.. _FlyteCTL_create_execution:

FlyteCTL create execution
-------------------------

Create execution resources

Synopsis
~~~~~~~~



Creates executions for a given workflow/task in a project and domain.

There are three steps in generating an execution:

- Generate the execution spec file using the get command.
- Update the inputs for the execution if needed.
- Run the execution by passing in the generated yaml file.

The spec file should be generated first and then, the execution should be run using the spec file.
You can reference the FlyteCTL get task for more details.

::

 flytectl get tasks -d development -p flytectldemo core.advanced.run_merge_sort.merge  --version v2 --execFile execution_spec.yaml

The generated file would look similar to this:

.. code-block:: yaml

	 iamRoleARN: ""
	 inputs:
	   sorted_list1:
	   - 0
	   sorted_list2:
	   - 0
	 kubeServiceAcct: ""
	 targetDomain: ""
	 targetProject: ""
	 task: core.advanced.run_merge_sort.merge
	 version: "v2"


The generated file can be modified to change the input values.

.. code-block:: yaml

	 iamRoleARN: 'arn:aws:iam::12345678:role/defaultrole'
	 inputs:
	   sorted_list1:
	   - 2
	   - 4
	   - 6
	   sorted_list2:
	   - 1
	   - 3
	   - 5
	 kubeServiceAcct: ""
	 targetDomain: ""
	 targetProject: ""
	 task: core.advanced.run_merge_sort.merge
	 version: "v2"

It can then be passed through the command line.
Notice that the source and target domain/projects can be different.
The root project and domain flags of -p and -d should point to the task/launch plans project/domain.

::

 flytectl create execution --execFile execution_spec.yaml -p flytectldemo -d development --targetProject flytesnacks

Also, an execution can be relaunched by passing in the current execution id.

::

 flytectl create execution --relaunch ffb31066a0f8b4d52b77 -p flytectldemo -d development

An execution can be recovered, i.e., recreated from the last known failure point for previously-run workflow execution.
See :ref:`ref_flyteidl.admin.ExecutionRecoverRequest` for more details.

::

 flytectl create execution --recover ffb31066a0f8b4d52b77 -p flytectldemo -d development

Generic data types are also supported for execution in a similar manner. Following is a sample of how the inputs need to be specified while creating the execution.
The spec file should be generated first and then, the execution should be run using the spec file.

::

 flytectl get task -d development -p flytectldemo  core.type_system.custom_objects.add --execFile adddatanum.yaml

The generated file would look similar to this. Here, empty values have been dumped for generic data type x and y. 

::

    iamRoleARN: ""
    inputs:
      "x": {}
      "y": {}
    kubeServiceAcct: ""
    targetDomain: ""
    targetProject: ""
    task: core.type_system.custom_objects.add
    version: v3

Modified file with struct data populated for x and y parameters for the task core.type_system.custom_objects.add

::

  iamRoleARN: "arn:aws:iam::123456789:role/dummy"
  inputs:
    "x":
      "x": 2
      "y": ydatafory
      "z":
        1 : "foo"
        2 : "bar"
    "y":
      "x": 3
      "y": ydataforx
      "z":
        3 : "buzz"
        4 : "lightyear"
  kubeServiceAcct: ""
  targetDomain: ""
  targetProject: ""
  task: core.type_system.custom_objects.add
  version: v3

Usage


::

  FlyteCTL create execution [flags]

Options
~~~~~~~

::

      --dryRun                   execute command without making any modifications.
      --execFile string          file for the execution params.If not specified defaults to <<workflow/task>_name>.execution_spec.yaml
  -h, --help                     help for execution
      --iamRoleARN string        iam role ARN AuthRole for launching execution.
      --kubeServiceAcct string   kubernetes service account AuthRole for launching execution.
      --recover string           execution id to be recreated from the last known failure point.
      --relaunch string          execution id to be relaunched.
      --targetDomain string      project where execution needs to be created.If not specified configured domain would be used.
      --targetProject string     project where execution needs to be created.If not specified configured project would be used.
      --task string              
      --version string           
      --workflow string          

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --admin.authorizationHeader string           Custom metadata header to pass JWT
      --admin.authorizationServerUrl string        This is the URL to your IdP's authorization server. It'll default to Endpoint
      --admin.clientId string                      Client ID (default "flytepropeller")
      --admin.clientSecretLocation string          File containing the client secret (default "/etc/secrets/client_secret")
      --admin.endpoint string                      For admin types,  specify where the uri of the service is located.
      --admin.insecure                             Use insecure connection.
      --admin.insecureSkipVerify                   InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name. Caution : shouldn't be use for production usecases'
      --admin.maxBackoffDelay string               Max delay for grpc backoff (default "8s")
      --admin.maxRetries int                       Max number of gRPC retries (default 4)
      --admin.perRetryTimeout string               gRPC per retry timeout (default "15s")
      --admin.pkceConfig.refreshTime string         (default "5m0s")
      --admin.pkceConfig.timeout string             (default "15s")
      --admin.scopes strings                       List of scopes to request
      --admin.tokenUrl string                      OPTIONAL: Your IdP's token endpoint. It'll be discovered from flyte admin's OAuth Metadata endpoint if not provided.
      --admin.useAuth                              Deprecated: Auth will be enabled/disabled based on admin's dynamically discovered information.
  -c, --config string                              config file (default is $HOME/.flyte/config.yaml)
  -d, --domain string                              Specifies the Flyte project's domain.
      --logger.formatter.type string               Sets logging format type. (default "json")
      --logger.level int                           Sets the minimum logging level. (default 4)
      --logger.mute                                Mutes all logs regardless of severity. Intended for benchmarks/tests only.
      --logger.show-source                         Includes source code location in logs.
  -o, --output string                              Specifies the output type - supported formats [TABLE JSON YAML DOT DOTURL]. NOTE: dot, doturl are only supported for Workflow (default "TABLE")
  -p, --project string                             Specifies the Flyte project.
      --storage.cache.max_size_mbs int             Maximum size of the cache where the Blob store data is cached in-memory. If not specified or set to 0,  cache is not used
      --storage.cache.target_gc_percent int        Sets the garbage collection target percentage.
      --storage.connection.access-key string       Access key to use. Only required when authtype is set to accesskey.
      --storage.connection.auth-type string        Auth Type to use [iam, accesskey]. (default "iam")
      --storage.connection.disable-ssl             Disables SSL connection. Should only be used for development.
      --storage.connection.endpoint string         URL for storage client to connect to.
      --storage.connection.region string           Region to connect to. (default "us-east-1")
      --storage.connection.secret-key string       Secret to use when accesskey is set.
      --storage.container string                   Initial container (in s3 a bucket) to create -if it doesn't exist-.'
      --storage.defaultHttpClient.timeout string   Sets time out on the http client. (default "0s")
      --storage.enable-multicontainer              If this is true,  then the container argument is overlooked and redundant. This config will automatically open new connections to new containers/buckets as they are encountered
      --storage.limits.maxDownloadMBs int          Maximum allowed download size (in MBs) per call. (default 2)
      --storage.stow.config stringToString         Configuration for stow backend. Refer to github/graymeta/stow (default [])
      --storage.stow.kind string                   Kind of Stow backend to use. Refer to github/graymeta/stow
      --storage.type string                        Sets the type of storage to configure [s3/minio/local/mem/stow]. (default "s3")

SEE ALSO
~~~~~~~~

* :doc:`FlyteCTL_create` 	 - Create various Flyte resources including tasks/workflows/launchplans/executions/project.

