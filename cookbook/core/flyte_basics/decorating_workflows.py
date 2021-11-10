"""
Decorating Workflows
--------------------

The behavior of workflows can be modified in a light-weight fashion by using the built-in :py:func:`~functools.wraps`
decorator pattern, similar to using decorators to
:ref:`customize task behavior <sphx_glr_auto_core_flyte_basics_decorating_tasks.py>`. However, unlike in the case of
tasks, we need to do a little extra work to make sure that the DAG underlying the workflow executes tasks in the
correct order.

Setup-Teardown Pattern
^^^^^^^^^^^^^^^^^^^^^^

The main use case of decorating
``@workflow``-decorated functions is when you want to establish a setup-teardown pattern that executes some task
before and after your main workflow logic. This is useful when integrating with other external services
like `wandb <https://wandb.ai/site>`__ or `clearml <https://clear.ml/>`__, which enable you to track metrics of model
training runs.

"""

from functools import partial, wraps

import flytekit
from flytekit import task, workflow, FlyteContextManager
from flytekit.core.node_creation import create_node

# %%
# First, let's define the tasks that we want for setup and teardown. In this example, we'll use the
# :py:class:`unittest.mock.MagicMock` class to create a fake external service that we want to initialize at the
# beginning of our workflow and finish at the end.

from unittest.mock import MagicMock

external_service = MagicMock()


@task
def setup():
    print("initializing external service")
    external_service.initialize(id=flytekit.current_context().execution_id)


@task
def teardown():
    print("finish external service")
    external_service.complete(id=flytekit.current_context().execution_id)


# %%
# As you can see, you can even use Flytekit's current context to access the ``execution_id`` of the current workflow
# if you need to link Flyte with the external service so that you reference the same unique identifier in both the
# external service and Flyte.
#
# Workflow Decorator
# ^^^^^^^^^^^^^^^^^^
#
# Next we create the decorator that we'll use to wrap our workflow function.

def setup_teardown(fn=None, *, before, after):
    @wraps(fn)
    def wrapper(*args, **kwargs):
        # get the current flyte context to obtain access to the compilation state of the workflow DAG.
        ctx = FlyteContextManager.current_context()

        # defines before node
        before_node = create_node(before)
        # ctx.compilation_state.nodes == [before_node]

        # under the hood, flytekit compiler defines and threads
        # together nodes within the `my_workflow` function body
        outputs = fn(*args, **kwargs)
        # ctx.compilation_state.nodes == [before_node, *nodes_created_by_fn]

        # defines the after node
        after_node = create_node(after)
        # ctx.compilation_state.nodes == [before_node, *nodes_created_by_fn, after_node]

        # compile the workflow correctly by making sure `before_node`
        # runs before the first workflow node and `after_node`
        # runs after the last workflow node.
        if ctx.compilation_state is not None:
            # ctx.compilation_state.nodes is a list of nodes defined in the
            # order of execution above
            workflow_node0 = ctx.compilation_state.nodes[1]
            workflow_node1 = ctx.compilation_state.nodes[-2]
            before_node >> workflow_node0
            workflow_node1 >> after_node
        return outputs

    if fn is None:
        return partial(setup_teardown, before=before, after=after)

    return wrapper

# %%
# There are a few key pieces to note in the ``setup_teardown`` decorator above:
#
# 1. It takes a ``before`` and ``after`` argument, both of which need to be ``@task``-decorated functions. These
#    tasks will run before and after the main workflow function body.
# 2. The :py:func:`~flytekit.core.node_creation.create_node` function to create nodes associated with the ``before``
#    and ``after`` tasks.
# 3. When ``fn`` is called, under the hood Flytekit creates all the nodes associated with the workflow function body
# 4. The code within the ``if ctx.compilation_state is not None:`` conditional is executed at compile time, which
#    is where we extract the first and last nodes associated with the workflow function body at index ``1`` and ``-2``.
# 5. Finally, we use the ``>>`` right shift operator to ensure that ``before_node`` executes before the
#    first node and ``after_node`` executes after the last node of the main workflow function body.

# %%
# Defining the DAG
# ^^^^^^^^^^^^^^^^
#
# Now let's define two tasks that will constitute the workflow

@task
def t1(x: float) -> float:
    return x - 1


@task
def t2(x: float) -> float:
    return x ** 2

# %%
# And then create our decorated workflow:

@workflow
@setup_teardown(before=setup, after=teardown)
def wf(x: float) -> float:
    return t2(x=t1(x=x))


if __name__ == "__main__":
    print(wf(x=10.0))


# %%
# In this example, you learned how to modify the behavior of a workflow by defining a ``setup_teardown`` decorator
# that can be applied to any workflow in your project. This is useful when integrating with other external services
# like `wandb <https://wandb.ai/site>`__ or `clearml <https://clear.ml/>`__, which enable you to track metrics of model
# training runs.
#
# If you want to define workflows imperatively, check out :ref:` this example <sphx_glr_auto_core_flyte_basics_imperative_wf_style.py>`,
# and to learn more about how to extend Flyte at a deeper level, for example creating custom types, custom tasks, or
# backend plugins, see :ref:`Extending Flyte <sphx_glr_auto_core_extend_flyte>`.
