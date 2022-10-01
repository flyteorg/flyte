.. _community:

##########
Community
##########

Flyte is an ambitious open source project and would not be possible without an
amazing community. We are a completely open community and strive to treat
every member with respect. You will find the community welcoming and responsive,
so please join us on:

.. panels::
    :container: container-lg pb-4
    :column: col-lg-4 col-md-4 col-sm-4 col-xs-12 p-2
    :body: text-center

    .. link-button:: https://slack.flyte.org
       :type: url
       :text: Slack
       :classes: btn-block stretched-link

    :fa:`slack`

    ---

    .. link-button:: https://twitter.com/flyteorg
       :type: url
       :text: Twitter
       :classes: btn-block stretched-link

    :fa:`twitter`

    ---

    .. link-button:: https://github.com/flyteorg/flyte/discussions
       :type: url
       :text: Github Discussions
       :classes: btn-block stretched-link

    :fa:`github`

    ---

    .. link-button:: https://www.linkedin.com/groups/13962256
       :type: url
       :text: LinkedIn Group
       :classes: btn-block stretched-link

    :fa:`linkedin`
    
    ---

    .. link-button:: https://groups.google.com/a/flyte.org/d/forum/users
       :type: url
       :text: Google Group
       :classes: btn-block stretched-link

    :fa:`google`

    ---

    .. link-button:: https://blog.flyte.org/
       :type: url
       :text: Flyte Blog
       :classes: btn-block stretched-link

    :fa:`blog`


Also, feel free to sign up for our newsletter, Flyte Monthly, for a quick update on what we've been up to and upcoming events.

.. link-button:: https://www.getrevue.co/profile/flyte
    :type: url
    :text: Flyte Monthly
    :classes: btn-outline-secondary
    

Open Source Community Sync
--------------------------

We host an Open Source Community Sync every other Tuesday, 9:00 AM PDT/PST.
Check it out, subscribe to the `calendar <https://www.addevent.com/calendar/kE355955>`_, or just pop in!

.. link-button:: https://zoom.us/s/93875115830?pwd=YWZWOHl1ODRRVjhjVkxSV0pmZkJaZz09#success
    :type: url
    :text: Zoom Link
    :classes: btn-outline-secondary
    
If you'd like to give a 3-5 minute ⚡ Lightning Talk ⚡ during our Open Source Community Sync, or even suggest a topic for someone else to talk about, let us know!

.. link-button:: https://docs.google.com/forms/d/e/1FAIpQLSekwk2fieIVxmRuuNelbIp8DdXKe_SanlRguBtETbcSNHD11w/viewform
    :type: url
    :text: Lightning Talk Sign Up Sheet
    :classes: btn-outline-secondary

Office Hours
------------

Flyte maintainers are available on Zoom for you to ask anything, weekly on Wednesdays:

.. link-button:: https://www.addevent.com/event/zF10349020/
    :type: url
    :text: 7:00 am PT / 3:00 pm GMT
    :classes: btn-outline-secondary

.. link-button:: https://www.addevent.com/event/dQ10349168/
    :type: url
    :text: 9:00 pm PT / 10:30 am IST
    :classes: btn-outline-secondary

.. toctree::
    :caption: Community
    :maxdepth: -1
    :name: communitytoc
    :hidden:

    contribute
    roadmap
    Frequently Asked Questions <https://github.com/flyteorg/flyte/discussions/categories/q-a>
    troubleshoot
    
Flyte Documentation Style Guide
-------------------------------

Overview
^^^^^^^^

This guide is an overview of style guidelines for the documentation. 

.. note::
    These are guidelines, not rules. Use your best judgment, and feel free to propose changes to this document.
    
Documentation formatting standards
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Usage of upper camel case
_________________________
- When you refer specifically to interacting with an API object, use `UpperCamelCase <https://en.wikipedia.org/wiki/Camel_case>`_, also known as Pascal case. You may see different capitalization, such as "Flytectl", in the `API Reference <https://docs.flyte.org/en/latest/reference/index.html>`_. 

- When referring to any of the libraries in the Flyte ecosystem, prefer un-capitalized inline code markup format instead of capitalized plain-text format.

**Example:**

.. code-block:: text

    “`flytekit` is an SDK for Flyte” is preferred over “Flytekit is an SDK for Flyte”.

- When you are referring to an API object in general, `use sentence-style capitalization <https://docs.microsoft.com/en-us/style-guide/text-formatting/using-type/use-sentence-style-capitalization>`_.

**Example:**

.. code-block:: text

    Flyte UI is a web-based user interface for Flyte. 

- Don't split an API object name into separate words. For example, use “Flytectl”, not “Flyte ctl”.

- The following examples focus on capitalization. For more information about formatting API object names, review the related guidance on `Code Styling <https://docs.google.com/document/d/1rhEZqd3Ae8Y9HUxg7yF6LarHrfIHKnPyeke8aMSPgis/edit#heading=h.ei4vpl6rqhz3>`_.

Terminology casing convention:
``````````````````````````````
A list of specific terms and words to be used consistently across the documentation.

* flytectl
* Flyte sandbox
* Flytesnacks
* CLI
* FlyteAdmin
* FlytePickle
* flytekit
* FlyteConsole
* FlytePropeller
* workflow
* LaunchPlan
-We use ‘launch plan’ when we mention it generally, e.g., ‘we use launch plan to execute workflow,’ AND ‘we use LaunchPlan A to execute workflow B’).

Use angle brackets for placeholders
___________________________________
Tell the reader what a placeholder represents.

**Example:**

.. code-block:: text

    docker push <registry/repo:version>
    
Use bold for UI labels and keywords 
___________________________________
Emphasize UI labels with bold. A UI label is a piece of text that the user sees in a graphical or web browser’s user interface. Use bold in the doc to delineate the text that the user sees in the UI.
Examples: names of table columns, titles 
Emphasize a keyword when you first define it. A **keyword** is a word or expression with a single, special meaning in Flyte that is important for the user to know. It probably has a different meaning from its conventional meaning outside of Flyte.
Examples: tasks, workflows, launch plans, schedules, executions, workflow executions
Apply bold to a keyword only in the sentence where you define it. Do not apply bold to the keyword otherwise.
Do not use bold to emphasize something important. Instead, use a Note or Caution. 
In rst, use double asterisks to show text as **bold**, or strong.
**RST syntax:**
**Place content here**.

**Examples:**

.. code-block:: text

    - Select Graph in FlyteConsole to see a diagram of the workflow.
    - Executions are instances of workflows, nodes or tasks. You can schedule or request an execution. 

International standards for punctuation inside quotes
_____________________________________________________
Use punctuation inside the quotes while writing sentences using quotes.
**Example:** 

.. code-block:: text

    flytectl create project --id myflyteproject --name "My Flyte Project" --description "My very first project onboarding onto Flyte."
    
Starting a sentence with a component tool or component name
___________________________________________________________
* While starting a sentence with a tool or component name, prefer to use the right article before it.
* Use a general descriptor for a component name.
**Examples:**

.. code-block:: text

    - The myflyteapp directory comes with a sample workflow, which can be found under flyte/workflows/example.py.
    - The Java/Scala SDK for Flyte.
    - The @task decorator actually turns the decorated function into an instance of the PythonFunctionTask object.
    - A workflow can contain a single functional node.
    - ‘Flytectl allows you to communicate with FlyteAdmin using a YAML file or by passing every configuration value on the command-line’. (Donot use This allows you to …)

Use code style for String and Integer field
___________________________________________
For string or integer values, use style (single backticks for rst and double backticks for markdown).

**Examples:**

.. code-block:: text

    - Set the value of the field to True.

Code Styling
^^^^^^^^^^^^
Use code style for

* Filenames 
* Directories
* Paths
* Commands
* Inline code
* Object field names
* Namespaces
* API objects
* Anything that’s related to code!

**Examples:**

.. code-block:: text

    - Commands: flytectl --help,flyte sandbox start,`flytectl exec`
    - Filename: Open example.py in your favorite editor.
    - Directory: The docs for this repo are present in /cookbook/docs folder
    - Path: The myflyteapp directory comes with a sample workflow, which can be found under flyte/workflows/example.py.
    - Inline code: Use flytectl get objectName, `create launchplan my_launchplan`
    - Set value of `flytectl flag` to < >.
    - `flytectl- apiserver …`
    - The @task decorator actually turns the decorated function into an instance of the PythonFunctionTask object.
* Include a title for the code example.
* Use backtick (**`<inline code here>`**) for code style.
* Enclose code samples with triple backticks. (```).
* Use **..code-block::** a multi-line code block for fenced code blocks.

**Example for code block:**

.. code-block:: text

    ```bash
       Make html
    ```

    ..code-block:: python
      :caption: this .py
       :name: this-py
       print 'Explicit is better than implicit.'

* Use meaningful variable names that match the context.

Python: Snake case
Go: camelCase
HTML: camelCase 

* Remove trailing spaces in the code.
* Begin with Upper case for inline comments (wherever applicable).
* No period at the end of inline comments.
* Code documentation should be precise.
* Prefer active voice while explaining code.
* Avoid filler words.

Flytectl
--------
* In Go, error strings shouldn’t be capitalized or end with a punctuation or newline - else it results in a lint error.
* For list items in Go script, no period at the end of the sentence.
* For ‘println’ statements in Go, no period or exclamation mark at the end of the sentence.

Common phrases for introducing code samples
-------------------------------------------
* Let’s understand …
* Run the workflow locally.
* In xxx tutorial, we did …
* Import the necessary dependencies (or) Let’s import the required libraries.
* Define a xxx task to …
* Refer to xxx section for a detailed background on xxx

Test code locally and on sandbox
--------------------------------
When creating a new code example, test it locally, and on a sandbox. Attach the screenshot of the successful code execution/workflow on the sandbox. It is also desirable to use ‘pyflyte’ to run the code locally and ‘pyflyte run –remote’ (remotely) and provide the execution link to the UI. 

Code Snippet formatting
-----------------------
While adding the code snippets, include the command prompt syntax.

**Example:**

.. code-block:: text

    (venv)$ pyflyte --flyte.workflows package --image myapp:v1 --fast --force
    
Separate commands from output
-----------------------------


**Example:**
First, let's import the libraries we use in this example.

.. code-block:: python

    import os
    import pathlib
    
    from flytekit import Resources, kwtypes, workflow
    from flytekitplugins.papermill import NotebookTask

**Out:**

.. code-block:: text

    Output of multiplier(my_input=3.0):9.0
    Output of multiplier(my_input=0.5):1.0

Versioning
----------

* If the information is version-specific, it needs to be defined in the prerequisites section at the beginning of the page.
* Verify and archive examples/tutorials that are not relevant to the updated version.
* Avoid ambiguous comments on older/earlier versions.
* Mention alternative versions only if relevant. 

**Example:**

.. code-block:: text

    Prerequisites:
    Make sure you have Git, and Python >= 3.7 installed. Also ensure you have pip3 (mostly the case).

Literal Include
---------------
Use literalinclude to include a source file instead of copying and pasting its contents into the markdown.
You can also use literalinclude to insert files other than Python or README files.
To show certain configurations/preferences/settings to the user (instead of expecting the user to move back and forth between files), use literalinclude. This will specify the contents of the configuration file. Refer to the Sphinx docs to learn more.

**Example:**

.. literalinclude:: ../../../../../integrations/kubernetes/k8s_spark/Dockerfile
    :linenos:
    :emphasize-lines: 26-38
    :language: docker

Short codes
-----------

The documentation supports four different short codes:

* Note
* Warning
* Caution
* Tip

Use the syntaxes mentioned below accordingly.

Note
----

* Use **‘Note’** when you want to highlight a piece of information that may be  helpful to know.
* You can also use a **‘Note’** in a list. Refer `Common short code issues <https://docs.google.com/document/d/1rhEZqd3Ae8Y9HUxg7yF6LarHrfIHKnPyeke8aMSPgis/edit#heading=h.ezfwdgke7u3>`_.


**Syntax**

..note::
We recommend using an OSX or a Linux machine, as we have not tested this on windows. If you happen to test it, please let us know.

**Example:**


.. note::

    We recommend using an OSX or a Linux machine, as we have not tested this on windows. If you happen to test it, please let us know.
    
Caution
-------

* Do not use it. Use `‘Warning’<https://docs.google.com/document/d/1rhEZqd3Ae8Y9HUxg7yF6LarHrfIHKnPyeke8aMSPgis/edit#heading=h.gvinpnji63wc>`_ instead.

Tip
----

* It is used to give out extra information about certain aspects that would help the user understand/follow the concept better.
* Tip content is optional for the reader. Write content around the tip so that the user can safely ignore it.

**Syntax:**

..tip:: Before getting started, it is always important to measure the performance. Flyte project publishes and manages some grafana templates as described in - :ref:`deployment-cluster-config-monitoring`.

**Example:**

.. tip:: 

    Before getting started, it is always important to measure the performance. Flyte project publishes and manages some grafana templates as described in - :ref:`deployment-cluster-config-monitoring`.

Warning
-------

* Use ‘Warning’ to indicate danger or how crucial a piece of information is.
* The first sentence in the warning should tell the user what not to do. Subsequent sentences should describe the why, and link to more information (if any).
* You can also use a ‘Warning’ in a list. Refer Common short code issues.

**Syntax:**

..warning::
Don't use sandbox deployment for production environments. For an in-depth overview of how to productionize your Flyte deployment, checkout the :ref:`deployment` guide.

**Example:**

.. warning::
    Don't use sandbox deployment for production environments. For an in-depth overview of how to productionize your Flyte deployment, checkout the :ref:`deployment` guide.
    
Common shortcode issues
-----------------------

**List:**

* Indent 4 spaces or a tab before adding any short code within an ordered/numbered list.

