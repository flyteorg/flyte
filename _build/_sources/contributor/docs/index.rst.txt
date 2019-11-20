.. _contribute-docs:

######################
Contributing to Docs
######################

Documentation for Flyte spans a few repos, and is hosted on GitHub Pages.  The contents of the ``docs/`` folder on the ``master`` branch are hosted on GitHub Pages.  See `GH Pages documentation <https://help.github.com/en/articles/configuring-a-publishing-source-for-github-pages>`_ for more information.

Documentation that the steps below will compile come from:

* This repository, comprising written RST pages
* RST files generated from the Flyte IDL repository
* RST files generated from the Flytekit Python SDK

Building
**********

In order to create this set of documentation run::

    $ make update_ref_docs && make all_docs

What happens is:
* ``./generate_docs.sh`` runs.  All this does is create a temp directory and clone the two aforementioned repos.
* The Sphinx builder container will run with files from all three repos (the two cloned one and this one) mounted.

  * It will generate RST files for Flytekit from the Python source code
  * Copy RST files from all three repos into a common location
  * Build HTML files into the ``docs/`` folder

Please then commit the newly generated files before merging your PR.  In the future we will invest in CI to help with this.

Notes
*******
We aggregate the doc sources in a single ``index.rst`` file so that ``flytekit`` and the Flyte IDL HTML pages to be together in the same index/table of contents.

There will be no separate Flyte Admin, Propeller, or Plugins documentation generation.  This is because we have a :ref:`contributor` guide and high level usage/architecture/runbook documentation should either there, or into the administrator's guide.
