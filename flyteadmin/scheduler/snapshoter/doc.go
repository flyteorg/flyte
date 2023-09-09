// Package snapshoter
// This package provides the ability to snapshot all the schedules in the scheduler job store and persist them in the DB
// in GOB binary format. Also it provides ability to bootstrap the scheduler from this snapshot so that the scheduler
// can run catchup for all the schedules from the snapshoted time.
package snapshoter
