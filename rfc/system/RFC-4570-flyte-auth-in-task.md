# [RFC Template] Flyte Authentication in Tasks

**Authors:**

- @zeryx

## 1 Executive Summary

This RFC proposes a method to flow Flyte authentication information into task containers. 
It aims to simplify access to execution-specific information and enable secure, streamlined usage of Flyte features within tasks. 
The proposal introduces a Flyte Auth Agent system to bind the authentication context of the task/workflow executor to all executed tasks.

## 2 Motivation

- In many workflows, accessing execution-specific information (e.g., used images, versions) is crucial. 
- Tasks often need to interact with Flyte for remote registration or to use Flyte features (e.g., pyflyte register, build, project creation).
- In order to access today, you're required to create a Kubernetes App, recording its ID and secret key, and then passing these to a Flyte Task. 
- This approach is either insecure (plaintext files) or overly complex (Kubernetes Secrets).
- Additionally, it requires the user to manually create and manage the App, which is a tedious process.

## 3 Proposed Implementation
![img.png](../img/flyte-auth-in-task-impl.png)

- **Flyte Auth Agent System:** Introduce a Flyte Auth Agent to bind the authentication of the task/workflow executor to all tasks.
- **Automatic Binding:** The bound authentication, possibly provided as mounted environment variables, will be automatically accessible within Flytekit.
- **Simplified Task Authorization:** Tasks will inherit the permissions of the workflow executor, ensuring a secure and streamlined permission model.
- **Enhanced Eager Mode:** Support for Eager Mode, enabling tasks to trigger downstream tasks securely without needing plaintext credentials or creating cloud-specific secrets.

### Client-side integration example
```python
from flytekit import task, remote
from typing import Optional

# No task arguments are required; this behavior would be automatic
@task
def get_execution_date(execution_id: str) -> Optional[str]:
    # The Flyte Auth Agent will automatically bind the authentication of the task executor to this context
    context = remote.default()
    # Behind the scenes, we would have a config.yaml (or similar) that would point to this Flyte deployment, and the agent would use that to authenticate
    execution = context.fetchExecution(execution_id)
    if execution.closure.phase == "SUCCEEDED":
        return execution.closure.started_at.isoformat()
    else:
        return None
```


### Client-side integration example (Eager Mode)

```python
from flytekit import task, remote, eager

@task
def say_hello(name: str) -> str:
    """A simple Flyte task to say "hello".

    The @task decorator allows Flyte to use this function as a Flyte task, which
    is executed as an isolated, containerized unit of compute.
    """
    return f"hello {name}!"

## No longer required to create a FlyteRemote, as it's automatically available within the task.
# remote = FlyteRemote(
#     config=Config.for_sandbox(),
#     default_project="flytesnacks",
#     default_domain="development",
# )

@eager(remote=remote.default())
async def hello_main(name: str = "flyte") -> str:
    results = await say_hello(name=name)
    return results


```

## 5 Drawbacks

- We now have a new Auth surface area that needs to be secured and maintained.
- New kubernetes infrastructure / services may need to be created to support the handling of task only authentication.
- We need to ensure that the Auth Agent is performant and does not introduce any significant overhead.

## 6 Alternatives

- Continue to use the existing method of passing credentials to tasks
  - This would include both adding plaintext credentials to the task container, attaching environment variables during registration; or  setting up AWS/GCP secrets.

## 7 Potential Impact and Dependencies
- **Integration with Existing Infrastructure:** The implementation of the Flyte Auth Agent will require integration with existing Flyte clusters and possibly modifications to the FlyteAdmin service for proper authentication handling.
- **Dependency on Kubernetes Features:** The automatic binding of authentication information might rely on specific Kubernetes features, such as Secrets or ConfigMaps, to securely transmit data to the task containers.
- **Client-side Flytekit Changes:** Adoption of this approach may necessitate updates to Flytekit, particularly in how it handles task execution context and authorization.

## 8 Security Considerations

- **Encryption and Security Protocols:** Ensuring that the transmission of authentication information from the Flyte Auth Agent to task containers is secure and encrypted.
- **Access Control:** Proper implementation of access control mechanisms to prevent unauthorized access to sensitive information or unauthorized execution of tasks.
- **Audit and Compliance:** Implementing logging and auditing capabilities to track the use of authentication tokens and access patterns for compliance and security analysis.

## 8 Unresolved questions

- **Integration into Flyte Backend** How will the Flyte Auth Agent be integrated into the Flyte backend? Will it be a separate service or part of an existing service?
- **Authentication Mechanisms:** What authentication mechanisms will be supported by the Flyte Auth Agent? Will it support multiple authentication mechanisms?
- **Task Execution Context:** How will the Flyte Auth Agent bind the authentication of the task executor to the task execution context? Will it use environment variables, Kubernetes Secrets, or some other mechanism?
- 
## 9 Conclusion

The introduction of the Flyte Auth Agent represents a significant step forward in enhancing the security, simplicity, and efficiency of task execution within the Flyte ecosystem. This proposal comes at a crucial time when the need for secure, scalable, and easy-to-manage authentication mechanisms is more pronounced than ever in the world of automated workflows and data processing.

The Flyte Auth Agent system addresses several key challenges currently faced by Flyte users, including the complexity of managing task-specific credentials, the security risks associated with handling plaintext credentials, and the operational overhead of managing Kubernetes secrets or external authentication systems. By automating the flow of authentication information into task containers.

## 10 RFC Process Guide, remove this section when done

*By writing an RFC, you're giving insight to your team on the direction you're taking. There may not be a right or better decision in many cases, but we will likely learn from it. By authoring, you're making a decision on where you want us to go and are looking for feedback on this direction from your team members, but ultimately the decision is yours.*

This document is a:

- thinking exercise, prototype with words.
- historical record, its value may decrease over time.
- way to broadcast information.
- mechanism to build trust.
- tool to empower.
- communication channel.

This document is not:

- a request for permission.
- the most up to date representation of any process or system

**Checklist:**

- [ ]  Copy template
- [ ]  Draft RFC (think of it as a wireframe)
- [ ]  Share as WIP with folks you trust to gut-check
- [ ]  Send pull request when comfortable
- [ ]  Label accordingly
- [ ]  Assign reviewers
- [ ]  Merge PR

**Recommendations**

- Tag RFC title with [WIP] if you're still ironing out details.
- Tag RFC title with [Newbie] if you're trying out something experimental or you're not entirely convinced of what you're proposing.
- Tag RFC title with [RR] if you'd like to schedule a review request to discuss the RFC.
- If there are areas that you're not convinced on, tag people who you consider may know about this and ask for their input.
- If you have doubts, ask on [#feature-discussions](https://slack.com/app_redirect?channel=CPQ3ZFQ84&team=TN89P6GGK) for help moving something forward.
