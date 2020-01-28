Golang Test Targets
~~~~~~~~~~~~~~~~~~~

Provides an ``install`` make target that uses ``go mod`` to install golang dependencies.

Provides a ``lint`` make target that uses golangci to lint your code.

Provides a ``test_unit`` target for unit tests.

Provides a ``test_unit_cover`` target for analysing coverage of unit tests, which will output the coverage of each function and total statement coverage.

Provides a ``test_unit_visual`` target for visualizing coverage of unit tests through an interactive html code heat map.

Provides a ``test_benchmark`` target for benchmark tests.

**To Enable:**

Add ``lyft/golang_test_targets`` to your ``boilerplate/update.cfg`` file.

Make sure you're using ``go mod`` for dependency management.

Provide a ``.golangci`` configuration (the lint target requires it).

Add ``include boilerplate/lyft/golang_test_targets/Makefile`` in your main ``Makefile`` _after_ your REPOSITORY environment variable

::

    REPOSITORY=<myreponame>
    include boilerplate/lyft/golang_test_targets/Makefile

(this ensures the extra make targets get included in your main Makefile)
