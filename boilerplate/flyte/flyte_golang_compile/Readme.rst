Flyte Golang Compile
~~~~~~~~~~~~~~~~~~~~

Common compile script for Flyte golang services.

**To Enable:**

Add ``flyteorg/flyte_golang_compile`` to your ``boilerplate/update.cfg`` file.

Add the following to your Makefile

::

  .PHONY: compile_linux
  compile_linux:
    PACKAGES={{ *your packages }} OUTPUT={{ /path/to/output }} ./boilerplate/flyte/flyte_golang_compile.sh
