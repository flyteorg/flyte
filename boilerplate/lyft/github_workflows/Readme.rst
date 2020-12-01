Golang Github Actions
~~~~~~~~~~~~~~~~~

Provides a two github actions workflows.

**To Enable:**

Add ``lyft/github_workflows`` to your ``boilerplate/update.cfg`` file.

Add a github secret ``flytegithub_repo`` with a the name of your fork (e.g. ``my_company/flytepropeller``).

The actions will push to 2 repos:

	1. ``docker.pkg.github.com/lyft/<repo>/operator``
	2. ``docker.pkg.github.com/lyft/<repo>/operator-stages`` : this repo is used to cache build stages to speed up iterative builds after.

There are two workflows that get deployed:

	1. A workflow that runs on Pull Requests to build and push images to github registy tagged with the commit sha.
	2. A workflow that runs on master merges that bump the patch version of release tag, builds and pushes images to github registry tagged with the version, commit sha as well as "latest"
