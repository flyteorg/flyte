# Flytekit Papermill Plugin

It is possible to run a Jupyter notebook as a Flyte task using Papermill. Papermill executes the notebook as a whole, so before using this plugin, it is essential to construct your notebook as recommended by Papermill.

To install the plugin, run the following command:

```bash
pip install flytekitplugins-papermill
```

An [example](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/flytekit_plugins/papermilltasks/simple.html#sphx-glr-auto-integrations-flytekit-plugins-papermilltasks-simple-py) can be found in the documentation. We also have a [tutorial](https://docs.flyte.org/projects/cookbook/en/latest/auto/case_studies/feature_engineering/eda/index.html) showcasing the various ways in which you can leverage the Papermill plugin.
