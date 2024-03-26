// Package core
// This is core package for the scheduler which includes
//   - scheduler interface
//   - scheduler implementation using gocron https://github.com/robfig/cron
//   - updater which updates the schedules in the scheduler by reading periodically from the DB
//   - snapshot runner which snapshot the schedules with there last exec times so that it can be used as check point
//     in case of a crash. After a crash the scheduler replays the schedules from the last recorded snapshot.
//     It relies on the admin idempotency aspect to fail executions if the execution with a scheduled time already exists with it.
package core
