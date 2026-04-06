# Settings API: Customer Flow

This walkthrough traces a realistic customer flow end-to-end. To keep the examples
readable, only three settings are used:

| Setting | Type | Description |
|---|---|---|
| `run.defaultQueue` | string | Queue that runs are submitted to |
| `taskResource.defaults.minCpu` | string | Minimum CPU requested for task pods |
| `environmentVariables` | map(string,string) | Environment variables injected into task pods — **additive** across scopes |

**Scope hierarchy:** `acme` (org) → `acme/production` (domain) → `acme/production/analytics` (project)

### API rules in effect

- All settings use a single `SettingValue` message. The oneof field set determines the type and state:
  - `state: INHERIT` (or empty) — delegate to parent scope
  - `state: UNSET` — explicitly clear; stops inheritance
  - `stringValue: "..."` — string value (implies VALUE state)
  - `intValue: N` — integer value (implies VALUE state)
  - `listValue: { values: [...] }` — list value (implies VALUE state)
  - `mapValue: { entries: {...} }` — map value (implies VALUE state)
- `GetSettings` resolves inheritance top-down and annotates each setting with `scopeLevel`. No descriptions.
- `GetSettingsForEdit` returns a `levels` array — one entry per scope level covered by the request key, ordered broadest to most specific. Each entry is `{ key, settings, version }` where `key` is a partial key identifying that level. A level with no record has `version: 0`. Descriptions are included; no `scopeLevel`.
- Map settings (`environmentVariables`) are **additive** on `GetSettings`: parent entries first, child entries merged on top (child wins on key conflict). `scopeLevel` reflects the most specific scope in the resolution chain.

---

## Starting State

Org `acme` has baseline settings from provisioning. The project `analytics` in
domain `production` was previously configured with a `minCpu` override.
No domain settings exist yet.

**Database**

| id | key                         | data (JSONB)                                                                                                                                                                               | version |
|----|-----------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|
| 1  | `acme::`                    | `{"run":{"defaultQueue":{"stringValue":"default"}},"taskResource":{"defaults":{"minCpu":{"stringValue":"500m"}}},"environmentVariables":{"mapValue":{"entries":{"LOG_LEVEL":"info","REGION":"us-east-1"}}}}` | 1       |
| 2  | `acme:production:analytics` | `{"taskResource":{"defaults":{"minCpu":{"stringValue":"1000m"}}}}`                                                                                                                         | 1       |

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
  "key": { "org": "acme" },
  "settings": {
    "run": {
      "defaultQueue": { "stringValue": "default", "scopeLevel": "ORG" }
    },
    "taskResource": {
      "defaults": {
        "minCpu": { "stringValue": "500m", "scopeLevel": "ORG" }
      }
    },
    "environmentVariables": {
      "mapValue": { "entries": { "LOG_LEVEL": "info", "REGION": "us-east-1" } },
      "scopeLevel": "ORG"
    }
  }
}
```

**Database:** unchanged.

---

## Step 2 — Update project settings

### 2a — GetSettingsForEdit at project scope

Before editing, fetch the current stored state at all levels. This returns every
setting field at every scope — inheritance is not applied — with descriptions.

**Request — `GetSettingsForEdit`**
```json
{
  "key": { "org": "acme", "domain": "production", "project": "analytics" }
}
```

**Response**
```json
{
  "key": { "org": "acme", "domain": "production", "project": "analytics" },
  "levels": [
    {
      "key": { "org": "acme" },
      "settings": {
        "run": {
          "defaultQueue": { "stringValue": "default", "description": "Queue that runs are submitted to" }
        },
        "taskResource": {
          "defaults": {
            "minCpu": { "stringValue": "500m", "description": "Minimum CPU requested for task pods" }
          }
        },
        "environmentVariables": {
          "mapValue": { "entries": { "LOG_LEVEL": "info", "REGION": "us-east-1" } },
          "description": "Environment variables injected into task pods — additive across scopes"
        }
      },
      "version": "1"
    },
    {
      "key": { "org": "acme", "domain": "production", "project": "analytics" },
      "settings": {
        "run": {
          "defaultQueue": { "description": "Queue that runs are submitted to" }
        },
        "taskResource": {
          "defaults": {
            "minCpu": { "stringValue": "1000m", "description": "Minimum CPU requested for task pods" }
          }
        },
        "environmentVariables": {
          "description": "Environment variables injected into task pods — additive across scopes"
        }
      },
      "version": "1"
    }
  ]
}
```

The `levels` array runs from broadest to most specific. Org shows its full values.
Domain is not present no record exists yet,
so `CreateSettings` would be used to write to it rather than `UpdateSettings`.
Project shows the existing `minCpu: 1000m`; `defaultQueue` and `environmentVariables` are empty
(INHERIT). Use `version: 1` from the project entry for the update.

**Database:** unchanged.

---

### 2b — UpdateSettings at project scope

Increase `minCpu` for this compute-heavy project and add a project-specific env var.
`defaultQueue` is left as INHERIT. Supply `version: 1` from `project` above.

**Request — `UpdateSettings`**
```json
{
  "key": { "org": "acme", "domain": "production", "project": "analytics" },
  "settings": {
    "run": {
      "defaultQueue": {}
    },
    "taskResource": {
      "defaults": {
        "minCpu": { "stringValue": "2000m" }
      }
    },
    "environmentVariables": { "mapValue": { "entries": { "LOG_LEVEL": "debug" } } }
  },
  "version": "1"
}
```

**Response**
```json
{
  "key": { "org": "acme", "domain": "production", "project": "analytics" },
  "settings": {
    "run": {
      "defaultQueue": {}
    },
    "taskResource": {
      "defaults": {
        "minCpu": { "stringValue": "2000m" }
      }
    },
    "environmentVariables": { "mapValue": { "entries": { "LOG_LEVEL": "debug" } } }
  },
  "version": "2"
}
```

The response echoes back exactly what was stored — no inheritance, no descriptions.
`version` is now `2`.

**Database**

| id | key                         | data (JSONB)                                                                                                                            | version |
|----|-----------------------------|-----------------------------------------------------------------------------------------------------------------------------------------|---------|
| 1  | `acme::`                    | `{"run":{"defaultQueue":{"stringValue":"default"}},...}` (unchanged)                                                                    | 1       |
| 2  | `acme:production:analytics` | `{"taskResource":{"defaults":{"minCpu":{"stringValue":"2000m"}}},"environmentVariables":{"mapValue":{"entries":{"LOG_LEVEL":"debug"}}}}` | 2       |

`run.defaultQueue` is absent from the project row — INHERIT is the zero value and
is never written to the database.

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
  "key": { "org": "acme", "domain": "production", "project": "analytics" },
  "settings": {
    "run": {
      "defaultQueue": { "stringValue": "default", "scopeLevel": "ORG" }
    },
    "taskResource": {
      "defaults": {
        "minCpu": { "stringValue": "2000m", "scopeLevel": "PROJECT" }
      }
    },
    "environmentVariables": {
      "mapValue": { "entries": { "LOG_LEVEL": "debug", "REGION": "us-east-1" } },
      "scopeLevel": "PROJECT"
    }
  }
}
```

- `defaultQueue` — no override in the chain; resolved from org.
- `minCpu` — project override from step 2b; `scopeLevel` is PROJECT.
- `environmentVariables` — additive merge: org contributes `LOG_LEVEL=info` and `REGION=us-east-1`, project contributes `LOG_LEVEL=debug` which overrides the org value. `scopeLevel` is PROJECT, the most specific scope in the resolution chain.

**Database:** unchanged.

---

## Step 4 — CreateSettings at domain scope

The `production` domain has no settings record yet (absent from step 2a).
Create one to route all production runs to a dedicated queue.

**Request — `CreateSettings`**
```json
{
  "key": { "org": "acme", "domain": "production" },
  "settings": {
    "run": {
      "defaultQueue": { "stringValue": "fast-queue" }
    },
    "taskResource": {
      "defaults": {
        "minCpu": {}
      }
    },
    "environmentVariables": {}
  }
}
```

**Response**
```json
{
  "key": { "org": "acme", "domain": "production" },
  "settings": {
    "run": {
      "defaultQueue": { "stringValue": "fast-queue" }
    },
    "taskResource": {
      "defaults": {
        "minCpu": {}
      }
    },
    "environmentVariables": {}
  },
  "version": "1"
}
```

**Database**

| id | key                         | data (JSONB)                                                                                                                            | version |
|----|-----------------------------|-----------------------------------------------------------------------------------------------------------------------------------------|---------|
| 1  | `acme::`                    | `{"run":{"defaultQueue":{"stringValue":"default"}},...}` (unchanged)                                                                    | 1       |
| 3  | `acme:production:`          | `{"run":{"defaultQueue":{"stringValue":"fast-queue"}}}`                                                                                 | 1       |
| 2  | `acme:production:analytics` | `{"taskResource":{"defaults":{"minCpu":{"stringValue":"2000m"}}},"environmentVariables":{"mapValue":{"entries":{"LOG_LEVEL":"debug"}}}}` (unchanged) | 2       |

Only `defaultQueue` is written to the domain row — INHERIT fields are not stored.

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
  "key": { "org": "acme", "domain": "production", "project": "analytics" },
  "settings": {
    "run": {
      "defaultQueue": { "stringValue": "fast-queue", "scopeLevel": "DOMAIN" }
    },
    "taskResource": {
      "defaults": {
        "minCpu": { "stringValue": "2000m", "scopeLevel": "PROJECT" }
      }
    },
    "environmentVariables": {
      "mapValue": { "entries": { "LOG_LEVEL": "debug", "REGION": "us-east-1" } },
      "scopeLevel": "PROJECT"
    }
  }
}
```

- `defaultQueue` — now resolves to `fast-queue` from the domain, overriding org's `default`. Nothing in the project record changed; the domain insert was enough.
- `minCpu` — still `2000m` from the project. The domain did not override it.
- `environmentVariables` — still the additive merge of org (`LOG_LEVEL=info`, `REGION=us-east-1`) + project (`LOG_LEVEL=debug`). The domain contributed no env vars, so the result is unchanged and `scopeLevel` remains PROJECT.

**Database:** unchanged.

---

## Summary

| Step | Scope touched | What changed |
|------|--------------|--------------|
| 1 | — | Read org settings |
| 2a | project (read) | GetSettingsForEdit: two levels returned — org with full values, project with existing `minCpu: 1000m`; domain has no record and is omitted |
| 2b | project (write) | Updated `minCpu` 1000m→2000m, added `environmentVariables={LOG_LEVEL:debug}`; version 1→2 |
| 3 | — | Read project: org+project merge, no domain yet; `LOG_LEVEL` overridden by project |
| 4 | domain (create) | Set `defaultQueue=fast-queue` at domain; `version: 0` → 1 |
| 5 | — | Read project: domain now intercepts `defaultQueue`; env vars and minCpu unchanged |

The key insight: changes at one scope are immediately visible to all child scopes
on the next `GetSettings` call. No project-level record needs to be touched when
a domain setting is added.
