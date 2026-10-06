package interfaces

type NamespaceMappingConfig struct {
	Mapping  string `json:"mapping"` // Deprecated
	Template string `json:"template"`
}

//go:generate mockery --name NamespaceMappingConfiguration --output=../mocks --case=underscore --with-expecter

type NamespaceMappingConfiguration interface {
	GetNamespaceTemplate() string
}
