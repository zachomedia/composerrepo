package composer

import (
	"fmt"
	"strings"
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
	Connectors      map[string]Connector
	UseProviders    bool
}

type PackageInfo struct {
	ConnectorID string
	PackageName string
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
			providerPath := fmt.Sprintf("p/provider-%s$%%hash%%.json", connector.GetID())
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

// Update updates packages in the repository.
func Update(conf *GenerateConfig, packageInfos []*PackageInfo) error {
	// Read the current repository
	repo, err := conf.OutputConnector.GetRepository()
	if err != nil {
		return err
	}

	providers := make(map[string]*Repository)

	for _, packageInfo := range packageInfos {
		pkg, err := conf.Connectors[packageInfo.ConnectorID].GetPackage(packageInfo.PackageName)
		if err != nil {
			return err
		}

		if conf.UseProviders {
			if _, ok := repo.ProviderIncludes[packageInfo.ConnectorID]; !ok {
				// Find the provider
				providerID := fmt.Sprintf("p/provider-%s$%%hash%%.json", packageInfo.ConnectorID)
				providerInfo, ok := repo.ProviderIncludes[providerID]
				if !ok {
					return fmt.Errorf("No connector matching %q", packageInfo.ConnectorID)
				}

				provider, err := conf.OutputConnector.Get(strings.Replace(providerID, "%hash%", providerInfo.SHA256, -1))
				if err != nil {
					return err
				}
				providers[packageInfo.ConnectorID] = provider
			}

			hash, err := conf.OutputConnector.Write(fmt.Sprintf("p/%s$%%hash%%.json", packageInfo.PackageName), &Repository{
				Packages: Packages{
					packageInfo.PackageName: pkg,
				},
			})
			if err != nil {
				return err
			}

			providers[packageInfo.ConnectorID].Providers[packageInfo.PackageName] = &Reference{
				SHA256: hash,
			}
		} else {
			repo.Packages[packageInfo.PackageName] = pkg
		}
	}

	// Write final
	if conf.UseProviders {
		for providerID, provider := range providers {
			providerPath := fmt.Sprintf("p/provider-%s$%%hash%%.json", providerID)
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
