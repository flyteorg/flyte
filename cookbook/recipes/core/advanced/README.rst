.. _advanced:

Advanced Examples: Extend Flytekit, combine techniques etc
-----------------------------------------------------------

Now that you have seen the capabilities of Flyte and flytekit, you might have found some cases in which you may want to
extend flytekit natively for your own project or better yet contribute to the open source community. This section provides
examples of how flytekit can be easily extended. Flytekit allows 2 fundamental extensions

#. Adding new user space task types. These task types do not need any backend changes and are as simple as writing python
   classes or extending existing python classes to achieve some new functionality.
#. Adding new Types to flytekit's type system. This could enable Flyte to extend its native understanding of more
   complex types or add new types of structured object handling

.. note::

    If you are interested in backend extensions for Flyte - read about it in :any:`working_hosted_service` section.


