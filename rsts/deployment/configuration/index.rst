.. _deployment-configuration:

##############
Cluster Config
##############

.. important::

   The configuration instructions in this section are for the ``flyte`` and ``flyte-core`` Helm charts. The configuration
   changes discussed will closely mirror the ``flyte-binary`` Helm chart, the specific documentation for which we will
   work on soon. For the most part, if you're using the ``flyte-binary`` chart, instead of specifying under a yaml file
   like ``cloud_events.yaml`` in :ref:`deployment-configuration-cloud-event`, you'll need to add those settings under
   the ``inline`` section here:

   .. literalinclude:: ../../../charts/flyte-binary/eks-production.yaml
      :language: yaml
      :lines: 30-41
      :caption: charts/flyte-binary/eks-production.yaml


.. panels::
    :header: text-center
    :column: col-lg-12 p-2

    .. link-button:: deployment-configuration-auth-setup
       :type: ref
       :text: Authenticating in Flyte
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Basic OIDC and Authentication Setup

    ---

    .. link-button:: deployment-configuration-auth-migration
       :type: ref
       :text: Migrating Your Authentication Config
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Migration guide to move to Admin's own authorization server.

    ---

    .. link-button:: deployment-configuration-auth-appendix
       :type: ref
       :text: Understanding Authentication in Detail
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Migration guide to move to Admin's own authorization server.

    ---

    .. link-button:: deployment-configuration-general
       :type: ref
       :text: Configuring Custom K8s Resources
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    How to use Flyte's cluster-resource-controller to control specific Kubernetes resources and administer project/domain-specific resource quotas (say, to limit the number of CPUs/GPUs/mem per tenant).

    ---

    .. link-button:: deployment-configuration-customizable-resources
       :type: ref
       :text: Adding New Customizable Resources
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Creating new default configurations or overriding certain values for specific combinations of user projects, domains and workflows through Flyte APIs.

    ---

    .. link-button:: deployment-configuration-notifications
       :type: ref
       :text: Notifications
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Guide to setting up and configuring notifications.

    ---

    .. link-button:: deployment-configuration-cloud-event
       :type: ref
       :text: External Events
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    How to set up Flyte to emit events to third-parties.

    ---

    .. link-button:: deployment-configuration-monitoring
       :type: ref
       :text: Monitoring
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Guide to setting up and configuring observability.

    ---

    .. link-button:: deployment-configuration-performance
       :type: ref
       :text: Optimizing Performance
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Improve the performance of the core Flyte engine.

    ---

    .. link-button:: deployment-configuration-eventing
       :type: ref
       :text: Platform Events
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Configure Flyte to to send events to external pub/sub systems.

    ---

    .. link-button:: deployment-configuration-generated
       :type: ref
       :text: Generated Documentation
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Generated documentation for the configuration objects for Flyte backend services.

.. toctree::
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
    generated/index
