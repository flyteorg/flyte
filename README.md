<html>
    <p align="center"> 
        <img src="https://github.com/flyteorg/flyte/blob/master/rsts/images/flyte_circle_gradient_1_4x4.png" alt="Flyte Logo" width="100">
    </p>
    <h1 align="center">
        Flyte User Guide & Tutorials
    </h1>
    <p align="center">
        Flytesnacks encompasses code examples built on top of Flytekit Python
    </p>
    <h3 align="center">
        <a href="https://docs.flyte.org/projects/cookbook/en/latest/index.html">User Guide</a>
        <span> · </span>
        <a href="https://docs.flyte.org/projects/cookbook/en/latest/tutorials.html">Tutorials</a>
        <span> · </span>
        <a href="#contribution-guide">Contribution Guide</a>
    </h3>
</html>

<html>
    <h2 id="quick-start"> 
        🚀 Quick Start
    </h2>
</html>

> To get the hang of Python SDK, refer to the [Getting Started](https://docs.flyte.org/en/latest/getting_started.html) tutorial before exploring the examples.

[User Guide](https://docs.flyte.org/projects/cookbook/en/latest/index.html) section has code examples, tips, and tricks that showcase the usage of Flyte features and integrations.

[Tutorials](https://docs.flyte.org/projects/cookbook/en/latest/tutorials.html) section has real-world examples, ranging from machine learning training, data processing to feature engineering.

<html>
    <h2 id="contribution-guide"> 
        📖 Contribution Guide
    </h2>
</html>

Flytesnacks currently has all examples in Python (Flytekit Python SDK). In the future, Java examples employing Flytekit JAVA will be out.

Here are the setup instructions to start contributing to `flytesnacks` repo:

### 💻 Contribute to examples

1. Determine where to put your new code:
  * [Core](https://github.com/flyteorg/blob/master/cookbook/core)
    
    Contains examples that demonstrates functionality available within core flytekit. These examples should be runnable
    locally.
    
  * [Integrations](https://github.com/flyteorg/blob/master/cookbook/integrations)

    Contains examples that leverage one or more of the available plugins.

  * [Case Studies](https://github.com/flyteorg/blob/master/cookbook/case_studies)

    Contains examples that demonstrates usage of Flyte to solve real world problems. These are more generally more complex
    examples that may require extra setup or that can only run on larger clusters..

2. Create a directory
   
   After determining where to put your example, create a directory under the appropriate parent directory. Each example
   directory should contain:
   
   * Dockerfile
   * Makefile
   * README.rst
   * __init__.py
   * requirements.in
   * sandbox.config
    
   It might be easier to copy one of the existing examples and modify it to your needs.

3. Add it to CI

   Examples are references in [this github workflow](https://github.com/flyteorg/flytesnacks/blob/master/.github/workflows/ghcr_push.yml)
   Add a new entry under ``strategy -> matrix -> directory`` with the name of your directory as well as its relative path.
   
4. Test your code!
    * If the Python code can be run locally, just use `python <my file>` to run it.
    * If the Python code has to be tested in a cluster:
        * Install [flytectl](https://docs.flyte.org/projects/flytectl/en/stable/) by running `brew install flyteorg/homebrew-tap/flytectl`. 
        * Run the `make start` command in the root directory of the flytesnacks repo
        * Visit https://localhost:30081 to view the Flyte console consisting of the examples present in flytesnacks/cookbook/core directory
        * To fetch new dependencies and rebuild the image, run `make register`
        * If examples from a different directory (other than `core`) have to registered, run the command `SANDBOX=1 make -C <directory> register` in the `flytesnacks` repo.

### 📝 Contribute to documentation

`docs` folder in `flytesnacks` houses the user guide and tutorials present in the documentation. Refer to the [documentation contribution guide](https://docs.flyte.org/en/latest/community/contribute.html#documentation) to get acquainted with the guidelines.

<html>
    <h2 id="file-an-issue"> 
        🐞 File an Issue
    </h2>
</html>

Refer to the [issues](https://docs.flyte.org/en/latest/community/contribute.html#issues) section in the contribution guide if you'd like to file an issue relating to `flytesnacks` code or documentation.
