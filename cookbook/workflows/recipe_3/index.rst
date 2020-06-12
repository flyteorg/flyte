.. _recipe-3:

##########################################
How do I pass a custom object to a Task?
##########################################

*****************
Arbitrary Struct
*****************

One of the types that Flyte supports is a `struct type <https://github.com/lyft/flyteidl/blob/f8181796dc5cafe019b1493af1b64384ae1358f5/protos/flyteidl/core/types.proto#L20>`__.  The addition of this type was debated internally for some time as it effectively removes type-checking from Flyte, so use this with caution. It is a `scalar literal <https://github.com/lyft/flyteidl/blob/f8181796dc5cafe019b1493af1b64384ae1358f5/protos/flyteidl/core/literals.proto#L63>`__, even though in most cases it won't represent a scalar.

Flytekit implements this `here <https://github.com/lyft/flytekit/blob/1926b1285591ae941d7fc9bd4c2e4391c5c1b21b/flytekit/common/types/primitives.py#L501>`__.  Please see the included task/workflow in this directory for an example of its usage.

The UI currently does not support passing structs as inputs to workflows, so if you need to rely on this, you'll have to use ``flyte-cli`` for now. ::

    flyte-cli -p flytesnacks -d development execute-launch-plan -u lp:flytesnacks:development:workflows.recipe_3.tasks.GenericDemoWorkflow:477b61e4d9be818bbe6514500760053f4bc890db -r demo -- a='{"a": "hello", "b": "how are you", "c": ["array"], "d": {"nested": "value"}}'
