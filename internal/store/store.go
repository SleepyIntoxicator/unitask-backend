package store

type Store interface {
	Auth() AuthRepository
	User() UserRepository
	Task() TaskRepository
	University() UniversityRepository
	Group() GroupRepository
	Subject() SubjectRepository
}
