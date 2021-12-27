.. _deployment-cluster-config:

##############
Cluster Config
##############


.. panels::
    :header: text-center

    .. link-button:: deployment-cluster-config-auth-setup
       :type: ref
       :text: Auth
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Basic OIDC and Authentication Setup

    ---

    .. link-button:: deployment-cluster-config-auth-migration
       :type: ref
       :text: Auth Migration Guide
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Migration guide to help you move to Admin's own authorization server.

    ---

    .. link-button:: deployment-cluster-config-general
       :type: ref
       :text: Configuration of Custom K8s Resources, Resource Quotas
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    How to use Flyte's cluster-resource-controller to control specific Kubernetes resources and administer project/domain-specific resource quotas (say, to limit the number of CPUs/GPUs/mem per tenant).

    ---

    .. link-button:: deployment-customizable-resources
       :type: ref
       :text: New Configurations for Specific Tenants
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Creating new default configurations or overriding certain values for specific combinations of user projects, domains and workflows through Flyte APIs.

    ---

    .. link-button:: flyteadmin-config-specification
       :type: ref
       :text: Flyte Admin Configuration
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    All about FlyteAdmin configuration specification.

    ---

    .. link-button:: flytepropeller-config-specification
       :type: ref
       :text: Flyte Propeller Configuration
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Detailed list of all Flyte Propeller configuration options and their default values.

    ---

    .. link-button:: deployment-cluster-config-notifications
       :type: ref
       :text: Notifications
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Guide to setting up and configuring notifications.

    ---

    .. link-button:: deployment-cluster-config-eventing
       :type: ref
       :text: External Events
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    How to set up Flyte to emit events to third-parties.

    ---

    .. link-button:: deployment-cluster-config-monitoring
       :type: ref
       :text: Monitoring
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Guide to setting up and configuring observability.

    ---

    .. link-button:: deployment-cluster-config-performance
       :type: ref
       :text: Performance
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Tweaks to improve performance of the core Flyte engine.

.. toctree::
    :maxdepth: 1
    :name: Cluster Config
    :hidden:

    auth_setup
    auth_migration
    auth_appendix
    general
    customizable_resources
    flyteadmin_config
    flytepropeller_config
    monitoring
    notifications
    performance
    eventing
