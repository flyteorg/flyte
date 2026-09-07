# Settings API: Customer Flow

This walkthrough traces a realistic customer flow end-to-end. To keep the examples
readable, only three settings are used:

| Setting                        | Type             | Description |
|--------------------------------|------------------|---|
| `run.defaultQueue`             | StringSetting    | Queue that runs are submitted to |
| `taskResource.min.cpu`         | QuantitySetting  | Minimum CPU requested for task pods |
| `environmentVariables`         | StringMapSetting | Environment variables injected into task pods — **additive** across scopes |

**Scope hierarchy:** `acme` (org) → `acme/production` (domain) → `acme/production/analytics` (project)

Every request and response below is executed against a running server by
`runs/test/api/settings_flow_test.go`, so this document is verified by CI rather
than maintained by hand.

### API rules in effect

- Every setting is a typed wrapper message (`StringSetting`, `Int64Setting`, `BoolSetting`, `QuantitySetting`, `StringMapSetting`) carrying an explicit `state`:
  - `SETTING_STATE_INHERIT` (the default) — take the value from the parent scope
  - `SETTING_STATE_UNSET` — explicitly clear; stops inheritance
  - `SETTING_STATE_VALUE` — use this message's value field
- Omitting a setting and sending it as `{}` both mean INHERIT, and neither is stored.
- **A value sent without `state: SETTING_STATE_VALUE` is treated as INHERIT and ignored.**
- `GetSettings` resolves inheritance top-down and returns one merged record wrapped in `settingsRecord`, with each resolved setting annotated by the `scopeLevel` it came from. No descriptions. Merged records carry no `version`, since a merged view corresponds to no single stored row; clients that need a version for an update use `GetSettingsForEdit`.
- `SCOPE_LEVEL_ORG` is the enum zero value and is omitted from JSON, so an absent `scopeLevel` means the value resolved from the org.
- `GetSettingsForEdit` returns `requestedKey` plus a `levels` array, one entry per scope level covered by the request key, ordered broadest to most specific. Each entry is `{ key, settings, version }` where `key` is a partial key identifying that level. No descriptions and no `scopeLevel`. A level with no stored record has empty settings and version 0; zero versions are omitted from JSON, so a missing `version` field means "no record yet, use `CreateSettings` for this level".
- Map settings (`environmentVariables`) are **additive** on `GetSettings`: parent entries first, child entries merged on top (child wins on key conflict), and a level in state `UNSET` clears everything accumulated above it. `scopeLevel` on a merged map is the most specific level that contributed entries.

---

## Starting State

Org `acme` has baseline settings from provisioning. The project `analytics` in
domain `production` was previously configured with a min CPU override.
No domain settings exist yet.

**Database**

| id | key                            | data (JSONB)                                                                                                                                                                                                                                                | version |
|----|--------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|
| 1  | `v1:acme::`                    | `{"run":{"defaultQueue":{"state":"SETTING_STATE_VALUE","stringValue":"default"}},"taskResource":{"min":{"cpu":{"state":"SETTING_STATE_VALUE","quantityValue":"500m"}}},"environmentVariables":{"state":"SETTING_STATE_VALUE","mapValue":{"entries":{"LOG_LEVEL":"info","REGION":"us-east-1"}}}}` | 1       |
| 2  | `v1:acme:production:analytics` | `{"taskResource":{"min":{"cpu":{"state":"SETTING_STATE_VALUE","quantityValue":"1000m"}}}}`                                                                                                                                                                  | 1       |

---

## Step 1 — GetSettings at org scope

Retrieve the effective settings for the org. At org scope there is no parent to
inherit from, so every setting either has a value or is absent from the response.

**Request — `GetSettings`**
```json
{
  "key": { "org": "acme" }
}
```

**Response**
```json
{
  "settingsRecord": {
    "key": { "org": "acme" },
    "settings": {
      "run": {
        "defaultQueue": { "state": "SETTING_STATE_VALUE", "stringValue": "default" }
      },
      "taskResource": {
        "min": {
          "cpu": { "state": "SETTING_STATE_VALUE", "quantityValue": "500m" }
        }
      },
      "environmentVariables": {
        "state": "SETTING_STATE_VALUE",
        "mapValue": { "entries": { "LOG_LEVEL": "info", "REGION": "us-east-1" } }
      }
    }
  }
}
```

All three values resolved from the org, and `SCOPE_LEVEL_ORG` is the enum zero
value, so no `scopeLevel` appears in the JSON. There is no `version` either:
merged records never carry one.

**Database:** unchanged.

---

## Step 2 — Update project settings

### 2a — GetSettingsForEdit at project scope

Before editing, fetch the current stored state at all levels. This returns each
level's stored settings as-is: inheritance is not applied.

**Request — `GetSettingsForEdit`**
```json
{
  "key": { "org": "acme", "domain": "production", "project": "analytics" }
}
```

**Response**
```json
{
  "requestedKey": { "org": "acme", "domain": "production", "project": "analytics" },
  "levels": [
    {
      "key": { "org": "acme" },
      "settings": {
        "run": {
          "defaultQueue": { "state": "SETTING_STATE_VALUE", "stringValue": "default" }
        },
        "taskResource": {
          "min": {
            "cpu": { "state": "SETTING_STATE_VALUE", "quantityValue": "500m" }
          }
        },
        "environmentVariables": {
          "state": "SETTING_STATE_VALUE",
          "mapValue": { "entries": { "LOG_LEVEL": "info", "REGION": "us-east-1" } }
        }
      },
      "version": "1"
    },
    {
      "key": { "org": "acme", "domain": "production" },
      "settings": {}
    },
    {
      "key": { "org": "acme", "domain": "production", "project": "analytics" },
      "settings": {
        "taskResource": {
          "min": {
            "cpu": { "state": "SETTING_STATE_VALUE", "quantityValue": "1000m" }
          }
        }
      },
      "version": "1"
    }
  ]
}
```

The `levels` array runs from broadest to most specific and always contains one
record per level covered by the request key. The domain has no stored record
yet, so its entry has empty settings and no `version` field (version 0 is
omitted from JSON). That missing version is the client's signal to use
`CreateSettings` for the domain rather than `UpdateSettings`. The project shows
only what is stored there: the `1000m` CPU override. Use `version: 1` from the
project entry for the update.

**Database:** unchanged.

---

### 2b — UpdateSettings at project scope

Increase the CPU minimum for this compute-heavy project and add a
project-specific env var. `defaultQueue` is sent as an empty object, one of the
two equivalent spellings of INHERIT. Supply `version: 1` from the project entry
above.

**Request — `UpdateSettings`**
```json
{
  "key": { "org": "acme", "domain": "production", "project": "analytics" },
  "settings": {
    "run": {
      "defaultQueue": {}
    },
    "taskResource": {
      "min": {
        "cpu": { "state": "SETTING_STATE_VALUE", "quantityValue": "2000m" }
      }
    },
    "environmentVariables": {
      "state": "SETTING_STATE_VALUE",
      "mapValue": { "entries": { "LOG_LEVEL": "debug" } }
    }
  },
  "version": "1"
}
```

**Response**
```json
{
  "settingsRecord": {
    "key": { "org": "acme", "domain": "production", "project": "analytics" },
    "settings": {
      "taskResource": {
        "min": {
          "cpu": { "state": "SETTING_STATE_VALUE", "quantityValue": "2000m" }
        }
      },
      "environmentVariables": {
        "state": "SETTING_STATE_VALUE",
        "mapValue": { "entries": { "LOG_LEVEL": "debug" } }
      }
    },
    "version": "2"
  }
}
```

The response echoes exactly what was stored. The empty `defaultQueue` object
was pruned before storing, so it appears in neither the stored row nor the
echo: sending a setting as `{}` and omitting it entirely produce identical
results. `version` is now `2`.

**Database**

| id | key                            | data (JSONB)                                                                                                                                                                                                                | version |
|----|--------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|
| 1  | `v1:acme::`                    | `{"run":{"defaultQueue":{"state":"SETTING_STATE_VALUE","stringValue":"default"}},...}` (unchanged)                                                                                                                            | 1       |
| 2  | `v1:acme:production:analytics` | `{"taskResource":{"min":{"cpu":{"state":"SETTING_STATE_VALUE","quantityValue":"2000m"}}},"environmentVariables":{"state":"SETTING_STATE_VALUE","mapValue":{"entries":{"LOG_LEVEL":"debug"}}}}` | 2       |

`run.defaultQueue` is absent from the project row: INHERIT is never written to
the database, whether the client omitted the field or sent it empty.

---

## Step 3 — GetSettings at project scope

Retrieve effective settings for the project. No domain record exists yet, so
values resolve through two levels: org → project.

**Request — `GetSettings`**
```json
{
  "key": { "org": "acme", "domain": "production", "project": "analytics" }
}
```

**Response**
```json
{
  "settingsRecord": {
    "key": { "org": "acme", "domain": "production", "project": "analytics" },
    "settings": {
      "run": {
        "defaultQueue": { "state": "SETTING_STATE_VALUE", "stringValue": "default" }
      },
      "taskResource": {
        "min": {
          "cpu": {
            "state": "SETTING_STATE_VALUE",
            "quantityValue": "2000m",
            "scopeLevel": "SCOPE_LEVEL_PROJECT"
          }
        }
      },
      "environmentVariables": {
        "state": "SETTING_STATE_VALUE",
        "mapValue": { "entries": { "LOG_LEVEL": "debug", "REGION": "us-east-1" } },
        "scopeLevel": "SCOPE_LEVEL_PROJECT"
      }
    }
  }
}
```

- `defaultQueue`: no override in the chain, resolved from the org. Its `scopeLevel` is `SCOPE_LEVEL_ORG`, the zero value, so the field is absent.
- `taskResource.min.cpu`: the project override from step 2b, annotated `SCOPE_LEVEL_PROJECT`.
- `environmentVariables`: additive merge. The org contributes `LOG_LEVEL=info` and `REGION=us-east-1`; the project contributes `LOG_LEVEL=debug`, which wins the key conflict. `scopeLevel` is `SCOPE_LEVEL_PROJECT`, the most specific level that contributed entries.

**Database:** unchanged.

---

## Step 4 — CreateSettings at domain scope

The `production` domain has no settings record yet (its level came back with no
version in step 2a). Create one to route all production runs to a dedicated
queue.

**Request — `CreateSettings`**
```json
{
  "key": { "org": "acme", "domain": "production" },
  "settings": {
    "run": {
      "defaultQueue": { "state": "SETTING_STATE_VALUE", "stringValue": "fast-queue" }
    },
    "taskResource": {
      "min": {
        "cpu": {}
      }
    },
    "environmentVariables": {}
  }
}
```

**Response**
```json
{
  "settingsRecord": {
    "key": { "org": "acme", "domain": "production" },
    "settings": {
      "run": {
        "defaultQueue": { "state": "SETTING_STATE_VALUE", "stringValue": "fast-queue" }
      }
    },
    "version": "1"
  }
}
```

**Database**

| id | key                            | data (JSONB)                                                                                                                                                                                       | version |
|----|--------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|
| 1  | `v1:acme::`                    | `{"run":{"defaultQueue":{"state":"SETTING_STATE_VALUE","stringValue":"default"}},...}` (unchanged)                                                                                                     | 1       |
| 3  | `v1:acme:production:`          | `{"run":{"defaultQueue":{"state":"SETTING_STATE_VALUE","stringValue":"fast-queue"}}}`                                                                                                                  | 1       |
| 2  | `v1:acme:production:analytics` | `{"taskResource":{"min":{"cpu":{"state":"SETTING_STATE_VALUE","quantityValue":"2000m"}}},...}` (unchanged)                                                                                             | 2       |

Only `defaultQueue` reaches the domain row: the empty CPU and env-var objects
were pruned before storing.

---

## Step 5 — GetSettings at project scope (after domain is set)

The project settings have not changed. But now a domain record exists, so the
resolution chain is org → domain → project.

**Request — `GetSettings`**
```json
{
  "key": { "org": "acme", "domain": "production", "project": "analytics" }
}
```

**Response**
```json
{
  "settingsRecord": {
    "key": { "org": "acme", "domain": "production", "project": "analytics" },
    "settings": {
      "run": {
        "defaultQueue": {
          "state": "SETTING_STATE_VALUE",
          "stringValue": "fast-queue",
          "scopeLevel": "SCOPE_LEVEL_DOMAIN"
        }
      },
      "taskResource": {
        "min": {
          "cpu": {
            "state": "SETTING_STATE_VALUE",
            "quantityValue": "2000m",
            "scopeLevel": "SCOPE_LEVEL_PROJECT"
          }
        }
      },
      "environmentVariables": {
        "state": "SETTING_STATE_VALUE",
        "mapValue": { "entries": { "LOG_LEVEL": "debug", "REGION": "us-east-1" } },
        "scopeLevel": "SCOPE_LEVEL_PROJECT"
      }
    }
  }
}
```

- `defaultQueue` now resolves to `fast-queue` from the domain, overriding the org's `default`, and for the first time a `scopeLevel` appears on it: `SCOPE_LEVEL_DOMAIN` is nonzero, so it serializes. Nothing in the project record changed; the domain insert was enough.
- `taskResource.min.cpu`: still `2000m` from the project. The domain did not override it.
- `environmentVariables`: still the additive merge of org and project. The domain contributed no entries, so the result and its `scopeLevel` are unchanged.

**Database:** unchanged.

---

## Summary

| Step | Scope touched | What changed |
|------|--------------|--------------|
| 1 | — | Read org settings; all annotations implicit (org is the zero scope level) |
| 2a | project (read) | GetSettingsForEdit: three levels returned, org and project with values and versions, the domain as a gap record with empty settings and no version |
| 2b | project (write) | Updated min CPU 1000m→2000m, added `environmentVariables={LOG_LEVEL:debug}`; the empty `defaultQueue` was pruned; version 1→2 |
| 3 | — | Read project: org+project merge, no domain yet; `LOG_LEVEL` overridden by the project |
| 4 | domain (create) | Set `defaultQueue=fast-queue` at the domain; empty settings pruned; version 1 |
| 5 | — | Read project: the domain now intercepts `defaultQueue`; env vars and CPU unchanged |

The key insight: changes at one scope are immediately visible to all child scopes
on the next `GetSettings` call. No project-level record needs to be touched when
a domain setting is added.
