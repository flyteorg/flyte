package interfaces

type Repository interface {
	ActionRepo() ActionRepo
	TaskRepo() TaskRepo
}
