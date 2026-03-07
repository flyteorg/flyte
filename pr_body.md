## What changed

Populate missing fields in the `WatchActions` response so that `EnrichedAction` messages include all expected data:

- **`action.id.run.name`**: Fixed `GetRunName()` on the `Action` model to extract the run name from the `ActionSpec` JSON for child actions (was returning empty string).
- **`action.status.start_time`**: Set from `action.CreatedAt`.
- **`action.status.end_time`**: Set from `action.EndedAt` when the action has terminated.
- **`action.status.attempts`**: Set to `1` (default single attempt).
- **`action.status.cache_status`**: Extracted from `ActionDetails` JSON blob.
- **`action.status.duration_ms`**: Computed from `EndedAt - CreatedAt` when the action has terminated.
- **`children_phase_counts`**: Added `GetChildrenPhaseCounts` repository method that queries child action phase counts grouped by parent action name, and wired it into `WatchActions`.

## Why

The `WatchActions` streaming endpoint was returning incomplete `EnrichedAction` messages — several fields defined in the proto were not being populated. This made the response unusable for clients that depend on status timing, cache info, or children aggregation data.

Fixes #6982

## Testing

- Added unit tests for `convertActionToEnrichedProto` covering:
  - All fields populated (start_time, end_time, duration_ms, attempts, cache_status, run.name)
  - Running actions without end time (no end_time or duration_ms)
  - Root actions (run name = action name)
- Added unit tests for `extractCacheStatus` (empty and populated details)
- Added unit tests for `GetRunName` (root, child with spec, child with empty/nil spec)
- All existing tests continue to pass
