package composer

type Connector interface {
	GetPackages() (map[string]map[string]*Package, error)
}
