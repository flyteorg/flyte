package common

type ResourceScope interface {
	GetOrg() string
	GetProject() string
	GetDomain() string
}

func (i resourceScope) GetProject() string {
	return i.project
}

func (i resourceScope) GetOrg() string {
	return i.org
}

func (i resourceScope) GetDomain() string {
	return i.domain
}

type resourceScope struct {
	org     string
	project string
	domain  string
}

func NewResourceScope(org, project, domain string) ResourceScope {
	return &resourceScope{project: project, org: org, domain: domain}
}

type ProjectScope interface {
	GetProject() string
	GetOrg() string
}

func NewProjectResourceScope(id ProjectScope) ResourceScope {
	return &resourceScope{project: id.GetProject(), org: id.GetOrg()}
}
