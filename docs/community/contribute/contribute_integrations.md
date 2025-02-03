(contribute_integrations)=

## Guidelines to contributing integrations

If you plan to contribute a [plugin](https://docs.flyte.org/en/latest/flytesnacks/integrations/index.html) or [Agent](https://docs.flyte.org/en/latest/user_guide/flyte_agents/index.html#flyte-agents-guide), first of all , thank you!

To better support the lifecycle of maintaining an integration, we ask you to follow these guidelines:

-  Submit the code to the `community/plugins` folder
- Explicitly mark plugins as community-maintained in the import via `import flytekitplugins.contrib.x`
- Let the `flyte-bot` user have the `Sole Owner` role for your package in `pypi`. This ensures that new releases for your integration follow the `flytekit` release process.

