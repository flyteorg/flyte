Containerization
----------------

As a container-native orchestration tool, all Flyte tasks are executed within isolated containers. The ``flytekit``
python SDK enables you to easily configure them for your needs, whether it's running a raw container with arbitrary
commands, specifying multiple container images in a single workflow, injecting secrets into a container, or using
spot instances.
