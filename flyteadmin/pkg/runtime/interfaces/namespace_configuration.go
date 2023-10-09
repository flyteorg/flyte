package interfaces

type NamespaceMappingConfig struct {
	Mapping      string       `json:"mapping"` // Deprecated
	Template     string       `json:"template"`
	TemplateData TemplateData `json:"templateData"`
}

//go:generate mockery -name NamespaceMappingConfiguration -output=../mocks -case=underscore

type NamespaceMappingConfiguration interface {
	GetNamespaceTemplate() string
}
