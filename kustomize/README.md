# Install Flyte using Kustomize
Flyte can be deployed to a kubernetes cluster using a generated deployment yaml file. This file is generated using [Kustomize](https://kubectl.docs.kubernetes.io/guides/introduction/kustomize/).
Please refer to Kustomize documentation to see how it works.

In brief, Kustomize allows composing a deployment yaml using multiple components. In Flyte all the core components are listed under [Base Components](./base). The Base components also consist of
a composed set of components that can be deployed to a [Single Kubernetes cluster](./base/single_cluster). This deployment configures various components using [Flyte Configuration
system](todo).

The *Single Cluster* configuration on its own is not deployable. But individuals [overlays](./overlays) are deployable. 

Refer to
1. [Base Components](./base): If you want to build your own overlay start here
1. [overlays](./overlays): If you want to build on top of an existing overlay start here
