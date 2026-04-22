package interfaces

type Repository interface {
	CachedOutputRepo() CachedOutputRepo
	ReservationRepo() ReservationRepo
}
