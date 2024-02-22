---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  main_language: python
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.16.1
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

(private_images)=

# Private images

As we learned in the {ref}`Flyte Fundamentals <containerizing_your_project>` guide,
Flyte uses OCI-compatible containers to package up your code and third-party
dependencies. For production use-cases your images may require proprietary code
and configuration that you want to keep private.

You can use different private container registries to host your images, such as
[AWS ECR](https://docs.aws.amazon.com/AmazonECR/latest/userguide/registry_auth.html),
[Docker Hub](https://docs.docker.com/docker-hub/repos/#private-repositories),
[GitLab Container Registry](https://docs.gitlab.com/ee/ci/docker/using_docker_images.html#access-an-image-from-a-private-container-registry),
and [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry#authenticating-to-the-container-registry).

To pull private images, ensure that you have the command line tools and login
information associated with the registry.

## Create a secret

First [create a secret](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/)
that contains all the credentials needed to log into the registry.

## Configure `imagePullSecrets`

Then, you'll need to specify a `imagePullSecrets` configuration to pull a
private image using one of two methods below.

```{eval-rst}
.. tabs::

   .. tab:: Service Account

      You can use the default or new service account for this option:

      1. Add your ``imagePullSecrets`` configuration to the
         `service account <https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#add-image-pull-secret-to-service-account>`__.
      2. Use this service account to log into the private registry and pull the image.
      3. When you create a task/workflow execution this service account should
         be specified to access the private image.

   .. tab:: Custom Pod Template

      This option uses a `custom pod template <https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/>`__
      to create a pod. This template is automatically added to every ``pod`` that
      Flyte creates.

      1. Add your ``imagePullSecrets`` configuration to this custom pod template.
      2. Update `FlytePropeller <https://github.com/flyteorg/flyteplugins/blob/e2971efbefd9126aca0290ddc931663605dec348/go/tasks/pluginmachinery/flytek8s/config/config.go#L157>`__ about the pod created in the previous step.
      3. FlytePropeller adds ``imagePullSecrets``, along with other customization for the pod,
         to the PodSpec, which should look similar to this
         `manifest <https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/#create-a-pod-that-uses-your-secret>`__.
      4. The pods with their keys can log in and access the images in the private registry.
         Once you set up the token to authenticate with the private registry, you can pull images from them.
```
