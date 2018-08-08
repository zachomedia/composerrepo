package composer

import (
	"fmt"
)

type Reference struct {
	SHA256 string `json:"sha256"`
}

type Packages map[string]PackageVersions

type Repository struct {
	Packages         Packages              `json:"packages,omitempty"`
	Providers        map[string]*Reference `json:"providers,omitempty"`
	ProviderIncludes map[string]*Reference `json:"provider-includes,omitempty"`
	ProvidersURL     string                `json:"providers-url,omitempty"`
}

type GenerateConfig struct {
	OutputConnector OutputConnector
	Connectors      []Connector
	UseProviders    bool
}

// Generate generates the repository.
func Generate(conf *GenerateConfig) error {
	repo := &Repository{}

	if conf.UseProviders {
		repo.ProviderIncludes = make(map[string]*Reference, 0)
		repo.ProvidersURL = fmt.Sprintf("%s/p/%%package%%$%%hash%%.json", conf.OutputConnector.GetBasePath())
	} else {
		repo.Packages = make(Packages)
	}

	// If UseProviders is false, save packages directly to packages.json
	for _, connector := range conf.Connectors {
		provider := &Repository{
			Providers: make(map[string]*Reference),
		}

		pkgs, err := connector.GetPackages()
		if err != nil {
			return err
		}

		for name, versions := range pkgs {
			if conf.UseProviders {
				// Add a unique ID to all versions
				for _, version := range versions {
					version.UID = fmt.Sprintf("%s@%s", version.Name, version.Version)
				}

				hash, err := conf.OutputConnector.Write(fmt.Sprintf("p/%s$%%hash%%.json", name), &Repository{
					Packages: Packages{
						name: versions,
					},
				})
				if err != nil {
					return err
				}

				provider.Providers[name] = &Reference{
					SHA256: hash,
				}
			} else {
				repo.Packages[name] = versions
			}
		}

		if conf.UseProviders {
			providerPath := fmt.Sprintf("p/provider-%s$%%hash%%.json", connector.GetName())
			hash, err := conf.OutputConnector.Write(providerPath, provider)
			if err != nil {
				return err
			}

			repo.ProviderIncludes[providerPath] = &Reference{
				SHA256: hash,
			}
		}
	}

	return conf.OutputConnector.WriteRepository(repo)
}
