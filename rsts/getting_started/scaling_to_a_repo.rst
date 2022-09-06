.. scaling-to-a-repo:

###############
Scaling to a repo
###############

So far we have covered use cases that are limited to one file. But in real life scenarios we might deal with multiple files inside a repository. 
In this section we will cover how to work with Flyte with an entire repository containing multiple files. 

Currently, `pyflyte run` doesn't work on multiple files. In order to achieve working with multiple files we introduce `pyflyte register`

What is `pyflyte register`?
The command "pyflyte register" uses fast-registration to register all workflows that are currently available in the repository or directory. 
It is the same as carrying out the same action with two commands (pyflyte package and flytectl register) (registration). 
The files are directly uploaded to FlyteAdmin after the Python code has been compiled into protobuf objects. 
The protobuf objects are not written to the local disc during the procedure, making it challenging to introspect these objects because they are lost.
It automatically also handles all networking. 

Between the commands pyflyte package + flytectl register and pyflyte run, there is the pyflyte register command. It provides the features of the Pyflyte package (with smarter naming semantics and combining the network call into one step).

Usage: 

.. prompt:: bash

    pyflyte register --image ghcr.io/flyteorg/flytecookbook:core-latest --image trainer=ghcr.io/flyteorg/flytecookbook:core-latest --image predictor=ghcr.io/flyteorg/flytecookbook:core-latest --raw-data-prefix s3://development-service-flyte/reltsts flyte_basics