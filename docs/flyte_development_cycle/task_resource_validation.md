# Task resource validation

In Union, when you attempt to execute a workflow with unsatisfiable resource requests, we fail the execution immediately rather than allowing it to queue forever.

We intercept execution creation requests in executions service to validate that their resource requirements can be met and fast-fail if not. A failed validation returns a message similar to

```{code-block} text
Request failed with status code 400 rpc error: code = InvalidArgument desc = no node satisfies task 'workflows.fotd.fotd_directory' resource requests
```

While we have ongoing work to improve user-facing information here, we may need to help debug why the execution could not placed.

Using logs
Ensure that executions service logs are configured for debug level

Find failing 400 requests using this query. Update the env (i.e., production) and time range as needed. Can also add a filter for the tenant (i.e., flawlessai)

Choose the one you care about and copy the request id from the message -- the long hex string at the end, i.e., 0c195ae8ce3aded38d90777c9dc33382

Copy that into this query. This will filter for all logs corresponding to that request in executions service (which handles validation). You should find debug messages indicating what constraints could not be met by various nodegroups.