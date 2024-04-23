(mmcloud_plugin_example)=
# Memory Machine Cloud

This example shows how to use the MMCloud plugin to execute tasks on [MemVerge Memory Machine Cloud][mmcloud].

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/mmcloud_plugin/mmcloud_plugin/mmcloud_plugin_example.py
:caption: mmcloud_plugin/mmcloud_plugin_example.py
:lines: 1-2
```

`MMCloudConfig` configures `MMCloudTask`. Tasks specified with `MMCloudConfig` will be executed using MMCloud. Tasks will be executed with requests `cpu="1"` and `mem="1Gi"` by default.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/mmcloud_plugin/mmcloud_plugin/mmcloud_plugin_example.py
:caption: mmcloud_plugin/mmcloud_plugin_example.py
:lines: 8-15
```

[Resource](https://docs.flyte.org/en/latest/user_guide/productionizing/customizing_task_resources.html) (cpu and mem) requests and limits, [container](https://docs.flyte.org/en/latest/user_guide/customizing_dependencies/multiple_images_in_a_workflow.html) images, and [environment](https://docs.flyte.org/en/latest/api/flytekit/generated/flytekit.task.html) variable specifications are supported.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/mmcloud_plugin/mmcloud_plugin/mmcloud_plugin_example.py
:caption: mmcloud_plugin/mmcloud_plugin_example.py
:lines: 18-32
```

[mmcloud]: https://www.mmcloud.io/
[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/mmcloud_plugin
