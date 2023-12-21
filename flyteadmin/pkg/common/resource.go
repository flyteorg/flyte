package common

type ResourceIdentifier interface {
	GetProject() string
	GetOrg() string
	GetDomain() string
}

func (i resourceIdentifier) GetProject() string {
	return i.project
}

func (i resourceIdentifier) GetOrg() string {
	return i.org
}

func (i resourceIdentifier) GetDomain() string {
	return i.domain
}

type resourceIdentifier struct {
	project string
	org     string
	domain  string
}

func NewResourceIdentifier(org, project, domain string) ResourceIdentifier {
	return &resourceIdentifier{project: project, org: org, domain: domain}
}

func NewProjectResourceIdentifier(org, project string) ResourceIdentifier {
	return &resourceIdentifier{project: project, org: org}
}
