# Pod disruption budgets

Each `components.<name>.podDisruptionBudget` and
`console.podDisruptionBudget` can create a `policy/v1` PodDisruptionBudget.
Budgets are disabled by default and are only created for enabled components.

For example, protect a three-replica runs service during voluntary evictions:

```yaml
components:
  runs:
    replicaCount: 3
    podDisruptionBudget:
      enabled: true
      maxUnavailable: 1
```

To use a minimum instead, explicitly clear the default maximum:

```yaml
console:
  replicaCount: 3
  podDisruptionBudget:
    enabled: true
    minAvailable: "66%"
    maxUnavailable: null
```

Set exactly one of `minAvailable` or `maxUnavailable` to an integer or a
percentage string. Zero is supported. The default `maxUnavailable: 1` permits
eviction of a single-replica component. Setting `minAvailable: 1` or
`maxUnavailable: 0` on a single replica can block node drains. In particular,
the executor currently runs with one replica; a PDB does not add failover.

PDBs constrain voluntary evictions, such as node drains. They do not prevent
node failures or control Deployment rolling updates. Configure replica counts,
topology spreading, and Deployment strategy separately for availability.
