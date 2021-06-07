Native Backend Plugins
----------------------

Native Backend Plugins are plugins that can be executed without any external service dependencies because the compute is
orchestrated by Flyte itself, within its provisioned kubernetes clusters.


.. panels::
    :header: text-center
 
    .. link-button:: auto/integrations/kubernetes/pod/index
       :type: ref
       :text: K8s Pods
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Execute K8s pods for arbitrary workloads.
    
    ---
 
    .. link-button:: auto/integrations/kubernetes/k8s_spark/index
       :type: ref
       :text: K8s Cluster Spark Jobs
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Run Spark jobs on a K8s Cluster.

    ---
 
    .. link-button:: kubeflow
       :type: ref
       :text: Kubeflow
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Run distributed training jobs using ``Kubeflow``.

.. toctree::
    :maxdepth: -1
    :caption: Integrations
    :hidden:
 
    auto/integrations/kubernetes/pod/index
    auto/integrations/kubernetes/k8s_spark/index
    kubeflow
