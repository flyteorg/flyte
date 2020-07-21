# Interactive development and debugging with Flyte

In this recipe we will talk about how to interact with an existing Workflow or execution. This is not a tutorial for developing a Workflow interactively.

Alright, so we have a workflow or a task that is already registered. A registered workflow/task refers to those that were passed to register workflow or task. FlyteAdmin is aware of the name of the
Workflow or task and they can be retrieved using flyte-cli. So before get get started with the following steps, lets register workflows and tasks in this directory.

## Part 0: Execute tasks locally
The method `unit_test` is implicitly defined on a Flyte task. This can be invoked from any interactive or programmatic medium.
It is useful in writing programmatic unit tests

### Jupyter Notebook
[Exhibit: part_0.ipynb](part_0.ipynb)

## Part 1: Executing a Pre-registered task
Once the registration is successful, task executions can be independently launched

### Jupyter Notebook
[Exhibit: part_1.ipynb # Secion I](part_1.ipynb)

### CLI (flyte-cli)
```bash
 $ flyte-cli -h localhost:30081 -i list-task-versions -p flytesnacks -d development --name recipes.interaction.interaction.scale
 # Get the URN
 $ flyte-cli -h localhost:30081 -i launch-task -p flytesnacks -d development -u <urn> -- image=https://miro.medium.com/max/1400/1*qL8UYfaStcEo_YVPrA4cbA.png
```

### Console
**Not yet supported**

*Currently launching a task from the console is not Supported*


## Part 2:  Executing a Pre-registered Workflow / LaunchPlan

### Jupyter Notebook
[Exhibit: part_1.ipynb # Section II](part_1.ipynb)

### CLI (flyte-cli)
```bash
$ flyte-cli -h localhost:30081 -i list-launch-plan-versions -p flytesnacks -d development --name recipes.interaction.interaction.FailingWorkflow
  Launch Plan Versions Found for flytesnacks:development:recipes.interaction.interaction.FailingWorkflow

Version                                            Urn                                                                              Schedule                       Schedule State
ade80023b74f9810fe720471b8926d6b991fc879           lp:flytesnacks:development:recipes.interaction.interaction.FailingWorkflow:ade80023b74f9810fe720471b8926d6b991fc879 

$ flyte-cli -h localhost:30081 -i execute-launch-plan -p flytesnacks -d development -u lp:flytesnacks:development:recipes.interaction.interaction.FailingWorkflow:ade80023b74f9810fe720471b8926d6b991fc879 -r <username> -- image=https://miro.medium.com/max/1400/1*qL8UYfaStcEo_YVPrA4cbA.png 
```

### Console
1. Navigate to **Console Homepage > Flyte Examples | development > recipes.interaction.interaction.FailingWorkflow**
    - If Using console on localhost sandbox (docker for mac mostly then)
      [Workflow Link](http://localhost:30081/console/projects/flytesnacks/domains/development/workflows/recipes.interaction.interaction.FailingWorkflow)
2. Click on the Launch Button

## Part 3: Retrieving outputs of error from a previous execution

### Jupyter Notebook
[Exhibit: part_1.ipynb # Section III](part_1.ipynb)

### CLI (flyte-cli)
1. Lets try to see how we can list all executions

```bash
 $ flyte-cli -h localhost:30081 -i list-executions -p flytesnacks -d development
```
 - if we want to see only the failed executions then, use *--filter* directive.
```bash
 $ flyte-cli -h localhost:30081 -i list-executions -p flytesnacks -d development -f 'eq(phase,FAILED)'
```
 - To get all executions not failed or succeeded, *you can and the filters by specifying multiple -f*

2. Lets try to retrieve the results using cli
```bash
 $ flyte-cli -h localhost:30081 -i get-execution -u ex:flytesnacks:development:f696cca32ffa44513b25
Using default config file at /Users/kumare/.flyte/config
Welcome to Flyte CLI! Version: 0.10.4

Execution flytesnacks:development:f696cca32ffa44513b25

	State:          FAILED
	Launch Plan:    lp:flytesnacks:development:recipes.interaction.interaction.FailingWorkflow:ade80023b74f9810fe720471b8926d6b991fc879
Error:
		Code: USER:Unknown
		Message:
			Traceback (most recent call last):				      File "/opt/venv/lib/python3.6/site-packages/flytekit/common/exceptions/scopes.py", line 206, in user_entry_point		        return wrapped(*args, **kwargs)		      File "/root/recipes/interaction/tasks.py", line 34, in rotate		        raise Exception("User signaled failure")				Message:				    User signaled failure				User error.

	Node Executions:

		ID: scale-task

			Status:         SUCCEEDED
			Started:        1970-01-01 00:00:00+00:00
			Duration:       0:00:00
			Input:          s3://my-s3-bucket/metadata/propeller/flytesnacks-development-f696cca32ffa44513b25/scale-task/data/inputs.pb
			Output:         s3://my-s3-bucket/metadata/propeller/flytesnacks-development-f696cca32ffa44513b25/scale-task/data/0/outputs.pb

			Task Executions:

				Attempt 0:

					Created:        2020-07-20 01:16:52.940739
					Started:        1970-01-01 00:00:00
					Updated:        2020-07-20 01:16:52.940739
					Duration:       0:00:00
					Status:         SUCCEEDED
					Logs:           (None Found Yet)


		ID: rotate-task

			Status:         FAILED
			Started:        2020-07-20 01:16:53.817785+00:00
			Duration:       0:00:31.571665
			Input:          s3://my-s3-bucket/metadata/propeller/flytesnacks-development-f696cca32ffa44513b25/rotate-task/data/inputs.pb
Error:
				Code: USER:Unknown
				Message:
					Traceback (most recent call last):				      File "/opt/venv/lib/python3.6/site-packages/flytekit/common/exceptions/scopes.py", line 206, in user_entry_point		        return wrapped(*args, **kwargs)		      File "/root/recipes/interaction/tasks.py", line 34, in rotate		        raise Exception("User signaled failure")		Message:				    User signaled failure				User error.

			Task Executions:

				Attempt 0:

					Created:        2020-07-20 01:16:53.719516
					Started:        1970-01-01 00:00:00
					Updated:        2020-07-20 01:17:25.325091
					Duration:       0:00:00
					Status:         FAILED
					Logs:

						Name:    Kubernetes Logs (User)
						URI:     http://localhost:30082/#!/log/flytesnacks-development/f696cca32ffa44513b25-rotate-task-0/pod?namespace=flytesnacks-development

Error:
						Code: USER:Unknown
						Message:
							Traceback (most recent call last):				      File "/opt/venv/lib/python3.6/site-packages/flytekit/common/exceptions/scopes.py", line 206, in user_entry_point		        return wrapped(*args, **kwargs)		      File "/root/recipes/interaction/interaction.py", line 34, in rotate		        raise Exception("User signaled failure")				Message:				    User signaled failure				User error.

``` 

### Console
1. Navigate to **Console Homepage > Flyte Examples | development > recipes.interaction.interaction.FailingWorkflow**
    - If Using console on localhost sandbox (docker for mac mostly then)
      [Workflow Link](http://localhost:30081/console/projects/flytesnacks/domains/development/workflows/recipes.interaction.interaction.FailingWorkflow)
2. find the execution (or filter using the UI dropdowns)
3. The main list page should show the execution failure reason. If you want to dig deeper, click on the execution for more information.

## Part 4: Relaunching a failed execution

### Jupyter Notebook
[Exhibit: part_1.ipynb # Section IV](part_1.ipynb)

### CLI (flyte-cli)
This CLI has a special function called relaunch
```bash
 $ flyte-cli -h localhost:30081 -i relaunch-execution -u ex:flytesnacks:development:f696cca32ffa44513b25 -- angle=20
```

### Console
1. Navigate to **Console Homepage > Flyte Examples | development > recipes.interaction.interaction.FailingWorkflow**
    - If Using console on localhost sandbox (docker for mac mostly then)
      [Workflow Link](http://localhost:30081/console/projects/flytesnacks/domains/development/workflows/recipes.interaction.interaction.FailingWorkflow)
2. Find the execution and click on it
3. On top right corner we have a "relaunch button". This replaces the terminate button and is available only if an execution exits.

## Part 5: Launch a new execution using partial outputs of a previous execution
### Jupyter Notebook
Only jupyter way is the easy to use way. This is possible to do through the UI, but that would be copy pasting the previous outputs.

[Exhibit: part_2.ipynb](part_2.ipynb)
