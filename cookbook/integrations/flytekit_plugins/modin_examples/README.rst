Modin
======

Modin is a pandas-accelerator that helps handle large datasets. 
Pandas works gracefully with small datasets since it is inherently single-threaded, and designed to work on a single CPU core. 
With large datasets, the performance of pandas drops (becomes slow or runs out of memory) due to single core usage. 
This is where Modin can be helpful.

Instead of optimizing pandas workflows for a specific setup, we can speed up pandas workflows by utilizing all the resources (cores) available in the system using the concept of ``parallelism``, which is possible through modin. `Here <https://modin.readthedocs.io/en/stable/getting_started/why_modin/pandas.html#scalablity-of-implementation>`__ is a visual representation of how the cores are utilized in case of Pandas and Modin.


Installation
------------

.. code:: bash

   pip install flytekitplugins-modin


How is Modin different?
-----------------------

Modin **scales** the Pandas workflows by changing only a **single line of code**.

The plugin supports the usage of Modin DataFrame as an input to and output of a task/workflow, similar to how a pandas DataFrame can be used.