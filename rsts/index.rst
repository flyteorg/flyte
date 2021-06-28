

.. toctree::
   :maxdepth: 1
   :name: mainsections
   :titlesonly:
   :hidden:

   |plane| Getting Started <getting_started>
   |book-reader| User Guide <https://docs.flyte.org/projects/cookbook/en/latest/user_guide.html>
   |chalkboard| Tutorials <https://docs.flyte.org/projects/cookbook/en/latest/tutorials.html>
   |project-diagram| Concepts <concepts/basics>
   |book| API Reference <reference/index>
   |hands-helping| Community <community/index>

.. toctree::
   :caption: Concepts
   :maxdepth: -1
   :name: divedeeptoc
   :hidden:

   concepts/basics
   concepts/control_plane
   concepts/architecture

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

   References <reference/index>

   
Meet Flyte
==========

.. raw:: html

   <p style="color: #808080; font-weight: 350; font-size: 25px; padding-top: 10px; padding-bottom: 10px;"> The workflow automation platform for complex, mission-critical data and ML processes at scale </p>


Flyte is an open-source, container-native, structured programming and distributed processing platform. It enables highly concurrent, scalable and maintainable workflows for machine learning and data processing.

Created at `Lyft <https://www.lyft.com/>`__ in collaboration with Spotify, Freenome and many others, Flyte provides first class support for Python, Java, and Scala, and is built directly on Kubernetes for all the benefits containerization provides: portability, scalability, and reliability.

The core unit of execution in Flyte is the ``task``, which you can easily write with the Flytekit Python SDK:

.. tabs::

    .. tab:: python

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
               return add_question(greeting=greeting)

           welcome("Traveler")
           # Output: "Welcome, Traveler! How are you?"

    .. tab:: java

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

        You can compose one or more tasks to create a ``workflow``:

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
                SdkBindingData name = builder.inputOfString("name", "The name for the welcome message");

                SdkBindingData greeting = builder.apply("greet", GreetTask.of(name)).getOutput("greeting");
                SdkBindingData greetingWithQuestion = builder.apply("add-question", AddQuestionTask.of(greeting)).getOutput("greeting");

                builder.output("greeting", greetingWithQuestion, "Welcome message");
              }
            }

    .. tab:: scala

        .. code:: scala

            case class GreetTaskInput(name: String)
            case class GreetTaskOutput(greeting: String)

            /** Example Flyte task that takes a name as the input and outputs a simple greeting message */
            class GreetTask
                extends SdkRunnableTask(
                  SdkScalaType[GreetTaskInput],
                  SdkScalaType[GreetTaskOutput]
                ) {

              /**
               * Defines task behavior. This task takes a name as the input, wraps it in a welcome message, and
               * outputs the message.
               *
               * @param input the name of the person to be greeted
               * @return the welcome message
               */
              override def run(input: GreetTaskInput): GreetTaskOutput = GreetTaskOutput(s"Welcome, ${input.name}!")
            }

            object GreetTask {
              /**
               * Binds input data to this task
               *
               * @param name the input name
               * @return 1 transformed instance of this class with input data
               */
              def apply(name: SdkBindingData): SdkTransform =
                new GreetTask().withInput("name", name)
            }

        You can compose one or more tasks to create a ``workflow``:

        .. code:: scala

            case class AddQuestionTaskInput(greeting: String)
            case class AddQuestionTaskOutput(greeting: String)

            /**
             * Example Flyte task that takes a greeting message as input, appends "How are you?", and outputs
             * the result
             */
            class AddQuestionTask
                extends SdkRunnableTask(
                  SdkScalaType[AddQuestionTaskInput],
                  SdkScalaType[AddQuestionTaskOutput]
                ) {

              /**
               * Defines task behavior. This task takes a greeting message as the input, append it with " How
               * are you?", and outputs a new greeting message.
               *
               * @param input the greeting message
               * @return the updated greeting message
               */
              override def run(input: AddQuestionTaskInput): AddQuestionTaskOutput = AddQuestionTaskOutput(s"${input.greeting} How are you?")
            }

            object AddQuestionTask {
              /**
               * Binds input data to this task
               *
               * @param greeting the input greeting message
               * @return a transformed instance of this class with input data
               */
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

For the most current list of Flyte's deployments, please click `here <https://github.com/flyteorg/flyte#%EF%B8%8F-current-deployments>`_.

Next Steps
----------

Whether you want to write Flyte workflows, deploy the Flyte platform to your k8 cluster, or extend and contribute to the architecture and design of Flyte, we have what you need.

* :ref:`Get Started <gettingstarted>`
* :ref:`Main Concepts <divedeep>`
* :ref:`Extend Flyte <cookbook:plugins_extend>`
* :ref:`Join the Community <community>`
