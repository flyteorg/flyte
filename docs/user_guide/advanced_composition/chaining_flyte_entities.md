(chain_flyte_entities)=

# Chaining Flyte entities

```{eval-rst}
.. tags:: Basic
```

Flytekit offers a mechanism for chaining Flyte entities using the `>>` operator.
This is particularly valuable when chaining tasks and subworkflows without the need for data flow between the entities.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

## Tasks

Let's establish a sequence where `t1()` occurs after `t0()`, and `t2()` follows `t1()`.
 
```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/chain_entities.py
:caption: advanced_composition/chain_entities.py
:lines: 1-30
```

(chain_subworkflow)=
## Subworkflows

Just like tasks, you can chain {ref}`subworkflows <subworkflow>`.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/chain_entities.py
:caption: advanced_composition/chain_entities.py
:lines: 34-49
```

To run the provided workflows on the Flyte cluster, use the following commands:

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/chain_entities.py \
  chain_tasks_wf
```

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/chain_entities.py \
  chain_workflows_wf
```

:::{note}
Chaining tasks and subworkflows is not supported in local environments.
Follow the progress of this issue [here](https://github.com/flyteorg/flyte/issues/4080).
:::

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/advanced_composition/
