# Configuring Logging Links in the UI

To debug your workflows in production, you want to access logs from your tasks as they run. These logs are different from the core Flyte platform logs and are specific to each task execution.

Flyte provides a flexible way to configure log links in the UI, supporting multiple backends such as AWS CloudWatch, GCP Stackdriver, Datadog, and others. These links can be fully customized using a templating engine based on your organization’s log storage and observability tooling.

---

## Supported Template Parameters

You can use the following variables inside your logging URL templates. These values are substituted dynamically at runtime.

| Parameter                      | Description                                                                 |
|-------------------------------|-----------------------------------------------------------------------------|
| `{{ .podName }}`              | Pod name as shown in Kubernetes dashboard                                  |
| `{{ .podUID }}`               | UID of the pod                                                              |
| `{{ .namespace }}`            | Kubernetes namespace where the pod runs                                    |
| `{{ .containerName }}`        | Name of the container                                                      |
| `{{ .containerId }}`          | ID of the container                                                        |
| `{{ .logName }}`              | Deployment-specific log identifier                                         |
| `{{ .hostname }}`             | Pod’s internal hostname                                                    |
| `{{ .nodeName }}`             | Node on which the pod is running                                           |
| `{{ .podRFC3339StartTime }}`  | Pod start time in RFC3339 format                                           |
| `{{ .podRFC3339FinishTime }}` | Pod end time in RFC3339 (approximate)                                      |
| `{{ .podUnixStartTime }}`     | Pod start time in Unix seconds                                             |
| `{{ .podUnixFinishTime }}`    | Pod finish time in Unix seconds (approximate)                              |

---

## Example: Configuring CloudWatch Logs (with EKS Add-on)

For Amazon EKS users, we recommend using the CloudWatch Observability EKS add-on. Below is an example configuration to generate CloudWatch log links per task:

```yaml
task_logs:
  plugins:
    logs:
      templates:
        - displayName: CloudWatch (EKS Observability)
          templateUris:
            - "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logsV2:log-groups/log-group=$252Faws$252Eeks$252Ecluster$252Fflyte-platform/log-events/{{ '{{' }} .podName {{ '}}' }}"
          messageFormat: 0
