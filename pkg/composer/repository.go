package composer

func Generate(connectors ...Connector) (*Repository, error) {
	repo := &Repository{
		Packages: make(map[string]map[string]*Package),
	}

	for _, connector := range connectors {
		pkgs, err := connector.GetPackages()
		if err != nil {
			return nil, err
		}

		for pkgName, pkg := range pkgs {
			repo.Packages[pkgName] = pkg
		}
	}

	return repo, nil
}
