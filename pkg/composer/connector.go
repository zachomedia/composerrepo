package composer

type OutputConnector interface {
	GetBasePath() string

	GetRepository() (*Repository, error)
	Get(name string) (*Repository, error)
	WriteRepository(repo *Repository) error
	Write(name string, repo *Repository) (string, error)
}

type Connector interface {
	GetID() string
	GetName() string
	GetPackages() (Packages, error)
	GetPackage(packageName string) (PackageVersions, error)
}
