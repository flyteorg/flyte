(deployment-configuration)=

# Configuration

This section will cover how to configure your Flyte cluster for features like
authentication, monitoring, and notifications.

````{important}
The configuration instructions in this section are for the `flyte` and `flyte-core` Helm charts, which is for
the {ref}`multi-cluster setup <deployment-deployment-multicluster>`.

If you're using the `flyte-binary` chart for the  {ref}`single cluster setup <deployment-deployment-cloud-simple>`,
instead of specifying configuration under a yaml file like `cloud_events.yaml` in {ref}`deployment-configuration-cloud-event`,
you'll need to add the configuration settings under the `inline` section in the `eks-production.yaml` file:

```{eval-rst}
.. literalinclude:: ../../../charts/flyte-binary/eks-production.yaml
   :language: yaml
   :lines: 30-41
   :caption: charts/flyte-binary/eks-production.yaml
```

````


```{list-table}
:header-rows: 0
:widths: 20 30

* - {ref}`Authenticating in Flyte <deployment-configuration-auth-setup>`
  - Basic OIDC and Authentication Setup
* - {ref}`Migrating Your Authentication Config <deployment-configuration-auth-migration>`
  - Migration guide to move to Admin's own authorization server.
* - {ref}`Understanding Authentication <deployment-configuration-auth-appendix>`
  - Migration guide to move to Admin's own authorization server.
* - {ref}`Configuring Custom K8s Resources <deployment-configuration-general>`
  - Use Flyte's cluster-resource-controller to control specific Kubernetes resources and administer project/domain-specific CPU/GPU/memory resource quotas.
* - {ref}`Adding New Customizable Resources <deployment-configuration-customizable-resources>`
  - Create new default configurations or overriding certain values for specific combinations of user projects, domains and workflows through Flyte APIs.
* - {ref}`Notifications <deployment-configuration-notifications>`
  - Guide to setting up and configuring notifications.
* - {ref}`External Events <deployment-configuration-cloud-event>`
  - How to set up Flyte to emit events to third-parties.
* - {ref}`Monitoring <deployment-configuration-monitoring>`
  - Guide to setting up and configuring observability.
* - {ref}`Optimizing Performance <deployment-configuration-performance>`
  - Improve the performance of the core Flyte engine.
* - {ref}`Platform Events <deployment-configuration-eventing>`
  - Configure Flyte to to send events to external pub/sub systems.
```

```{toctree}
:maxdepth: 1
:name: Cluster Config
:hidden:

auth_setup
auth_migration
auth_appendix
general
customizable_resources
monitoring
notifications
performance
cloud_event
```
