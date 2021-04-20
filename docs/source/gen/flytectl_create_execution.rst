.. _flytectl_create_execution:

flytectl create execution
-------------------------

Create execution resources

Synopsis
~~~~~~~~



Create the executions for given workflow/task in a project and domain.

There are three steps in generating an execution.

- Generate the execution spec file using the get command.
- Update the inputs for the execution if needed.
- Run the execution by passing in the generated yaml file.

The spec file should be generated first and then run the execution using the spec file.
You can reference the flytectl get task for more details

::

 flytectl get tasks -d development -p flytectldemo core.advanced.run_merge_sort.merge  --version v2 --execFile execution_spec.yaml

The generated file would look similar to this

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

And then can be passed through the command line.
Notice the source and target domain/projects can be different.
The root project and domain flags of -p and -d should point to task/launch plans project/domain.

::

 flytectl create execution --execFile execution_spec.yaml -p flytectldemo -d development --targetProject flytesnacks

Also an execution can be relaunched by passing in current execution id.

::

 flytectl create execution --relaunch ffb31066a0f8b4d52b77 -p flytectldemo -d development

Usage


::

  flytectl create execution [flags]

Options
~~~~~~~

::

      --execFile string          file for the execution params.If not specified defaults to <<workflow/task>_name>.execution_spec.yaml
  -h, --help                     help for execution
      --iamRoleARN string        iam role ARN AuthRole for launching execution.
      --kubeServiceAcct string   kubernetes service account AuthRole for launching execution.
      --relaunch string          execution id to be relaunched.
      --targetDomain string      project where execution needs to be created.If not specified configured domain would be used.
      --targetProject string     project where execution needs to be created.If not specified configured project would be used.

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --admin.authorizationHeader string           Custom metadata header to pass JWT
      --admin.authorizationServerUrl string        This is the URL to your IDP's authorization server'
      --admin.clientId string                      Client ID
      --admin.clientSecretLocation string          File containing the client secret
      --admin.endpoint string                      For admin types,  specify where the uri of the service is located.
      --admin.insecure                             Use insecure connection.
      --admin.maxBackoffDelay string               Max delay for grpc backoff (default "8s")
      --admin.maxRetries int                       Max number of gRPC retries (default 4)
      --admin.perRetryTimeout string               gRPC per retry timeout (default "15s")
      --admin.scopes strings                       List of scopes to request
      --admin.tokenUrl string                      Your IDPs token endpoint
      --admin.useAuth                              Whether or not to try to authenticate with options below
      --adminutils.batchSize int                   Maximum number of records to retrieve per call. (default 100)
      --adminutils.maxRecords int                  Maximum number of records to retrieve. (default 500)
      --config string                              config file (default is $HOME/config.yaml)
  -d, --domain string                              Specifies the Flyte project's domain.
      --logger.formatter.type string               Sets logging format type. (default "json")
      --logger.level int                           Sets the minimum logging level. (default 4)
      --logger.mute                                Mutes all logs regardless of severity. Intended for benchmarks/tests only.
      --logger.show-source                         Includes source code location in logs.
  -o, --output string                              Specifies the output type - supported formats [TABLE JSON YAML] (default "TABLE")
  -p, --project string                             Specifies the Flyte project.
      --root.domain string                         Specified the domain to work on.
      --root.output string                         Specified the output type.
      --root.project string                        Specifies the project to work on.
      --storage.cache.max_size_mbs int             Maximum size of the cache where the Blob store data is cached in-memory. If not specified or set to 0,  cache is not used
      --storage.cache.target_gc_percent int        Sets the garbage collection target percentage.
      --storage.connection.access-key string       Access key to use. Only required when authtype is set to accesskey.
      --storage.connection.auth-type string        Auth Type to use [iam, accesskey]. (default "iam")
      --storage.connection.disable-ssl             Disables SSL connection. Should only be used for development.
      --storage.connection.endpoint string         URL for storage client to connect to.
      --storage.connection.region string           Region to connect to. (default "us-east-1")
      --storage.connection.secret-key string       Secret to use when accesskey is set.
      --storage.container string                   Initial container to create -if it doesn't exist-.'
      --storage.defaultHttpClient.timeout string   Sets time out on the http client. (default "0s")
      --storage.enable-multicontainer              If this is true,  then the container argument is overlooked and redundant. This config will automatically open new connections to new containers/buckets as they are encountered
      --storage.limits.maxDownloadMBs int          Maximum allowed download size (in MBs) per call. (default 2)
      --storage.type string                        Sets the type of storage to configure [s3/minio/local/mem/stow]. (default "s3")

SEE ALSO
~~~~~~~~

* :doc:`flytectl_create` 	 - Used for creating various flyte resources including tasks/workflows/launchplans/executions/project.

