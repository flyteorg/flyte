# [WIP] Support Decentralized Custom Agent Configuration Management

**Authors:**
- @shuyingliang

## 1. Executive Summary

While the Flyte agent framework provides an efficient and productive way for users to develop Flyte plugins, all agents must currently be configured within the FlytePropeller configuration, with the config map mounted under `/etc/flyte/config`. 

In an enterprise environment, different teams are responsible for developing custom agents and managing Flyte components like Propeller, Scheduler, Admin, and Console. These teams often have their own CI/CD pipelines to deploy custom agents. The centralized agent configuration process can lead to inefficiencies and errors, as the deployment of custom agents may fall out of sync with the centralized configuration or cause unnecessary operational cost. 

This RFC proposes support for decentralized custom agent configuration management, with benefits including:
- **Separation of concerns**: Teams can *fully* independently manage their own custom agent configurations.
- **Decentralized management**: Custom agents can be deployed and managed separately from core Flyte infrastructure.
- **Flexibility**: FlytePropeller can dynamically adapt to varying configurations without requiring centralized maintenance of all settings.

## 2. Motivation

In enterprise environments, the team managing the core Flyte infrastructure is usually different from the teams developing custom agents/plugins. These teams operate with separate development cycles and CI/CD pipelines. For instance, a team writing its own custom agent would create its own service account, Docker image, and deploy it in a separate namespace with its own monitoring dashboard.

However, the final step of deployment—updating the Propeller configuration—remains centralized. This leads to issues like:
1. **Out-of-sync deployment**: When a custom agent endpoint needs to be updated, it requires coordination between teams for the deployment of both the agent and the Propeller configuration update, leading to potential mismatches.
2. **Coordination overhead**: Misaligned rollouts could result in the Propeller being unable to reach the agent, potentially causing widespread issues across Flyte plugins.

To address these challenges and fully enable the decentralized custom agent framework, this RFC proposes a decentralized approach to custom agent configuration management, allowing:
- **Separation of concerns**: Teams can manage their own configurations independently.
- **Decentralized management**: Custom agents can be deployed and managed separately from Flyte’s core infrastructure.
- **Flexibility**: FlytePropeller can adapt to various configurations dynamically.

## 3. Proposed Implementation

Currently, Flyte [parses configuration files under `/etc/flyte/config` and decodes them](https://github.com/flyteorg/flyte/tree/master/flytestdlib/config/viper) for use later. In decentralized management, the custom agent configuration might not be accessible in the same volume or mount, meaning traditional parsing from the Propeller pod's filesystem won't be possible. 

Instead, the configuration could be managed dynamically via Kubernetes APIs, allowing FlytePropeller to discover, validate, and merge custom agent configurations at runtime.

### Discovery of Custom Agent ConfigMaps
Implement a mechanism to systematically discover and collect custom agent ConfigMaps across namespaces or predefined locations. This ensures any new custom agent configurations are properly detected and processed.

For example, it is not unreasonable to ask the custom agent team to create configmap named with specific pattern, such as `flyteplugins-agentservice-` so that the propeller can list the config maps in the cluster via K8s API.

### Validation of ConfigMaps
Before using any custom agent configuration, each ConfigMap should undergo automated validation. If a ConfigMap fails validation, it will be excluded without impacting other valid configurations. This ensures that invalid configurations do not prevent FlytePropeller from starting.

This can be achieved by implementing a K8s validating webhook. For example, the information for a custom agent is only needed as follows


```yaml
apiVersion: v1
king: ConfigMap
name: flyteplugins-agentservice-config-xxx
data:
    plugins:
      agent-service:
        agents:
          custom-agent:
           endpoint: custom_agent_end_point_FQDN:8000
           insecure: true
        supportedTaskTypes:
        - custom_sensor
        - customtask
```

```yaml
apiVersion: v1
king: ConfigMap
name: flyteplugins-agentservice-enabled-plugins-xxx
data:
    plugins:
      agent-service:
        supportedTaskTypes:
        - customsensor
        - customtask
    tasks:
      task-plugins:
        default-for-task-types:
          customtask: agent-service
          customsensor: agent-service
```

### Aggregation and Conflict Resolution
After validation, valid ConfigMaps should be aggregated. The system must handle potential conflicts (e.g., two agents with the same name or endpoint) by either resolving them or flagging them for review. This ensures that only conflict-free configurations are applied, isolating issues and preventing one bad agent configuration from affecting the entire Flyte setup.

## 4. Metrics & Dashboards

What key metrics should we measure? For example:
- Are we tracking system latency when interacting with external systems?
- How quickly are configurations applied?
- Are agent failures detected and logged efficiently?

## 5. Drawbacks

Are there reasons not to proceed with this proposal? 
- What are the risks involved?

* This requires the rewrite of viper in statically parsing the configuration files

- Could decentralized management introduce more complexity than it solves?


## 6. Alternatives

What are other ways to achieve the same outcome?
- Could the problem be solved through better coordination of centralized configuration?
- Could existing infrastructure tools handle custom configurations better?

## 7. Potential Impact and Dependencies

This change will affect various teams and systems:
- What other systems or teams could be impacted by this proposal?
- Could this be exploited by malicious actors, and if so, how can we mitigate that risk?

## 8. Unresolved Questions

What parts of this proposal still need more thought or refinement? What questions remain unanswered?

## 9. Conclusion

Flyte's agent framework has enabled teams to efficiently develop, test, and release custom agents. By supporting decentralized custom agent configuration management, we streamline deployment and operations while unlocking the full potential of Flyte’s decoupled architecture.

## 10. RFC Process Guide (Remove this section when done)

By writing an RFC, you're providing insight to your team on the direction you're taking. There may not always be a right or better decision, but we will likely learn from the process. By authoring this RFC, you're deciding where you want us to go and are looking for feedback on that direction.

This document is:
- A thinking exercise, a prototype in words.
- A historical record, which may lose value over time.
- A way to broadcast information.
- A mechanism to build trust.
- A tool to empower.
- A communication channel.

This document is *not*:
- A request for permission.
- The most up-to-date representation of any process or system.

**Checklist:**
- [ ] Copy template
- [ ] Draft RFC (think of it as a wireframe)
- [ ] Share as WIP with trusted colleagues for feedback
- [ ] Send pull request when comfortable
- [ ] Label accordingly
- [ ] Assign reviewers
- [ ] Merge PR

**Recommendations:**
- Tag RFC title with [WIP] if you're still ironing out details.
- Tag RFC title with [Newbie] if you're experimenting or unsure of the proposal.
- Tag RFC title with [RR] if you'd like to schedule a review request.
- Tag people who may have expertise on uncertain areas for feedback.
- If you're uncertain, ask for help in relevant Slack channels.
