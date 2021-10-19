##############################
Flytesnacks Contribution Guide
##############################

First off, thank you for thinking about contributing! 
Below you’ll find instructions that will hopefully guide you through how to contribute to and improve Flytesnacks.

💻 Contribute to Examples
=========================

1. Determine where to put your new code
   
   * `Core <https://github.com/flyteorg/flytesnacks/tree/master/cookbook/core>`__: Contains examples that demonstrate functionality available within core flytekit. These examples should be runnable locally.
   * `Integrations <https://github.com/flyteorg/flytesnacks/tree/master/cookbook/integrations>`__: Contains examples that leverage one or more of the available plugins.
   * `Case Studies <https://github.com/flyteorg/flytesnacks/tree/master/cookbook/case_studies>`__: Contains examples that demonstrate the usage of Flyte to solve real-world problems. These are generally more complex examples that may require extra setup or that can only run on larger clusters.
       
2. Create a directory (applicable for ``integrations`` and ``case_studies`` directories)

   After determining where to put your example, create a directory under the appropriate parent directory. Each example
   directory should contain:

   * Dockerfile
   * Makefile
   * README.rst
   * __init__.py
   * requirements.in
   * sandbox.config

   It might be easier to copy one of the existing examples and modify it to your needs.
3. Add the example to CI

   Examples are references in `this github workflow <https://github.com/flyteorg/flytesnacks/blob/master/.github/workflows/ghcr_push.yml>`__.
   Add a new entry under ``strategy -> matrix -> directory`` with the name of your directory as well as its relative path. 
   Also, add the example to `flyte_tests_manifest.json <https://github.com/flyteorg/flytesnacks/tree/master/cookbook/flyte_tests_manifest.json>`__.
4. Test your code!
    * If the Python code can be run locally, just use ``python <my file>`` to run it.
    * If the Python code has to be tested in a cluster:
        * Install flytectl by running ``brew install flyteorg/homebrew-tap/flytectl``. Learn more about installation and configuration of flytectl `here <https://docs.flyte.org/projects/flytectl/en/latest/index.html>`__.
        * Run the ``flytectl sandbox start --source=$(pwd)`` command in the directory that's one level above the directory that has Dockerfile. 
          For example, to register `house_price_prediction <https://github.com/flyteorg/flytesnacks/tree/master/cookbook/case_studies/ml_training/house_price_prediction>`__ example, run the start command in the ``ml_training`` directory. 
          To register ``core`` examples, run the start command in the ``cookbook`` directory. So, ``cd`` to the required directory and run all the upcoming commands in there!

          Following are the commands to run if examples in ``core`` directory are to be tested on sandbox:
            1. Build Docker container using the command: ``flytectl sandbox exec -- docker build . --tag "core:v1" -f core/Dockerfile``. 
            2. Package the examples by running ``pyflyte --pkgs core package --image core:v1 -f``.
            3. Register the examples by running ``flytectl register files --archive -p flytesnacks -d development --archive flyte-package.tgz --version v1``.
            4. Visit https://localhost:30081/console to view the Flyte console, which consists of the examples present in the ``flytesnacks/cookbook/core`` directory.
            5. To fetch new dependencies and rebuild the image, run 
               ``flytectl sandbox exec -- docker build . --tag "core:v2" -f core/Dockerfile``, 
               ``pyflyte --pkgs core package --image core:v2 -f``, and 
               ``flytectl register files --archive -p flytesnacks -d development --archive flyte-package.tgz --version v2``.
            6. Refer to `Build & Deploy Your Application “Fast”er! <https://docs.flyte.org/en/latest/getting_started_iterate.html#bonus-build-deploy-your-application-fast-er>`__ if code in itself is updated and requirements.txt is the same.

📝 Contribute to Documentation
==============================

``docs`` folder in ``flytesnacks`` houses the required documentation configuration. All the core, case_studies, and integrations docs are written in the respective README.md and the Python code files. 

1. README.md file needs to capture the *what*, *why*, and *how* 
    * What is the integration about? Its features, etc.
    * Why do we need this integration? How is it going to benefit the Flyte users?
    * Showcase the uniqueness of the integration
    * How to install the plugin?
  
    Refer to any repo in the cookbook directory to understand this better.

2. Explain what the code does  
3. Update `conf.py <https://github.com/flyteorg/flytesnacks/tree/master/cookbook/docs/conf.py>`__ (imagine you have added ``snowflake`` to the ``integrations/external_services`` folder):
   
      * Add the Python file names to the ``CUSTOM_FILE_SORT_ORDER`` list
      * Add ``../integrations/external_services/snowflake`` to ``example_dirs``
      * Add ``auto/integrations/external_services/snowflake`` to ``gallery_dirs``
4. Verify if the code and documentation look as expected
   
   1. Learn about the documentation tools `here <https://docs.flyte.org/en/latest/community/contribute.html#documentation>`__
   2. Install the requirements by running ``pip install -r docs-requirements.txt`` in the ``cookbook`` folder
   3. Run ``make html`` in the ``docs`` folder

   .. tip::
      For implicit targets, run ``make -C docs html``.
   4. Open HTML pages present in the ``docs/_build`` directory in the browser
   
5. After creating the pull request, check if the docs are rendered correctly by clicking on the documentation check 
   
   .. image:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flytesnacks/contribution-guide/test_docs_link.png
       :alt: Docs link in a PR