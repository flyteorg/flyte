.. _type system:

Type System
------------

.. currentmodule:: core

.. _flyte_type_system:

Flyte is a data-aware DAG scheduling system. The Graph itself is derived automatically from the flow of data and this
closely resembles how a functional programming language passes data between methods.

Data awareness is powered by Flyte's own type system, which closely maps most programming languages. These types are
what power Flyte's magic of:

- Data lineage
- Memoization
- Auto parallelization
- Simplifying access to data
- Auto generated CLI and Launch UI

It also opens up possibilities of future optimizations.
