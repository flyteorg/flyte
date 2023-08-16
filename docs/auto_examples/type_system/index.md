(flyte_type_system)=

# Type System

Flyte is a data-aware DAG scheduling system. The graph itself is derived
automatically from the flow of data and this closely resembles how a functional
programming language passes data between methods.

Data awareness is powered by Flyte's own type system, which closely maps most programming languages. These types are what power Flyte's magic of:

- Data lineage
- Memoization
- Auto parallelization
- Simplifying access to data
- Auto generated CLI and Launch UI

```{auto-examples-toc}
flyte_python_types
schema
structured_dataset
typed_schema
pytorch_types
custom_objects
enums
flyte_pickle
```
