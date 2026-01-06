package interfaces

//go:generate mockery --output=../mocks --case=underscore --all --with-expecter
type Repository interface {
	ActionRepo() ActionRepo
	TaskRepo() TaskRepo
}
