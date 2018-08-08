package composer

type OutputConnector interface {
	GetBasePath() string

	WriteRepository(repo *Repository) error
	Write(name string, repo *Repository) (string, error)
}

type Connector interface {
	GetName() string
	GetPackages() (Packages, error)
}
