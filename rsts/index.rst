.. toctree::
   :maxdepth: 1
   :name: mainsections
   :titlesonly:
   :hidden:

   |plane| Getting Started <getting_started/index>
   |book-reader| User Guide <https://docs.flyte.org/projects/cookbook/en/latest/user_guide.html>
   |chalkboard| Tutorials <https://docs.flyte.org/projects/cookbook/en/latest/tutorials.html>
   |project-diagram| Concepts <concepts/basics>
   |rocket| Deployment <deployment/index>
   |book| API Reference <reference/index>
   |hands-helping| Community <community/index>

.. toctree::
   :caption: Getting Started
   :maxdepth: -1
   :name: gettingstarted
   :hidden:

   getting_started/index

.. toctree::
   :caption: Concepts
   :maxdepth: -1
   :name: divedeeptoc
   :hidden:

   concepts/basics
   concepts/control_plane
   concepts/architecture

.. toctree::
   :caption: Deployment
   :maxdepth: -1
   :name: deploymenttoc
   :hidden:

   deployment/index

.. toctree::
   :caption: Community
   :maxdepth: -1
   :name: roadmaptoc
   :hidden:

   Join the Community <community/index>

.. toctree::
   :caption: API Reference
   :maxdepth: -1
   :name: apireference
   :hidden:

   API Reference <reference/index>

   
What is Flyte?
==============

.. raw:: html

   <p style="color: #808080; font-weight: 350; font-size: 25px; padding-top: 10px; padding-bottom: 10px;"> The workflow automation platform for complex, mission-critical data and machine learning processes at scale. </p>


Flyte is an open-source, container-native, structured programming and distributed processing platform implemented in Golang. It enables highly concurrent, scalable and maintainable workflows for machine learning and data processing.

Created at `Lyft <https://www.lyft.com/>`__ in collaboration with Spotify, Freenome, and many others, Flyte provides first-class support for Python, Java, and Scala. It is built directly on Kubernetes for all the benefits of containerization like portability, scalability, and reliability.


The core unit of execution in Flyte is a :ref:`task <divedeep-tasks>`, which you can easily write with the Flytekit SDK:

.. tabbed:: Python

    .. code:: python

        from flytekit import task, workflow

        @task
        def greet(name: str) -> str:
            return f"Welcome, {name}!"

    You can compose one or more tasks to create a :ref:`workflow <divedeep-workflows>`:

    .. code:: python

        @task
        def add_question(greeting: str) -> str:
            return f"{greeting} How are you?"

        @workflow
        def welcome(name: str) -> str:
            greeting = greet(name=name)
            return add_question(greeting=greeting)

        welcome(name="Traveller")
        # Output: "Welcome, Traveller! How are you?"

.. tabbed:: Java

    .. code:: java

        @AutoService(SdkRunnableTask.class)
        public class GreetTask extends SdkRunnableTask<GreetTask.Input, GreetTask.Output> {
          public GreetTask() {
            super(JacksonSdkType.of(Input.class), JacksonSdkType.of(Output.class));
          }

          public static SdkTransform of(SdkBindingData name) {
            return new GreetTask().withInput("name", name);
          }

          @AutoValue
          public abstract static class Input {
            public abstract String name();
          }

          @AutoValue
          public abstract static class Output {
            public abstract String greeting();

            public static Output create(String greeting) {
              return new AutoValue_GreetTask_Output(greeting);
            }
          }

          @Override
          public Output run(Input input) {
            return Output.create(String.format("Welcome, %s!", input.name()));
          }
        }

    You can compose one or more tasks to create a :ref:`workflow <divedeep-workflows>`:

    .. code:: java

        @AutoService(SdkRunnableTask.class)
        public class AddQuestionTask extends SdkRunnableTask<AddQuestionTask.Input, AddQuestionTask.Output> {
          public AddQuestionTask() {
            super(JacksonSdkType.of(Input.class), JacksonSdkType.of(Output.class));
          }

          public static SdkTransform of(SdkBindingData greeting) {
            return new AddQuestionTask().withInput("greeting", greeting);
          }

          @AutoValue
          public abstract static class Input {
            public abstract String greeting();
          }

          @AutoValue
          public abstract static class Output {
            public abstract String greeting();

            public static Output create(String greeting) {
              return new AutoValue_AddQuestionTask_Output(greeting);
            }
          }

          @Override
          public Output run(Input input) {
            return Output.create(String.format("%s How are you?", input.greeting()));
          }
        }

    .. code:: java

        @AutoService(SdkWorkflow.class)
        public class WelcomeWorkflow extends SdkWorkflow {

          @Override
          public void expand(SdkWorkflowBuilder builder) {
            // defines the input of the workflow
            SdkBindingData name = builder.inputOfString("name", "The name for the welcome message");

            // uses the workflow input as the task input of the GreetTask
            SdkBindingData greeting = builder.apply("greet", GreetTask.of(name)).getOutput("greeting");

            // uses the output of the GreetTask as the task input of the AddQuestionTask
            SdkBindingData greetingWithQuestion =
                builder.apply("add-question", AddQuestionTask.of(greeting)).getOutput("greeting");

            // uses the task output of the AddQuestionTask as the output of the workflow
            builder.output("greeting", greetingWithQuestion, "Welcome message");
          }
        }

    Link to the example code: `WelcomeWorkflow.java <https://github.com/flyteorg/flytekit-java/blob/5cd638af1b131450e1dcf44b268112fca1f8de57/flytekit-examples/src/main/java/org/flyte/examples/WelcomeWorkflow.java>`_

.. tabbed:: Scala

    .. code:: scala

        case class GreetTaskInput(name: String)
        case class GreetTaskOutput(greeting: String)

        class GreetTask
            extends SdkRunnableTask(
              SdkScalaType[GreetTaskInput],
              SdkScalaType[GreetTaskOutput]
            ) {

          override def run(input: GreetTaskInput): GreetTaskOutput = GreetTaskOutput(s"Welcome, ${input.name}!")
        }

        object GreetTask {
          def apply(name: SdkBindingData): SdkTransform =
            new GreetTask().withInput("name", name)
        }

    You can compose one or more tasks to create a :ref:`workflow <divedeep-workflows>`:

    .. code:: scala

        case class AddQuestionTaskInput(greeting: String)
        case class AddQuestionTaskOutput(greeting: String)

        class AddQuestionTask
            extends SdkRunnableTask(
              SdkScalaType[AddQuestionTaskInput],
              SdkScalaType[AddQuestionTaskOutput]
            ) {

          override def run(input: AddQuestionTaskInput): AddQuestionTaskOutput = AddQuestionTaskOutput(s"${input.greeting} How are you?")
        }

        object AddQuestionTask {
          def apply(greeting: SdkBindingData): SdkTransform =
            new AddQuestionTask().withInput("greeting", greeting)
        }

    .. code:: scala

        class WelcomeWorkflow extends SdkWorkflow {

          def expand(builder: SdkWorkflowBuilder): Unit = {
            // defines the input of the workflow
            val name = builder.inputOfString("name", "The name for the welcome message")

            // uses the workflow input as the task input of the GreetTask
            val greeting = builder.apply("greet", GreetTask(name)).getOutput("greeting")

            // uses the output of the GreetTask as the task input of the AddQuestionTask
            val greetingWithQuestion = builder.apply("add-question", AddQuestionTask(greeting)).getOutput("greeting")

            // uses the task output of the AddQuestionTask as the output of the workflow
            builder.output("greeting", greetingWithQuestion, "Welcome message")
          }
        }

    Link to the example code: `WelcomeWorkflow.scala <https://github.com/flyteorg/flytekit-java/blob/5cd638af1b131450e1dcf44b268112fca1f8de57/flytekit-examples-scala/src/main/scala/org/flyte/examples/flytekitscala/WelcomeWorkflow.scala>`_



Why Flyte?
==========

Flyte aims to increase the development velocity for data processing and machine learning applications and enable large-scale compute execution without the operational overhead. Teams can, therefore, focus on the business goals rather than the infrastructure.

Core Features
-------------

* Container Native
* Reproducibility
* Extensible Backend and SDKs
* Ergonomic SDKs in Python, Java and Scala
* Versioned and Auditable - all actions are recorded
* Matches your workflow - Start with one task, define workflows, and launchplans, convert to a pipeline, attach multiple schedules or trigger using a programmatic API or on-demand
* Battle-tested - millions of pipelines executed per month
* Vertically-Integrated Compute - serverless experience
* Deep understanding of data-lineage and provenance
* Operation Visibility - cost, performance, etc.
* Cross-Cloud Portable Pipelines
* Support for cross language workflows

Who's Using Flyte?
------------------

At `Lyft <https://eng.lyft.com/introducing-flyte-cloud-native-machine-learning-and-data-processing-platform-fb2bb3046a59>`__, Flyte has served production model training and data processing for over four years, becoming the de-facto platform for the Pricing, Locations, ETA, Mapping teams, Growth, Autonomous and other teams.

For the most current list of Flyte's deployments, click `here <https://github.com/flyteorg/flyte#%EF%B8%8F-current-deployments>`_.

Next Steps
----------

Whether you want to write Flyte workflows, deploy the Flyte platform to your K8s cluster, or extend and contribute to the architecture and design of Flyte, we have what you need.

* :ref:`Get Started <getting-started>`
* :ref:`Main Concepts <divedeep>`
* :ref:`Extend Flyte <cookbook:plugins_extend>`
* :ref:`Join the Community <community>`