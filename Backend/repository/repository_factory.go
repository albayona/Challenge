package repository

// RepositoryFactory handles repository creation and management
type RepositoryFactory struct {
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory() *RepositoryFactory {
	return &RepositoryFactory{}
}

// CreateDataRepository creates a new data repository instance
func (f *RepositoryFactory) CreateDataRepository() DataRepositoryInterface {
	// Create CockroachDB repository - it will handle its own connection
	repo := NewCockroachDBRepository(nil)
	// Connect to the database
	repo.Connect()
	return repo
}
