# How do I yield a dynamic workflow?

If all you need to do is to run a **dynamic_task** that will yield an arbitrary number of tasks, or arbitrary kinds of tasks, depending on the values of some input, then all you need are plain dynamic tasks. The contents of this example are not what you are looking for.

This example is for those rare instances where you would like to add specify changes to the structure of tasks itself. That is, if you want some **task_X** to depend on the output of some **task_Y** sometimes, but have **task_X** depend on **task_Z** and not **task_Y** some other times, then this should help you out. Think carefully however - users have gone down this road before only to realize later that this wasn't what they were looking for.


### Examples

### SimpleDynamicSubworkflow

This is the simplest example to just get the reader started. It demonstrates that a **dynamic_task** can generate an entirely new workflow, unregistered against Flyte Admin (that is, you won't be able to see it/launch it as a standalone workflow from the UI).


### InverterDynamicWorkflow

This workflow is a bit more complicated. The workflow has one more input, called **inverter_input**, a boolean. If not set, then the task that will be run as part of the yielded workflow is an identity task. If set, the inverter task is used.



