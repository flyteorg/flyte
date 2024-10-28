Flyte’s interaction with blob storage occurs through several key components and API calls, especially when executing workflows remotely. Here's an overview to help you understand how `pyflyte run --remote` interacts with blob storage and where keys or credentials are involved.

1. Blob Storage Usage in `pyflyte run --remote`:
   - When you execute a workflow remotely with `pyflyte run --remote`, the client communicates with FlyteAdmin to get a signed URL for uploading the workflow's inputs to blob storage. This URL is then used by the client to upload data directly to storage.

2. CreateUploadLocation API Call:
   - FlyteAdmin handles the `CreateUploadLocation` call, generating pre-signed URLs that allow users or services to upload inputs to storage without exposing sensitive keys directly. This ensures controlled and secure access to storage resources.

3. Input/Output Handling in Native and Raw Container Tasks:
   - For native Flyte tasks, inputs and outputs are uploaded/downloaded through FlyteAdmin using appropriate credentials configured on the platform.
   - In container-based tasks managed by Flyte’s Copilot component, similar blob storage access occurs, but Copilot handles the interactions based on container-level configurations. This allows tasks to offload large data efficiently.

4. Default Configuration and Credentials Management:
   - Flyte commonly uses IRSA (IAM Roles for Service Accounts) in AWS environments to handle authentication to blob storage without relying on long-term credentials. However, other configurations can also be used (e.g., environment variables or secret management systems).

5. Customization and Flexibility:
   - Since Flyte supports a variety of environments, customizing storage access patterns may involve configuring the FlyteAdmin service and plugins. This ensures compatibility across clouds and storage systems beyond the defaults.

These components come together to ensure data flows securely and efficiently between Flyte services and storage endpoints. For a deeper dive, including the internal logic of FlyteAdmin and details about `pyflyte`, you can explore Flyte’s official documentation on [FlyteAdmin API interaction](https://docs.flyte.org/en/latest/concepts/data_management.html).