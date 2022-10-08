.. _code_Styling:

######################
Code Styling
######################

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
