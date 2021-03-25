Flyte
=====

.. image:: images/flyte_lockup_gradient_on_light.png

.. toctree::
   :maxdepth: 1
   :caption: Getting Started
   :name: gettingstartedtoc
   :hidden:

   getting_started/first_run
   getting_started/first_example
   getting_started/run_on_flyte
   getting_started/learn

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
   reference/index

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

Created at Lyft, Flyte provides first class support for Python, Java, and Scala, and is built directly on Kubernetes for all the benefits containerization provides: portability, scalability, and reliability.

Flyte provides a single unit of execution (task) as a top-level concept. Multiple tasks arranged in a data producer-consumer order create the workflow, which is pure specification created in any language.

Why Flyte?
==========

Flyte's main purpose is to increase the development velocity for data processing and machine learning, enabling large-scale compute execution without the operational overhead. Teams can therefore focus on the business goals and not the infrastructure.

What makes Flyte different?
---------------------------

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

At Lyft, Flyte has served production model training and data processing for over four years, becoming the de-facto platform for the Pricing, Locations, ETA, Mapping teams, Growth, Autonomous and other teams

Whether you will be writing Flyte workflows, deploying the Flyte platform to your k8 cluster, or would like to extend and contribute to the architecture and design of Flyte, we have what you need.

Welcome to the documentation hub for Flyte!

Beginners: see :ref:`getting-started-firstrun`

Intermediate Users: see :ref:`plugins`

Advanced Users: see :ref:`divedeep`
