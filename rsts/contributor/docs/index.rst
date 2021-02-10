.. _contribute-docs:

######################
Contributing to Docs
######################

Documentation for Flyte spans a few repos, and is hosted on GitHub Pages.  The contents of the ``docs/`` folder on the ``master`` branch are hosted on GitHub Pages.  See `GH Pages documentation <https://help.github.com/en/articles/configuring-a-publishing-source-for-github-pages>`_ for more information.

Documentation that the steps below will compile come from:

* This repository, comprising written RST pages
* RST files generated from the Flyte IDL repository
* RST files generated from the Flytekit Python SDK

*****************************************************
Building all documentation including dependent repos
*****************************************************

In order to create this set of documentation run::

    $ make generate-docs

What happens is:
  * ``./generate_docs.sh`` runs.  All this does is create a temp directory and clone the two aforementioned repos.
  * The Sphinx builder container will run with files from all three repos (the two cloned one and this one) mounted.
  * It will generate RST files for Flytekit from the Python source code
  * Copy RST files from all three repos into a common location
  * Build HTML files into the ``docs/`` folder

Please then commit the newly generated files before merging your PR.  In the future we will invest in CI to help with this.

******************************************************************
Building a local copy of documentation for RST modifications only
******************************************************************
This can be used if one wants to quickly verify documentation updates in the rst files housed in the flyte repo. These are the main docs that flyte publishes.
Building all documentation can be a slow process. To speed up the iteration one can use the make target::
  
  $ make generate-local-docs

This needs a local virtual environment with sphinx-build installed.


*******
Notes
*******
We aggregate the doc sources in a single ``index.rst`` file so that ``flytekit`` and the Flyte IDL HTML pages to be together in the same index/table of contents.

There will be no separate Flyte Admin, Propeller, or Plugins documentation generation.  This is because we have a :ref:`contributor` guide and high level usage/architecture/runbook documentation should either there, or into the administrator's guide.


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

    Even though the ``sphobjinv`` CLI returns ``:py:function:...``, when actually creating a link you should just use ``:py:func:...``. See `this <https://www.sphinx-doc.org/en/master/usage/restructuredtext/domains.html#cross-referencing-python-objects>`__.

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

    Task: :std:doc:`flytekit:tasks`

Task: :std:doc:`flytekit:tasks`

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


