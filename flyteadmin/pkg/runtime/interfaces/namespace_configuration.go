package interfaces

type NamespaceMappingConfig struct {
	Mapping      string       `json:"mapping"` // Deprecated
	Template     string       `json:"template"`
	TemplateData TemplateData `json:"templateData"`
}

//go:generate mockery-v2 --name NamespaceMappingConfiguration --output=../mocks --case=underscore --with-expecter

type NamespaceMappingConfiguration interface {
	GetNamespaceTemplate() string
}
