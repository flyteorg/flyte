# Flyte v1.2.1 Release

## Improved Local Experience
This change introduces a new version of the `flytectl demo start` experience. With the `v0.6.24` version of `flytectl` and later, when you start the demo environment, you'll get

* Faster start up times after image downloaded as dependent images (like minio and postgres) now come pre-bundled.
* A local Docker registry to use. Build custom task images tagged with `localhost:30000` and push to make them available inside Flyte.

This release merges in all the changes here.
https://github.com/flyteorg/flyte/compare/v1.2.0...819457850b634e626cb233b500e7855b65fc9719

