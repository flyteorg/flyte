(chain_flyte_entities)=

# Chaining Flyte entities

```{eval-rst}
.. tags:: Basic
```

Flytekit offers a mechanism for chaining Flyte entities using the `>>` operator.
This is particularly valuable when chaining tasks, subworkflows, and launch plans without the need for data flow between the entities.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

## Tasks

Let's establish a sequence where `t1()` occurs after `t0()`, and `t2()` follows `t1()`.
 
```{literalinclude} /examples/advanced_composition/advanced_composition/chain_entities.py
:caption: advanced_composition/chain_entities.py
:lines: 1-30
```

(chain_subworkflow)=
## Subworkflows

Just like tasks, you can chain {ref}`subworkflows <subworkflow>`.

```{literalinclude} /examples/advanced_composition/advanced_composition/chain_entities.py
:caption: advanced_composition/chain_entities.py
:lines: 34-49
```

## Launch plans

Like subworkflows, you can chain {ref}`launch plans <Launch plans>`.


```{literalinclude} /examples/advanced_composition/advanced_composition/chain_entities.py
:caption: advanced_composition/chain_entities.py
:lines: 55-60
```

To run the provided workflows on the Flyte cluster, use the following commands:

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/ab2e8e84362c5b06d2eea0d1d6e29ea7fe460608/examples/advanced_composition/advanced_composition/chain_entities.py \
  chain_tasks_wf
```

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/ab2e8e84362c5b06d2eea0d1d6e29ea7fe460608/examples/advanced_composition/advanced_composition/chain_entities.py \
  chain_workflows_wf
```

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/ab2e8e84362c5b06d2eea0d1d6e29ea7fe460608/examples/advanced_composition/advanced_composition/chain_entities.py \
  chain_launchplans_wf
```

:::{note}
Chaining tasks, subworkflows, and launch plans is not supported in local environments.
Follow the progress of this issue [here](https://github.com/flyteorg/flyte/issues/4080).
:::

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/advanced_composition/
