# 0.18.2 Release ChangeLog

[Closed Issues](https://github.com/flyteorg/flyte/issues?q=is%3Aissue+milestone%3A0.18.2+is%3Aclosed)

## UX
* Added advanced options to launch form
* Added support for all tasks-types (task execution view)
* Replaced execution id's with node id's on execution list-view
* Fixed bug with some properties not being repopulated on relaunch
* minor fixes
 
### FlyteKit
See the flytekit [0.25.0 release notes](https://github.com/flyteorg/flytekit/releases/tag/v0.25.0) for the full list of changes. Here are some of the highlights:

* Improved support for tasks that [run shell scripts](https://github.com/flyteorg/flytekit/pull/747)
* Support for more types in dataclasses:
    * [enums](https://github.com/flyteorg/flytekit/pull/753)
    * [FlyteFile](https://github.com/flyteorg/flytekit/pull/725) and [FlyteSchema](https://github.com/flyteorg/flytekit/pull/722)
* flyteremote improvements, including:
    * [Access to raw inputs and outputs](https://github.com/flyteorg/flytekit/pull/675)
    * [Ability to serialize tasks containing arbitrary images](https://github.com/flyteorg/flytekit/pull/733)
    * [Improved UX for navigation of subworkflows and launchplans](https://github.com/flyteorg/flytekit/pull/751)
    * [Support for FlytePickle](https://github.com/flyteorg/flytekit/pull/764)

## System
* Various stability fixes.
* Helm changes
    * [flyte-core](https://artifacthub.io/packages/helm/flyte/flyte-core) helm chart has reached release preview and can be leveraged to install your cloud(AWS/GCP) deployments of flyte.
    * Going forward flyte-core will install flyte native scheduler, For AWS backword compatibility you need to define `workflow_schedule.type` to `aws`. (https://github.com/flyteorg/flyte/pull/1896)
    * [flyte](https://artifacthub.io/packages/helm/flyte/flyte) helm chart has been refactored to depend on flyte-core helm chart and install additional dependencies to continue to provide a sandboxed installation of flyte.

    **Migration Notes**
    
    As part of this move, ``flyte`` helm chart is becoming the canonical sandbox cluster. It comes with all external resources needed to fully standup a Flyte cluster. If you have previously been using this chart to deploy flyte on your cloud providers, there will be changes you need to do to migrate:
    * If you have your own ``myvalues.yaml``, you will need to add another nesting level under ``flyte:`` for the sections that are now managed through ``flyte-core``. For example:
      ```yaml
      configmaps:
        ...
      flyteadmin:
        ...
      minio:
        ...
      countour:
        ...
      ```
    
      to:
      ```yaml
      flyte:
        configmaps:
          ...
        flyteadmin:
          ...
      minio:
        ...
      countour:
        ...    
      ```
    * Alternatively, if you do not have any dependency on external flyte depdencies, you can keep your ``myvalues.yaml`` and switch to using ``flyte-core`` helm chart directly with no changes.
