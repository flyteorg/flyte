================
Flyte IDL
================
This repository contains all of Flyte's IDLs.

The contents of this Readme will be deprecated shortly.  Please refer to the Flyte `contributor's guide <https://github.com/lyft/flyte>`__ in the future.

Generating Protobufs -> Code
#############################

Mac OS
*******

1. Install Docker for Mac (https://www.docker.com/docker-mac)
2. Start 'docker' and sign in.  You should see a whale icon in your toolbar.
3. cd to the root of your flyteidl repository.
4. Run ``./generate_protos.sh`` to generate just the protobuf files.  Side note: running ``make generate`` will generate protos along with some other things like mock classes.

Admin Client Generation
*************************

Please see the Flyte Tools documentation on the `generator <https://github.com/lyft/flytetools#swagger-client-code-generator>`__ for more information.
