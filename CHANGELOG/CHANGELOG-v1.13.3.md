# Flyte 1.13.3 Release Notes

This is an in-between release in preparation for the big 1.14 release slated for early December.

## What's Changed
* Add finalizer to avoid premature CRD Deletion by @RRap0so in https://github.com/flyteorg/flyte/pull/5788
* Handle CORS in secure connections by @eapolinario in https://github.com/flyteorg/flyte/pull/5855
* Upstream contributions from Union.ai by @andrewwdye in https://github.com/flyteorg/flyte/pull/5769
  - Fix cluster pool assignment validation (@iaroslav-ciupin )
  - Don't send inputURI for start-node (@iaroslav-ciupin )
  - Unexpectedly deleted pod metrics (@iaroslav-ciupin )
  - CreateDownloadLink: Head before signing (@iaroslav-ciupin )
  - Fix metrics scale division in timer (@iaroslav-ciupin )
  - Add histogram stopwatch to stow storage (@andrewwdye )
  - Override ArrayNode log links with map plugin (@hamersaw )
  - Fix k3d local setup prefix (@andrewwdye )
  - adjust Dask LogName to (Dask Runner Logs) (@fiedlerNr9 )
  - Dask dashboard should have a separate log config (@EngHabu )
  - Log and monitor failures to validate access tokens (@katrogan )
  - Fix type assertion when an event is missed while connection to apiser… (@EngHabu )
  - Add read replica host config and connection (@squiishyy )
  - added lock to memstore make threadsafe (@hamersaw )
  - Move storage cache settings to correct location (@mbarrien )
  - Add config for grpc MaxMessageSizeBytes (@andrewwdye )
  - Add org to CreateUploadLocation (@katrogan )
  - Histogram Bucket Options (@squiishyy )
  - Add client-go metrics (@andrewwdye )
  - Enqueue owner on launchplan terminal state (@hamersaw )
  - Add configuration for launchplan cache resync duration (@hamersaw )
  - Overlap fetching input and output data (@andrewwdye )
  - Fix async notifications tests (@andrewwdye )
  - Overlap FutureFileReader blob store writes/reads (@andrewwdye )
  - Overlap create execution blob store reads/writes (@andrewwdye )
  
**Full Changelog**: https://github.com/flyteorg/flyte/compare/v1.13.2...v1.13.3
