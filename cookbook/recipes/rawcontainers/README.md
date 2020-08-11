[Back to Cookbook Menu](../..)

# Run arbitrary containers without *flytekit* (java or python)


Since Flyte is container native and specification driven, there is no need for users to have flytekit as a dependency in their containers.
Flytekit makes it easier for users to author workflows and provides a lot of tools and systems to optimize their flow. But, this is not necessary.
It is possible to use a local script with a open source container (or custom built container), and still access the type system of Flyte. 

Recipes in this section provide some powerful patterns that this could expose. Also remember, you can always mix and match, some tasks could be written using
flytekit natively (as described in other sections) and some tasks could be written using Raw container support.

## Recipe 1: Write a workflow where tasks are shell scripts
Small shell scripts (limited by the max size of a command) can be written in-line in Flyte Raw containers and coupled with open source containers, this can make it 
extremely powerful to write data processing using shell scripts.

In this contrived example we are simply performing a multi-step (expensive) sum of squares. Since we are using open sourced containers, you can completely run
these examples in any notebook environment (or even the python repl) as long as you can access FlyteAdmin.

Peruse and replay the notebook to try out an example
[Exhibit: raw-container-shell.ipynb](raw-container-shell.ipynb)

## Recipe 2: Use Raw containers to perform rapid iteration
Now that you know we can pass in data into arbitrary containers and read outputs from the arbitrary containers, we can leverage this to build a task where the actual execution code is passed in as an argument.

[Exhibit: hot-loading.ipynb](hot-loading.ipynb)

## NOTES
 - This is an Alpha feature
 - All data is passed into the container using the file-system. So it is necessary to have a file like system available
 - Containers cannot have processes like init.d that need to be pid-1
 - Exact specification of inputs and outputs is WIP, and may change slightly as we progress through beta etc. But, the raw specification is available [here](https://github.com/lyft/flyteidl/blob/master/protos/flyteidl/core/tasks.proto#L159-L225)
 - Please contact us in the Slack channel if you have questions, want to use it and are having trouble using it. 

