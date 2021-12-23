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
       :text: Configuration of Custom K8s resources, Resource quotas etc
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    This section covers how you can use Flyte's cluster-resource-controller to control your specific kubernetes resources, administer project/domain specific resource quotas, so that you can limit number of cpus/gpus/mem per tenant etc.

    ---

    .. link-button:: deployment-customizable-resources
       :type: ref
       :text: Create new configurations for specific combinations of user projects, domains and workflows that override default values.
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    This section covers how you can create new default configuration or override certain values for specific tenants through Flyte API's

    ---

    .. link-button:: flyteadmin-config-specification
       :type: ref
       :text: Flyte Admin Configuration
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Flyte Admin configuration specification are covered here.

    ---

    .. link-button:: flytepropeller-config-specification
       :type: ref
       :text: Flyte Propeller Configuration
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Detailed list of all Flyte Propeller configuration options and their default values can be found here.

    ---

    .. link-button:: deployment-cluster-config-notifications
       :type: ref
       :text: Notifications
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Guide to setting up and configuring notifications.

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
    Tweaks to improve performance of the core Flyte engine

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
