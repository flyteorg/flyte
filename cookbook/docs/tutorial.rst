.. _flyte-tutorial:
.. currentmodule:: flyte_tutorial

############################################
Tutorial: Hello World!
############################################

.. rubric:: Estimated time to complete: 3 minutes.

The goal of this tutorial is to modify an existing example (hello world) to take the name of the person to green as an argument.

Prerequisites
*************

* Ensure that you have completed the `Getting Started <https://docs.flyte.org/en/latest/tutorials/first_run.html>`__ guide.

* Ensure that you have `git <https://git-scm.com/>`__ installed.

Steps
*****

1. First install the python Flytekit SDK and clone the ``flytesnacks`` repo:

.. prompt:: bash

  pip install flytekit>=0.16.0b6
  git clone git@github.com:flyteorg/flytesnacks.git flytesnacks
  cd flytesnacks


2. Open `cookbook/core/basic/hello_world.py` in your favorite editor.

3. Add `name: str` as an argument to both `my_wf` and `say_hello` functions. Then update the body of `say_hello` to consume that argument.

4. Update the simple test at the bottom of the file to pass in a name. E.g. `print(f"Running my_wf(name="adam") {my_wf(name="adam")}")`

5. Run this file in Python:

.. prompt:: bash
  python cookbook/core/basic/hello_world.py

  It should output `hello world, adam`.

Congratulations! You have just ran your first workflow. Let's now run it on the sandbox cluster you deployed earlier.

6. Run:

.. prompt:: bash

  make fast_register

7. Visit `the console <http://localhost:30081/console/projects/flytesnacks/domains/development/workflows/core.basic.hello_world.my_wf>`__, click launch, and enter your name.

8. Give it a minute and one it's done, check out "Inputs/Outputs" on the top right corner to see your greeting updated.