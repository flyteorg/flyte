# Data Proxy

Data proxy is the service that communicates with the data storage to handle operations like:
- Generate signed URLs for uploading data
- Generate download URLs for retrieving data
- Manage data access and permissions

Data proxy lives in the Flyte control plane, while the data storage lives in the data plane. Data proxy is the bridge
to request URLs for uploading/downloading data to any data storage you have.

## Quick Start

For local development, see [DEVELOPMENT.md](DEVELOPMENT.md) for detailed steps on setting up MinIO and running the service locally.

For production deployment:

1. Configure your storage backend (S3, MinIO, GCS, etc.)
2. Create a configuration file based on `config/config.example.yaml`
3. Run the service:

```bash
go run cmd/main.go --config config/config.yaml
```

The service will start on port 8088 by default. You can override with `--port` and `--host` flags.

## Configuration

The data proxy service uses a YAML configuration file. See [config.example.yaml](config/config.example.yaml) for a complete example.

### Data Proxy Settings

| Field | Description | Default |
|-------|-------------|---------|
| `dataproxy.upload.maxSize` | Maximum allowed upload size | `100Mi` |
| `dataproxy.upload.maxExpiresIn` | Maximum expiration time for signed upload URLs | `1h` |
| `dataproxy.upload.defaultFileNameLength` | Default length for auto-generated filenames | `20` |
| `dataproxy.upload.storagePrefix` | Prefix for all uploaded files in storage | `uploads` |
| `dataproxy.download.maxExpiresIn` | Maximum expiration time for download URLs | `1h` |

### Storage Backend Settings

| Field | Description | Options |
|-------|-------------|---------|
| `storage.type` | Storage backend type | `s3`, `minio`, `local`, `mem`, `stow` |
| `storage.container` | Initial bucket/container name | - |
| `storage.enable-multicontainer` | Allow access to multiple buckets | `true`/`false` |
| `storage.connection.endpoint` | Storage endpoint URL | - |
| `storage.connection.auth-type` | Authentication method | `iam`, `accesskey` |
| `storage.connection.access-key` | Access key (when using `accesskey` auth) | - |
| `storage.connection.secret-key` | Secret key (when using `accesskey` auth) | - |
| `storage.connection.region` | Storage region | `us-east-1` |
| `storage.connection.disable-ssl` | Disable SSL (for local dev only) | `true`/`false` |

### MinIO-Specific Configuration

For MinIO, use the following configuration:

```yaml
storage:
  type: minio
  container: "flyte-data"
  connection:
    endpoint: "http://minio.flyte-dataproxy.svc.cluster.local:9000"
    auth-type: accesskey
    access-key: "minioadmin"
    secret-key: "minioadmin"
    region: "us-east-1"
    disable-ssl: true
```

### S3-Compatible Storage

For AWS S3 or S3-compatible storage:

```yaml
storage:
  type: s3
  container: "my-flyte-bucket"
  connection:
    auth-type: iam  # Use IAM roles
    region: "us-west-2"
    disable-ssl: false
```

## Architecture

```
┌─────────────────┐
│   Flyte Client  │
└────────┬────────┘
         │ Request signed URL
         │
         ▼
┌─────────────────┐      ┌──────────────────┐
│   Data Proxy    │─────▶│  MinIO/S3 API    │
│   (Control      │      │  (Data Plane)    │
│    Plane)       │◀─────│                  │
└─────────────────┘      └──────────────────┘
         │                        ▲
         │ Return signed URL      │
         │                        │
         ▼                        │
┌─────────────────┐               │
│   Flyte Client  │───────────────┘
└─────────────────┘    Upload/Download
                       using signed URL
```

## Development

For local development and testing, see [DEVELOPMENT.md](DEVELOPMENT.md).

## Troubleshooting

### MinIO Connection Issues

1. **Check MinIO is running:**
   ```bash
   kubectl get pods -n flyte-dataproxy
   ```

2. **Check MinIO logs:**
   ```bash
   kubectl logs -n flyte-dataproxy -l app=minio
   ```

3. **Verify network connectivity:**
   ```bash
   kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
     curl http://minio.flyte-dataproxy.svc.cluster.local:9000/minio/health/live
   ```

### Data Proxy Service Issues

1. **Check configuration:**
   - Verify the storage endpoint is correct
   - Ensure credentials are valid
   - Check that the bucket/container exists

2. **Enable debug logging:**
   ```yaml
   logger:
     level: 5  # Debug level
     show-source: true
   ```

3. **Common errors:**
   - `connection refused`: MinIO service not reachable
   - `access denied`: Invalid credentials
   - `bucket not found`: Container doesn't exist (enable auto-creation or create manually)

## Production Deployment

For production use:

1. **Use a production-grade storage backend** (AWS S3, GCS, Azure Blob Storage, or MinIO in distributed mode)
2. **Enable TLS/SSL** for all connections
3. **Use IAM roles** instead of static credentials when possible
4. **Configure resource limits** and proper scaling
5. **Set up monitoring and alerting** for storage operations
6. **Review security settings** for signed URL expiration times

See the Flyte documentation for production deployment guides.
