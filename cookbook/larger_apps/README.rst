.. _larger_apps:

Building Large Apps
--------------------

So far in the *User Guide* you've been running Flyte workflows as one-off
scripts, which is useful for quick prototyping and iteration of small ideas
on a Flyte cluster.

However, if you need to build a larger Flyte app with sub-modules or
sub-packages to organize the logic of your tasks and workflows, you can use
`flyte python templates <https://github.com/flyteorg/flytekit-python-template>`__
to quickly materialize the directory structure with boilerplate files that
we recommend for structuring an app with multiple workflows or helper
modules.
