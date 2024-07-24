from flytekit.models import matchable_resource


def test_cluster_resource_attributes():
    obj = matchable_resource.ClusterResourceAttributes({"cpu": "one million", "gpu": "just one"})
    assert obj.attributes == {"cpu": "one million", "gpu": "just one"}
    assert obj == matchable_resource.ClusterResourceAttributes.from_flyte_idl(obj.to_flyte_idl())


def test_execution_queue_attributes():
    obj = matchable_resource.ExecutionQueueAttributes(["foo", "bar", "baz"])
    assert obj.tags == ["foo", "bar", "baz"]
    assert obj == matchable_resource.ExecutionQueueAttributes.from_flyte_idl(obj.to_flyte_idl())


def test_execution_cluster_label():
    obj = matchable_resource.ExecutionClusterLabel("my_cluster")
    assert obj.value == "my_cluster"
    assert obj == matchable_resource.ExecutionClusterLabel.from_flyte_idl(obj.to_flyte_idl())


def test_plugin_override():
    obj = matchable_resource.PluginOverride(
        "task_type", ["acceptable", "override"], matchable_resource.PluginOverride.USE_DEFAULT
    )
    assert obj.task_type == "task_type"
    assert obj.plugin_id == ["acceptable", "override"]
    assert obj.missing_plugin_behavior == matchable_resource.PluginOverride.USE_DEFAULT
    assert obj == matchable_resource.PluginOverride.from_flyte_idl(obj.to_flyte_idl())


def test_plugin_overrides():
    plugin_override = matchable_resource.PluginOverride(
        "task_type", ["acceptable", "override"], matchable_resource.PluginOverride.USE_DEFAULT
    )
    obj = matchable_resource.PluginOverrides(overrides=[plugin_override])
    assert obj.overrides == [plugin_override]
    assert obj == matchable_resource.PluginOverrides.from_flyte_idl(obj.to_flyte_idl())


def test_matchable_resource():
    cluster_resource_attrs = matchable_resource.ClusterResourceAttributes({"cpu": "one million", "gpu": "just one"})
    obj = matchable_resource.MatchingAttributes(cluster_resource_attributes=cluster_resource_attrs)
    assert obj.cluster_resource_attributes == cluster_resource_attrs
    assert obj == matchable_resource.MatchingAttributes.from_flyte_idl(obj.to_flyte_idl())

    execution_queue_attributes = matchable_resource.ExecutionQueueAttributes(["foo", "bar", "baz"])
    obj2 = matchable_resource.MatchingAttributes(execution_queue_attributes=execution_queue_attributes)
    assert obj2.execution_queue_attributes == execution_queue_attributes
    assert obj2 == matchable_resource.MatchingAttributes.from_flyte_idl(obj2.to_flyte_idl())

    execution_cluster_label = matchable_resource.ExecutionClusterLabel("my_cluster")
    obj2 = matchable_resource.MatchingAttributes(execution_cluster_label=execution_cluster_label)
    assert obj2.execution_cluster_label == execution_cluster_label
    assert obj2 == matchable_resource.MatchingAttributes.from_flyte_idl(obj2.to_flyte_idl())

    plugin_overrides_obj = matchable_resource.PluginOverrides(
        overrides=[
            matchable_resource.PluginOverride(
                "task_type", ["acceptable", "override"], matchable_resource.PluginOverride.USE_DEFAULT
            )
        ]
    )
    plugin_override_attributes = matchable_resource.MatchingAttributes(plugin_overrides=plugin_overrides_obj)
    assert plugin_override_attributes.plugin_overrides == plugin_overrides_obj
    assert plugin_override_attributes == matchable_resource.MatchingAttributes.from_flyte_idl(
        plugin_override_attributes.to_flyte_idl()
    )
