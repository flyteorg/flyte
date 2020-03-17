Flyte End-to-End Testing
===============================

Flyte has implemented some light constructs around testing the platform end-to-end.  Service level integration tests generally only deal with API boundaries across one layer.  Service A might talk to Service B and to Service A's database in an integration test.  End-to-end testing is meant to simulate the entire Flyte platform, from a user's workflow code, all the way from registration to execution and verification of results.

Design
##########

Infrastructure
****************
End-to-end testing uses the `Dockernetes <https://github.com/lyft/dockernetes>`_ container that the Flyte team has put together.  This is a Docker container, that runs a Kubernetes cluster within it.  That means that from the Docker container, you have access to ``kubectl``.  Dockernetes doesn't currently give you access to the images that the Docker daemon running Dockernetes already has, but this is a feature that we may one day build out.


Container Entrypoint
**********************
Any workflow repository that will be the target of an end-to-end test will need to have the ``end2end_test`` Make target available.  This is the common entrypoint that the tooling in this repository will call.


Background
***********
End-to-end testing was initally going to be implemented with K8s Jobs.  We thought that we'd have somewhere on the order of 100 jobs.  In implementing tests, we quickly realized that managing that number of jobs would be a lot of work - keeping track of timeouts, failures, partial pod failures like ImagePullBackoff issues, retries, etc.  What we'd end up writing for that would look very much like a higher level Jobs controller.  That would be way too much work.

We thought about directly running Docker, like ``docker run flytetester make endtoend``. That would work, except that we still have some networking limitations around gRPC ingress at time of writing, and port-forwarding within the Dockernetes container sounded like too much virtualization hocus-pocus.

So we stuck with the in-cluster idea, but decided to just keep it simple - run a Pod instead of a Job, and use simple Bash constructs to tell when things are done.  And keep the number of tests small, like one or two, at least for now.

For future tests, it may be best to run different execute.sh scripts.  That is, instead of multiple end-to-end tests in one Dockernetes container/cluster, each end-to-end test gets its own Dockernetes container/cluster.  This keeps things simpler, and also enables tests to do things like write a workflow with one version of FlyteAdmin, then upgrade Admin, and read it back with a newer version.
