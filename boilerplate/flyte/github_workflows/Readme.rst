Golang Github Actions
~~~~~~~~~~~~~~~~~

Provides a two github actions workflows.

**To Enable:**

Add ``flyteorg/github_workflows`` to your ``boilerplate/update.cfg`` file.

Add a github secret ``package_name`` with the name to use for publishing (e.g. ``flytepropeller``). Typicaly, this will be the same name as the repository.

*Note*: If you are working on a fork, include that prefix in your package name (``myfork/flytepropeller``).

The actions will push to 2 repos:

	1. ``docker.pkg.github.com/flyteorg/<repo>/<package_name>``
	2. ``docker.pkg.github.com/flyteorg/<repo>/<package_name>-stages`` : this repo is used to cache build stages to speed up iterative builds after.

There are two workflows that get deployed:

	1. A workflow that runs on Pull Requests to build and push images to github registy tagged with the commit sha.
	2. A workflow that runs on master merges that bump the patch version of release tag, builds and pushes images to github registry tagged with the version, commit sha as well as "latest"
