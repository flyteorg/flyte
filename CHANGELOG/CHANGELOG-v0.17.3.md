# Flyte v0.17.3

## Platform
1.  Native scheduler
2.  Support for Snowflake including backend plugin and flytekit tasks
3.  Expose default MaxParallelism in flyteadmin conter
4.  Support custom resource cleanup policy in backend plugins
5.  Improved error message in the case of images with invalid names


## flytekit
1.  Tuples are now supported as the return type
2.  Duplicate tasks names are now caught at serialization time.
3.  Support flag to specify config file
4.  Bug fixes:
    1.  Configuration generated via `flyte-cli setup-config --insecure` no longer raise an error
    2.  Improved detection of remote files


## UI
1.  Additional information when a Task in a non-terminal state
2.  Support for workflow versions


## flytectl
1.  Sandbox docker images can now be provided as a parameter
2.  Bug fixes:
    -   panics in calls to get execution details and launchplans.
    -   datetime format generated in execFile are now valid
