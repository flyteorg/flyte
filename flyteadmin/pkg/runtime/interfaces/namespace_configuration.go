package interfaces

type NamespaceMappingConfig struct {
	Mapping      string       `json:"mapping"` // Deprecated
	Template     string       `json:"template"`
	TemplateData TemplateData `json:"templateData"`
}

type NamespaceMappingConfiguration interface {
	GetNamespaceTemplate() string
}
