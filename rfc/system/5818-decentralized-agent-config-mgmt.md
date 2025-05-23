# Support Decentralized Custom Agent Configuration Management

**Authors:**
- @shuyingliang

## 1. Executive Summary

While the Flyte agent framework provides an efficient and productive way for users to develop Flyte plugins, all agents must currently be configured within the FlytePropeller configuration, with the config map mounted under `/etc/flyte/config`. 

In an enterprise environment, different teams are responsible for developing custom agents and managing Flyte components like Propeller, Scheduler, Admin, and Console. These teams often have their own CI/CD pipelines to deploy custom agents. The centralized agent configuration process can lead to inefficiencies and outage (real case[^1]), as the deployment of custom agents may fall out of sync with the centralized configuration or cause unnecessary operational cost.

[^1]: The outage we have encountered at LinkedIn environment: The team developing custom agent roll out the agent cluster by cluster, but the propeller and custom agent configuration was "set up" somehow because they are out of sync. As a result, the propeller pods went in Crashloop causing a bunch of jobs un `UnKnown` state

This RFC proposes support for decentralized custom agent configuration management, with benefits including:
- **Separation of concerns**: Teams can *fully* independently manage their own custom agent configurations.
- **Decentralized management**: Custom agents can be deployed and managed independently from the core Flyte infrastructure, yet operate as a cohesive unit.
- **Flexibility**: FlytePropeller can dynamically adapt to varying configurations without requiring centralized maintenance of all settings.

## 2. Motivation

In enterprise environments, the team managing the core Flyte infrastructure is usually different from the teams developing custom agents/plugins. These teams operate with separate development cycles and CI/CD pipelines. For instance, a team writing its own custom agent would create its own service account, Docker image, and deploy it in a separate namespace with its own monitoring dashboard.

However, the final step of deployment—updating the Propeller configuration—remains centralized. This leads to issues like:
1. **Out-of-sync deployment**: When a custom agent endpoint needs to be updated, it requires coordination between teams for the deployment of both the agent and the Propeller configuration update, leading to potential mismatches and outage.
2. **Coordination overhead**: Misaligned rollouts could result in the Propeller being unable to reach the agent, potentially causing blast radius of impact due to propeller is down.

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

There can be different metrics to track the custom agent configuration validation such as total request, validated cases and its latency. These [metrics](https://github.com/kubernetes-sigs/controller-runtime/blob/main/pkg/webhook/internal/metrics/metrics.go) are natively supported in controller framework such as controller-runtime

Any failures in aggregation and pre-checks of the agent reachability should be properly logged. K8s events can be emitted as well. This enables the enterprise environment to aggregate the logs/events for troubleshooting. 


## 5. Drawbacks

Are there reasons not to proceed with this proposal? 
- What are the risks involved?

* This requires the rewrite of viper in statically parsing the configuration files

- Could decentralized management introduce more complexity than it solves?

Instead of statically parse from the config files, the change is more or less to dynamically interact the K8s API to constructs the `Config` object



## 6. Alternatives

An alternative approach could be to specify the ConfigMap as a volume and mount the custom agents' ConfigMap in the Propeller pod. This way, the existing Viper parsing might still work, along with additional aggregation and merging capabilities.

```yaml
spec:
  containers:
  - name: propeller-container
    image: image
    volumeMounts:
    - name: config-volume
      mountPath: /etc/config/flyte/custom-agents/custom-agent-configmap.yaml
  volumes:
  - name: config-volume
    configMap:
      name: custom-agent-configmap
```

However, this still requires modifications to the centralized Propeller `Deployment` YAML file, which is similar to the current approach of directly updating the centralized Propeller ConfigMap. Thus, the same challenges will persist.


## 7. Potential Impact and Dependencies

This change requires authentication and authorization to the Kubernetes API server. However, this should not be an issue because:

* The Propeller or Admin's service account can authenticate with the Kubernetes API server.
* The Propeller or Admin operates in the control plane, and enterprise security already authorizes these roles to perform CRUD operations on ConfigMap objects. The deployment Helm template should include a role with permissions to manage ConfigMaps.

## 8. Unresolved Questions

None that I am aware of, based on the discussions in the `#flyte-agents` channel so far.

## 9. Conclusion

Flyte's agent framework has empowered teams to efficiently develop, test, and release custom agents. By supporting decentralized custom agent configuration management, we can streamline deployment and operations reliably while maximizing the benefits of Flyte’s decoupled architecture.


**Recommendations:**
- Tag RFC title with [WIP] if you're still ironing out details.
- Tag RFC title with [Newbie] if you're experimenting or unsure of the proposal.
- Tag RFC title with [RR] if you'd like to schedule a review request.
- Tag people who may have expertise on uncertain areas for feedback.
- If you're uncertain, ask for help in relevant Slack channels.
