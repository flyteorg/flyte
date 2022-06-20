###########################
Contributing Guide
###########################

First off, thank you for thinking about contributing! 
Here are the instructions that will guide you through contributing, fixing, and improving Flytectl.

📝 Contribute to Documentation
==============================

Docs are generated using Sphinx and are available at [flytectl.rtfd.io](https://flytectl.rtfd.io).

To update the documentation, follow these steps:

1. Install the requirements by running ``pip install -r doc-requirements.txt`` in the root folder.
2. Make modifications in the `cmd <https://github.com/flyteorg/flytectl/tree/master/cmd>`__ folder.
3. Run ``make gendocs`` from within the `docs <https://github.com/flyteorg/flytectl/tree/master/docs>`__ folder.
4. Open html files produced by Sphinx in your browser to verify if the changes look as expected (html files can be found in the ``docs/build/html`` folder).

💻 Contribute Code
==================

1. Run ``make compile`` in the root directory to compile the code.
2. Set up a local cluster by running ``./bin/flytectl sandbox start`` in the root directory.
3. Run ``flytectl get project`` to see if things are working.
4. Run the command you want to test in the terminal.
5. If you want to update the command (add additional options, change existing options, etc.):

   * Navigate to `cmd <https://github.com/flyteorg/flytectl/tree/master/cmd>`__ directory
   * Each sub-directory points to a command, e.g., ``create`` points to ``flytectl create ...``
   * Here are the directories you can navigate to:

     .. list-table:: Flytectl cmd directories
        :widths: 25 25 50
        :header-rows: 1

        * - Directory
          - Command
          - Description
        * - ``config``
          - ``flytectl config ...``
          - Common package for all commands; has root flags
        * - ``configuration``
          - ``flytectl configuration ...``
          - Validates/generates Flytectl config
        * - ``create``
          - ``flytectl create ...``
          - Creates a project/execution
        * - ``delete``
          - ``flytectl delete ...``
          - Aborts an execution and deletes the resource attributes
        * - ``get``
          - ``flytectl get ...``
          - Gets a task/workflow/launchplan/execution/project/resource attribute
        * - ``register``
          - ``flytectl register ...``
          - Registers a task/workflow/launchplan
        * - ``sandbox``
          - ``flytectl sandbox ...``
          - Interacts with sandbox
        * - ``update``
          - ``flytectl update ...``
          - Updates a project/launchplan/resource attribute
        * - ``upgrade``
          - ``flytectl upgrade ...``
          - Upgrades/rollbacks Flytectl version
        * - ``version``
          - ``flytectl version ...``
          - Fetches Flytectl version

     Find all the Flytectl commands :ref:`here <nouns>`.
   * Run appropriate tests to view the changes by running ``go test ./... -race -coverprofile=coverage.txt -covermode=atomic  -v`` in the root directory.

