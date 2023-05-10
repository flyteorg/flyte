# Flyte 1.5.1 Patch Release

This is a patch release that only contains one change - https://github.com/flyteorg/flyteadmin/pull/560
which cherry-picks https://github.com/flyteorg/flyteadmin/pull/554.

PR #554 adds a migration that remediates an issue that we discovered with very old installations of Flyte. Basically one of the tables `node_executions` has a self-referencing foreign key. The main `id` column of the table is a `bigint` whereas the self-foreign-key `parent_id` was an `int`. This was a rooted in an early version of gorm and should not affect most users. Out of an abundance of caution however, we are adding a migration to patch this issue in a manner that minimizes any locking.

## To Deploy
When you deploy this release of Flyte, you should make sure that you have more than one pod for Admin running. (If you are running the flyte-binary helm chart, this patch release does not apply to you at all. All those deployments should already have the correct column type.) When the two new migrations that #554 added runs, the first one may take an extended period of time (hours). However, this is entirely non-blocking as long as there is another Admin instance available to serve traffic.

The second migration is locking, but even on very large tables, this migration was over in ~5 seconds, so you should not see any noticeable downtime whatsoever.

The migration will also check to see that your database falls into this category before running (ie, the `parent_id` and the `id` columns in `node_executions` are mismatched). You can also do check this yourself using psql. If this migration is not needed, the migration will simply mark itself as complete and be a no-op otherwise.
