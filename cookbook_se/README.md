[Back to Snacks Menu](../README.md)

# :books: Flytekit Cookbook (2nd Edition)

This is a collection of short "how to" articles demonstrating the various capabilities and idiomatic usage of Flytekit.
Note this currently does not include articles/tid-bits on how to use the Flyte platform at large, though in the future we may expand it to that.

Also note that this is different than the first cookbook we had written (also in this repo). It may be premature to consider this version a second edition, but the style is a bit different.
  * This book is divided into four sections currently. The first three advance through increasingly complicated Flytekit examples, and the fourth deals with writing Flytekit as it relates to a Flyte deployment.
  * Also, this iteration of the cookbook is written in the literate programming style. That is, the contents of this directory should serve as both documentation, as well as a fully-functional Flyte workflow repo.

## :airplane: :closed_book: Built Book

The built documentation is available at https://flytecookbook.readthedocs.io/en/latest/.

## Using the Cookbook

The examples written in this cookbook are meant to used in few ways, as a reference to read if you have a specific question about a specific topic, as template to iterate on locally as many parts of Flytekit are built to be locally usable, and finally as the aforementioned complete workflow repo.

### Local Iteration

Play around with the various tasks and workflows locally through just in your IDE making sure your virtualenv is set correctly. Workflows are capable of local execution and tasks are as well. Custom tasks that cannot run locally (a Sagemaker or a Hive task for instance) can be mocked.

### Serialization
Getting your tasks, workflows, and launch plans to run on a Flyte platform is effectively a two-step process (though there are single commands in flytekit that combine both).  Serialization is the first step of that process, and it translates all your Flyte entities as defined in Python into Flyte IDL entities, defined in a protobuf file.

1. First commit your changes. Some of the steps below default to referencing the git sha.
1. Run `make serialize_sandbox`. This will build the image tagged with just `flytecookbook:<sha>`, no registry will be prefixed. See the image building section below for additional information.

Technically it is not required to build the image, and we are looking to change that, but for now it's done for two reasons:
1. Ensure that the same libraries used when running thing on Flyte are used when translating the tasks and workflows to their protobuf counterparts.
1. Container tasks (most tasks) need to be serialized with a static container image string, which is stored inside the [TaskTemplate](https://github.com/lyft/flyteidl/blob/cee566b2e6e109120f1bb34c980b1cfaf006a473/protos/flyteidl/core/tasks.proto#L134). The image building process stores the image name as an environment variable in the image itself, which is read during the serialization process.

### Registration
Registration is the second step of the process and it is merely sending those protobuf files to Flyte Admin, the control plane for Flyte, and telling Admin where to store them. This is done by another CLI that's also installed along with flytekit.

`flyte-cli register-files -p <myproject> -d development -v <gitsha> -h localhost:30081 [<my_files>]`

While it's not required to use the git sha for the version, it is highly recommended. If you are registering to your local Docker Desktop Flyte cluster, both the serialization and the registration command have been combined into `make register_sandbox`.

### Fast serialization and registration
When you make changes that affect code and not dependencies or other container attributes you can use fast registration to avoid having to rebuild your docker container between iterations.
Like above, the fast iteration process requires serialization and registration using two separate commands.

Serialization is as simple as running `make fast_serialize_sandbox`.

To register and serialize in one go simply run `make fast_register_sandbox`. The output of this command will show entities registered under `--local-source-root` scoped by `--pkgs` with a version of the form `fast...`. Simply execute
this version of your entities once the register command terminates as your normally do.

To run fast registration against a remote host, update the host argument in the `fast-register-files` command to use your
hosted endpoint:

```
flyte-cli fast-register-files -p ${PROJECT} -d development -h YOUR_HOST_HERE -i \
		--additional-distribution-dir s3://my-s3-bucket/fast/ --dest-dir /root/recipes \
		--assumable-iam-role user-role --output-location-prefix s3://my-s3-bucket/output/ ${CURDIR}/_pb_output/*
```

### Building Images
If you are just iterating locally, there is no need to push your Docker image. For Docker for Desktop at least, locally built images will be available for use in its K8s cluster.

If you would like to later push your image to a registry (Dockerhub, ECR, etc.), you can run,

```bash
REGISTRY=docker.io/corp make all_docker_push
```

## Additional Notes
1. Follow instructions in the main Flyte documentation to set up a local Flyte cluster. All the commands in this book assume that you are using Docker Desktop. If you are using minikube or another K8s deployment, the commands will need to be modified.
1. Please also ensure that you have a Python virtual environment installed and activated, and have pip installed the requirements.
1. Use flyte-cli to create a project named `flytesnacks` (the name the Makefile is set up to use) if it's not already there.

### Building the repo itself
```bash
make all_requirements
```


