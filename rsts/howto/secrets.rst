.. _howto-secrets:

##################################
How do I inject secrets into tasks
##################################


*************************
What is secrets injection
*************************

Flyte supports running a wide variety of tasks; from containers to sql queries and service calls. In order for flyte-run
containers to request and access secrets, flyte now natively supports a Secret construct.

For a simple task that launches a Pod, the flow will look something like this:

.. image:: https://mermaid.ink/img/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4gICAgUHJvcGVsbGVyLT4-K1BsdWdpbnM6IENyZWF0ZSBLOHMgUmVzb3VyY2VcbiAgICBQbHVnaW5zLT4-LVByb3BlbGxlcjogUmVzb3VyY2UgT2JqZWN0XG4gICAgUHJvcGVsbGVyLT4-K1Byb3BlbGxlcjogU2V0IExhYmVscyAmIEFubm90YXRpb25zXG4gICAgUHJvcGVsbGVyLT4-K0FwaVNlcnZlcjogQ3JlYXRlIE9iamVjdCAoZS5nLiBQb2QpXG4gICAgQXBpU2VydmVyLT4-K1BvZCBXZWJob29rOiAvbXV0YXRlXG4gICAgUG9kIFdlYmhvb2stPj4rUG9kIFdlYmhvb2s6IExvb2t1cCBnbG9iYWxzXG4gICAgUG9kIFdlYmhvb2stPj4rUG9kIFdlYmhvb2s6IEluamVjdCBTZWNyZXQgQW5ub3RhdGlvbnMgKGUuZy4gSzhzLCBWYXVsdC4uLiBldGMuKVxuICAgIFBvZCBXZWJob29rLT4-LUFwaVNlcnZlcjogTXV0YXRlZCBQb2RcbiAgICBcbiAgICAgICAgICAgICIsIm1lcm1haWQiOnt9LCJ1cGRhdGVFZGl0b3IiOmZhbHNlfQ
   :target: https://mermaid.ink/img/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4gICAgUHJvcGVsbGVyLT4-K1BsdWdpbnM6IENyZWF0ZSBLOHMgUmVzb3VyY2VcbiAgICBQbHVnaW5zLT4-LVByb3BlbGxlcjogUmVzb3VyY2UgT2JqZWN0XG4gICAgUHJvcGVsbGVyLT4-K1Byb3BlbGxlcjogU2V0IExhYmVscyAmIEFubm90YXRpb25zXG4gICAgUHJvcGVsbGVyLT4-K0FwaVNlcnZlcjogQ3JlYXRlIE9iamVjdCAoZS5nLiBQb2QpXG4gICAgQXBpU2VydmVyLT4-K1BvZCBXZWJob29rOiAvbXV0YXRlXG4gICAgUG9kIFdlYmhvb2stPj4rUG9kIFdlYmhvb2s6IExvb2t1cCBnbG9iYWxzXG4gICAgUG9kIFdlYmhvb2stPj4rUG9kIFdlYmhvb2s6IEluamVjdCBTZWNyZXQgQW5ub3RhdGlvbnMgKGUuZy4gSzhzLCBWYXVsdC4uLiBldGMuKVxuICAgIFBvZCBXZWJob29rLT4-LUFwaVNlcnZlcjogTXV0YXRlZCBQb2RcbiAgICBcbiAgICAgICAgICAgICIsIm1lcm1haWQiOnt9LCJ1cGRhdGVFZGl0b3IiOmZhbHNlfQ

Where:

1. Flyte invokes a plugin to create the K8s object. This can be a Pod or a more complex CRD (e.g. Spark, PyTorch... etc.)

   .. tip:: It's the plugin's responsibility to ensure labels and annotations are passed through to any Pod that will be spawneddue to the creation of the CRD.

3. Flyte will apply labels and annotations that has references to all secrets the task is requesting access to.
4. Flyte will send a POST request to ApiServer to create the object.
5. Before persisting the Pod, ApiServer will invoke all registered Pod Webhooks. Flyte's Pod Webhook will be called.
6. Flyte Pod Webhook will then lookup globally mounted secrets for each of the requested secrets, if found it'll mount
   them directly in the Pod.
7. If not found, it'll inject the appropriate annotations to load the secrets for K8s (or Vault or Confidant or any other
   secret management system plugin configured) into the Pod.

******************************
How to enable secret injection
******************************

This feature is available in Flytekit v0.17.0+. Refer here for an example annotated task: <TODO FlyteSnacks Cross Reference>

The webhook is included in all overlays in this repo. The deployment file creates (mainly) two things; a Job and a Deployment.

1) flyte-pod-webhook-secrets Job: This job runs ``flytepropeller webhook init-certs`` command that issues self-signed
   CA Certificate as well as a derived TLS certificate and its private key. It stores them into a new secret ``flyte-pod-webhook-secret``.
2) flyte-pod-webhook Deployment: This deployment creates the Webhook pod which creates a MutatingWebhookConfiguration
   on startup. This serves as the registration contract with ApiServer to know about the Webhook before it starts serving
   traffic.

*******************
Scaling the webhook
*******************

The Webhook does not make any external API Requests in response to Pod mutation requests. It should be able to handle traffic
quickly (a benchmark is needed). In case it needs to be horizontally scaled, adding additional replicas for the Pod in the
deployment should be all that's required. A single MutatingWebhookConfiguration object will be used, the same TLS certificate
will be shared across the pods and the Service created will automatically load balance traffic across the available pods.