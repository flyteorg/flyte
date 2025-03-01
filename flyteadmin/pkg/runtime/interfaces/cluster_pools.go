package interfaces

//go:generate mockery --name ClusterPoolAssignmentConfiguration --output=mocks --case=underscore --with-expecter

type ClusterPoolAssignment struct {
	Pool string `json:"pool"`
}

type ClusterPoolAssignments = map[DomainName]ClusterPoolAssignment

type ClusterPoolAssignmentConfig struct {
	ClusterPoolAssignments ClusterPoolAssignments `json:"clusterPoolAssignments"`
}

type ClusterPoolAssignmentConfiguration interface {
	GetClusterPoolAssignments() ClusterPoolAssignments
}
