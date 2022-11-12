# flyte-binary

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.16.0](https://img.shields.io/badge/AppVersion-1.16.0-informational?style=flat-square)

Chart for basic single Flyte executable deployment

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| database.dbname | string | `"flyteadmin"` |  |
| database.host | string | `"127.0.0.1"` |  |
| database.password | string | `""` |  |
| database.port | int | `5432` |  |
| database.username | string | `"postgres"` |  |
| image.fqn | string | `"ghcr.io/flyteorg/flyte-sandbox:flytebinary_1007"` |  |
| logger.level | int | `1` |  |
| networking.host | bool | `true` |  |
| paths.externalFlyteConfig | string | `""` |  |
| proxy.enabled | bool | `false` |  |
| storage.metadataContainer | string | `"my-s3-bucket"` |  |
| storage.region | string | `"my-region"` |  |
| storage.type | string | `"minio"` |  |
| storage.userDataContainer | string | `"my-s3-bucket"` |  |

