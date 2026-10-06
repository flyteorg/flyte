# Redis client configuration

`Config.NewClient` returns a go-redis `UniversalClient` for standalone, cluster,
or Sentinel connections. Shared credentials, TLS, timeouts, retries, and pool
settings apply to all modes. `GetUniversalOptions` exposes the resolved settings
for all modes. Both methods accept optional `Option` arguments: use
`cfg.NewClient(ctx)` or `cfg.NewClient(ctx, redis.WithSecretManager(secretManager))`.

Use these fields under `cache.redis.options` or `storage.redis`:

```yaml
# Standalone (existing configuration remains supported)
addr: localhost:6379

# Cluster with seed nodes
addrs:
  - redis-1:6379
  - redis-2:6379

# Cluster with a single configuration endpoint
addr: redis-cluster:6379
isClusterMode: true

# Sentinel: addrs are Sentinel endpoints, not Redis server endpoints
addrs:
  - sentinel-1:26379
  - sentinel-2:26379
masterName: mymaster
sentinelUsername: sentinel-user
sentinelPassword: sentinel-password
username: redis-user
password: redis-password
```

Each example is a separate configuration. `addrs` takes precedence over `addr`.
Multiple addresses select cluster mode unless `masterName` selects Sentinel.
Cluster supports database 0 only. Cluster and Sentinel connections use TCP.
`maxRedirects`, `readOnly`, `routeByLatency`, and `routeRandomly` configure routing;
`readOnly` selects replicas in Sentinel mode.

`passwordSecretName` and `sentinelPasswordSecretName` resolve passwords through
an injected `SecretManager`. Cache supplies that manager; storage currently
requires directly configured passwords. `useTLS` enables TLS in all modes.

Storage listings scan all cluster masters and deduplicate the results. Reference
URLs use the first configured endpoint as their host.
