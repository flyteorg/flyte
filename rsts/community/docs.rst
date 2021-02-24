.. _contribute-docs:

######################
Contributing to Docs
######################

************************
Docs for various repos
************************
Flyte is a large project and all the docs span multiple repositories. The core of the documention is in the `flyteorg/flyte <https://github.com/flyteorg/flyte>`_ repository.
Flyte uses `Sphinx <https://www.sphinx-doc.org/en/master/>`_ to compile it docs. Docs are automatically pushed on merge to master and docs are hosted using `readthedocs.org <https://readthedocs.org/>`_

***************
Sphinx and RST
***************

Style
=========

Headers
--------
Typically, we try to follow these characters in this order for heading separation.

.. code-block:: text

    # with overline
    * with overline
    =
    -
    ^

Intersphinx
=============
`Intersphinx <https://www.sphinx-doc.org/en/master/usage/extensions/intersphinx.html>`__ is a plugin that all Flyte repos that build Sphinx documentation use for cross-referencing with each other. There's some good background information on it on these `slides <https://docs.google.com/presentation/d/1vkvsxp_64dhFuf7g3W8EHjK77lFOBhdeSg9VSA9-Ikc/>`__.

Inventory File
----------------
When Sphinx runs, an inventory file gets created and is available alongside each repo's HTML pages. For example at ``https://readthedocs.org/projects/flytecookbook/objects.inv``. This file is a compressed inventory of all the sections, tags, etc. in the flyte cookbook documentation. This inventory file is what allows the intersphinx plugin to cross-link between projects.

There is an open-source tool called ``sphobjinv`` that has managed to `reverse engineer these files <https://sphobjinv.readthedocs.io/en/stable/syntax.html>`__, and offers a CLI to help search for things inside them.

Setup
-------
Installing ``sphobjinv`` and simple usage ::

    $ pip install sphobjinv

    # Using the CLI to query a hosted inventory file, note the -u switch
    $ sphobjinv suggest https://flytekit.readthedocs.io/en/latest/ -u task

    No inventory at provided URL.
    Attempting "https://flytekit.readthedocs.io/en/latest/objects.inv" ...
    Remote inventory found.

    :py:function:`flytekit.task`
    :std:doc:`tasks`
    :std:doc:`tasks.extend`
    :std:label:`tasks:tasks`

    # Using the CLI to query a local file, useful when iterating locally
    $ sphobjinv suggest ~/go/src/github.com/flyteorg/flytekit/docs/build/html/objects.inv task

    :py:function:`flytekit.task`
    :std:doc:`tasks`
    :std:doc:`tasks.extend`
    :std:label:`tasks:tasks`

.. note::

    Even though the ``sphobjinv`` CLI returns ``:py:function:...``, when actually creating a link you should just use ``:py:func:...``. See `this <https://www.sphinx-doc.org/en/master/usage/restructuredtext/domains.html#cross-referencing-python-objects>`__. Here is a quick list of mappings

    .. list-table:: Conversion table for - ``sphobjinv``
        :widths: 50 50
        :header-rows: 1

        * - What the tool returns?
          - What you should use instead?
        * - :py:module:
          - :py:mod:
        * - :py:function:
          - :py:func:
        * - :std:label:
          - :std:ref:
        * - :py:method:
          - :py:meth:



Linking Examples
------------------
In the ``conf.py`` file of each repo, there is an intersphinx mapping argument that looks something like this ::

    intersphinx_mapping = {
        "python": ("https://docs.python.org/3", None),
        "flytekit": ("https://flyte.readthedocs.io/projects/flytekit/en/master/", None),
        ...
    }

This file is what tells the plugin where to look for these inventory files, and what project name to refer to each inventory file as. The project name is important because they're used when actually referencing something from the inventory.

Here are some examples, first the code and then the link

.. code-block:: text

    Task: :std:doc:`generated/flytekit.task`

Task: :std:doc:`generated/flytekit.task`

-----

.. code-block:: text

    :std:doc:`Using custom words<flytekit:tasks>`

:std:doc:`Using custom words<flytekit:tasks>`

Python
^^^^^^^
Linking to Python elements changes based on what you're linking to. Check out this `section <https://www.sphinx-doc.org/en/master/usage/restructuredtext/domains.html#cross-referencing-python-objects>`__. For instance linking to the ``task`` decorator in flytekit uses the ``func`` role.

.. code-block:: text

    Link to flytekit code :py:func:`flytekit:flytekit.task`

Link to flytekit code :py:func:`flytekit:flytekit.task`

Other elements use different Sphinx roles, here are some examples using Python core docs. ::

    :py:mod:`Module <python:typing>`
    :py:class:`Class <python:typing.Type>`
    :py:data:`Data <python:typing.Callable>`
    :py:func:`Function <python:typing.cast>`
    :py:meth:`Method <python:pprint.PrettyPrinter.format>`


:py:mod:`Module <python:typing>`

:py:class:`Class <python:typing.Type>`

:py:data:`Data <python:typing.Callable>`

:py:func:`Function <python:typing.cast>`

:py:meth:`Method <python:pprint.PrettyPrinter.format>`
