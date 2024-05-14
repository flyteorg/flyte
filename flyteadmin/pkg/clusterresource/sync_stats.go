package clusterresource

// ResourceSyncStats is a simple struct to track the number of resources created, updated, already there, and errored
type ResourceSyncStats struct {
	Created      int
	Updated      int
	AlreadyThere int
	Errored      int
}

// Add adds the values of the other ResourceSyncStats to this one
func (m *ResourceSyncStats) Add(other ResourceSyncStats) {
	m.Created += other.Created
	m.Updated += other.Updated
	m.AlreadyThere += other.AlreadyThere
	m.Errored += other.Errored
}
