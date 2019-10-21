.. _install-dependencies:

Optional Components
-------------------

Scheduling
**********

Scheduling component of Flyte allows workflows to run at a configured cadence.
Currently, Flyte does not have a built in cron style scheduler.
It uses external services to manage these schedules and launch executions of scheduled workflows at specific intervals.

Data Catalog
------------

The Data Catalog component enables features like *Data Provenance*, *Data Lineage* and *Data Caching*.
Data Catalog is enabled in Flyte by default.
