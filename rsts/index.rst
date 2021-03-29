

.. toctree::
   :maxdepth: 1
   :name: mainsections
   :hidden:

   getting_started
   Tutorials <https://flytecookbook.readthedocs.io>
   reference/index

.. toctree::
   :caption: How-Tos
   :maxdepth: 1
   :name: howtotoc
   :hidden:

   plugins/index
   howto/index

.. toctree::
   :caption: Deep Dive
   :maxdepth: 1
   :name: divedeeptoc
   :hidden:

   dive_deep/index

.. toctree::
   :caption: Contributor Guide
   :maxdepth: 1
   :name: roadmaptoc
   :hidden:

   community/index
   community/docs
   community/roadmap
   community/compare

Meet Flyte
==========

Flyte is an open-source, container-native, structured programming and distributed processing platform. It enables highly concurrent, scalable and maintainable workflows for machine learning and data processing.

Created at `Lyft <https://www.lyft.com/>`__, Flyte provides first class support for Python, Java, and Scala, and is built directly on Kubernetes for all the benefits containerization provides: portability, scalability, and reliability.

The core unit of execution in Flyte is the ``task``, which you can easily write with the Flytekit Python SDK:

.. code:: python
   
   @task
   def greet(name: str) -> str:
       return f"Welcome, {name}!"

You can compose one or more tasks to create a ``workflow``:

.. code:: python

   @task
   def add_question(greeting: str) -> str:
       return f"{greeting} How are you?"
   
   @workflow
   def welcome(name: str) -> str:
       greeting = greet(name=name)
       return add_question(greeting)

   welcome("Traveler")
   # Output: "Welcome, Traveler! How are you?"


Why Flyte?
==========

Flyte's main purpose is to increase the development velocity for data processing and machine learning, enabling large-scale compute execution without the operational overhead. Teams can therefore focus on the business goals and not the infrastructure.

Core Features
-------------

* Container Native
* Extensible Backend & SDK’s
* Ergonomic SDK’s in Python, Java & Scala
* Versioned & Auditable - all actions are recorded
* Matches your workflow - Start with one task, convert to a pipeline, attach multiple schedules or trigger using a programmatic API or on-demand
* Battle-tested - millions of pipelines executed per month
* Vertically-Integrated Compute - serverless experience
* Deep understanding of data-lineage & provenance
* Operation Visibility - cost, performance, etc.
* Cross-Cloud Portable Pipelines

Who's Using Flyte?
------------------

At `Lyft <https://eng.lyft.com/introducing-flyte-cloud-native-machine-learning-and-data-processing-platform-fb2bb3046a59>`__, Flyte has served production model training and data processing for over four years, becoming the de-facto platform for the Pricing, Locations, ETA, Mapping teams, Growth, Autonomous and other teams.

Next Steps
----------

Whether you want to write Flyte workflows, deploy the Flyte platform to your k8 cluster, or extend and contribute to the architecture and design of Flyte, we have what you need.

* :ref:`Get Started <gettingstarted>`
* :ref:`Main Concepts <divedeep>`
* :ref:`Extend Flyte <plugins>`
* :ref:`Join the Community <community>`
