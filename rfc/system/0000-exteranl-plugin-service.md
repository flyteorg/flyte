# Out of Core Plugin [RFC]

## Motivation
Currently, it is hard to implement backend plugins, especially for data-scientists & MLE’s who do not have working knowledge of Golang. Also, performance requirements, maintenance and development is cumbersome.

The document here proposes a path to make it possible to write plugins rapidly, while decoupling them from the core flytepropeller engine.

## Goals
* Plugins should be easy to author - no need of code generation, using tools that MLEs and Data Scientists are not accustomed to using.
* Most important plugins for Flyte today are plugins that communicate with external services.
* It should be possible to test these plugins independently and also deploy them privately.
* It should be possible for users to use backend plugins for local development, especially in flytekit and unionML
* It should be possible to author plugins in any language.
* Plugins should be scalable
* Plugins should have a very simple API
* Plugins should show up in the UI (extra details)

## Overview
**Flytekit Backend Plugin** is a plugin registry written by FastAPI. Both users and Propeller can send the request to it to run the jobs (Bigquery, Databricks). Moreover, this registry should be stateless, so we can easily to scale up to multiple machines.

![](https://i.imgur.com/3FQcW99.png)

![](https://i.imgur.com/9xBeEri.jpg)



**FlytePropeller Backend System Plugin**: New web api plugin used to send the request to flytekit backend service to create a job on an external platform.

### Create Job
```sequence
Propeller->Grpc API Server: REST (Task spec)
Grpc API Server->BigQuery: REST (BQ job Spec)
BigQuery->Grpc API Server: Job Id
Grpc API Server->Propeller: Job Id
```
### Poll Job
```sequence
Propeller->Async Cache: Create Cache
Async Cache->Grpc API Server: REST (Job Id)
Grpc API Server->BigQuery: REST (Job Id)
BigQuery->Grpc API Server: Job State, Message
Grpc API Server->Async Cache: Job State, Message
Propeller->Async Cache: Fetch Cache
```

### Register a backend plugin
Users should be able to write the plugin using a rest-like interface. In this proposal we recommend using a simplified form of WebAPI plugins. the goal of the plugins is to enable the following functions

It should be  possible to implement this using a Web Service framework like `FastAPI`. Raw implementation for a `fastAPI` like server can be as follows. 

**Note**: It should be possible to implement multiple **resource** plugins per python library. Thus each resource should be delimited.

```python
from fastapi import FastAPI
from flytekit.backend.plugins import TaskInfo, IO, PollResultSuccess,
               PollResultFailure, OperationInProgress, PollResult

app = FastAPI()

@app.put("/plugins/v1/{task_type}/{version}")
def create(task_info: TaskInfo) -> ResourceInfo:
    """
    Creates resource.
    
    return job_id, error code, message
    200 Job was created.
    401 The request was unauthorized.
    500 The request was not handled correctly due to a server error.
    """
    return {"job_id": job_id, "error_code": error_code, "message": message}

@app.get("/plugins/v1/{task_type}/{version}")
async def poll(resource_id: int, task_info: TaskInfo, io: IO, resource_info: ResourceInfo) -> PollResult:
    """
    Retrieves resource and returns the current observable status
    
    return job_spec, job_state (succeed, failed, running), error core, message.
    200 Job was retrieved successfully.
    400 The request was malformed. See JSON response for error details.
    401 The request was unauthorized.
    500 The request was not handled correctly due to a server error.
    """
    if resp.state == "success":
       state = succeed
    elif resp.state == "failure":
       state = failed
    elif resp.state == "running":
        state = running
    return {"job_spec": job_spec, "job_state": state, "error_code": error_code, "message": message}

@app.delete("/plugins/v1/{task_type}/{version}")
async def delete_res(resource_id: str):
    """
    Delete the resource
    
    return job_id, error core, message.
    200 Job was deleted successfully.
    400 The request was malformed. See JSON response for error details.
    401 The request was unauthorized.
    500 The request was not handled correctly due to a server error.
    """
    return {"job_id": job_id, "error_code": error_code, "message": message}

```
**Ideally**, we should provide a simplified interface to this in flytekit. This would mean that the flytekit plugin can simply use this interface to create a local plugin and a backendplugin.

```python
class BackendPluginBase:

    def __init__(self, task_type: str, version: str = "v1"):
        self._task_type = task_type
        self._version = version

    @property
    def task_type(self) -> str:
        return self._task_type

    @property
    def version(self) -> str:
        return self._version

    @abstractmethod
    async def create(self):
        pass

    @abstractmethod
    async def poll(self):
        pass
    
    @abstractmethod
    async def delete(self):
        pass
    
BackendPluginRegistry.register(BQPlugin())
```

### (Alternative) GRPC Backend Server
```python
class BackendPluginServer(BackendPluginServiceServicer):
    def CreateTask(self, request: plugin_system_pb2.TaskCreateRequest, context) -> plugin_system_pb2.TaskCreateResponse:
        req = TaskCreateRequest.from_flyte_idl(request)
        plugin = BackendPluginRegistry.get_plugin(req.template.type)
        return plugin.create(req.inputs, req.output_prefix, req.template)

    def GetTask(self, request: plugin_system_pb2.TaskGetRequest, context):
        plugin = BackendPluginRegistry.get_plugin(request.task_type)
        return plugin.get(job_id=request.job_id, output_prefix=request.output_prefix, prev_state=request.prev_state)

    def DeleteTask(self, request: plugin_system_pb2.TaskDeleteRequest, context):
        plugin = BackendPluginRegistry.get_plugin(request.task_type)
        return plugin.delete(request.job_id)
```
# Register a New plugin


### Run a Flytekit Backend Plugin
The workflow code running Snowflake, Databricks should not be changed. We can add a new config `enable_backend_flytekit_plugin_system`. If it's true, the job will be handled by plugin system. If it can't find the plugin in the registry, it falls back to use the default web api plugin in the propeller.

### Secrets Management
Secrets management should not be imposed using Flyte’s convention, though we should provide a simplified secrets reader using flytekit secret system?

## Backend Plugin System Deployment
There are two options to deploy backend plugin system. 

### 1. All the plugin running in one deployment
Pros: 
* Easy to maintain.
* One image, and one deployment yaml file.

Cons: 
* Dependency conflict between two plugins.
* Need to restart the deployment and rebuild the image when we add a new backend plugin to flytekit.

### 2. One plugin per deployment

Pros:
* We can scale up the specific plugin deployments when the requests increase in some plugins.
* Only need to deploy the plugins that people will use.

Cons:
* Hard to maintain. (tons of images, and deployment yaml files)

### Deployment yaml
Backend plugin system can be run in the deployment independently. Because it's stateless, we can just scale up the replica if request increases.

This yaml file can be added into the Flyte helm chart (`flyte-core`).

```ymal
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend-plugin-system
  labels:
    app: backend-plugin-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: backend-plugin-system
  template:
    metadata:
      labels:
        app: backend-plugin-system
    spec:
      containers:
      - name: backend-plugin-system
        image: pingsutw/backend-plugin-system:v1
        ports:
        - containerPort: 8000

---
apiVersion: v1
kind: Service
metadata:
  name: backend-plugin-system
spec:
  selector:
    app: backend-plugin-system
  ports:
    - protocol: TCP
      port: 8000
      targetPort: 8000
```
### Docker Image
```dockerfile
FROM python:3.9-slim-buster

MAINTAINER Flyte Team <users@flyte.org>
LABEL org.opencontainers.image.source=https://github.com/flyteorg/flytekit

WORKDIR /root
ENV PYTHONPATH /root

RUN apt-get update && apt-get install -y git
RUN pip install "git+https://github.com/flyteorg/flytekit@backend-plugin-system"
RUN pip install fastapi uvicorn[standard]
RUN pip install numpy==1.23.1

CMD uvicorn flytekit.extend.backend.fastapi:app --host 0.0.0.0 --port 8000
```

## Authentication
- [Authentication with FastAPI](https://www.freecodecamp.org/news/how-to-add-jwt-authentication-in-fastapi/)
- [Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)

## Phase 1 (POC)
- **Flytekit**: Use Fast api to create a backend plugin service that can submit the job to Bigquery or Databricks.
- **FlytePropeller**: Add a web api plugin that can talk to flytekit backend plugin service.

## Phase 2
- **Authentication** - only propeller and users having access token can submit the job to backend plugin system.
- **Deployment**: Add backend plugin system deployment to the helm chart.
- **Benchmark**: Measure the overhead, and improve the performance.

## Open Question:

1. Should we provide a way to deploy these plugins using Flyte? IMO this should be optional and not required for MVP?
2. We could implement something similar for golang


## Benchmark
Compare the memory and CPU consumption of the FastAPI server with that of the current Propeller server.
![](https://i.imgur.com/EXXmdyt.png)

The round latency between grpc and FastAPI server.
![](https://i.imgur.com/hp8wss4.png =50%x)

The round latency between grpc and current propeller.
![](https://i.imgur.com/42mdktu.png =50%x)

