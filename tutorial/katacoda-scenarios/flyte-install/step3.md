Register Workflow

`docker run --network host -e FLYTE_PLATFORM_URL='[[HOST_IP]]:30081' lyft/flytesnacks:b347efa300832f96d6cc0900a2aa6fbf6aad98da pyflyte -p flytesnacks -d development -c sandbox.config register workflows`{{execute HOST2}}

Please check tutorials for writing [your tasks ](https://lyft.github.io/flyte/user/getting_started/create_first.html)
