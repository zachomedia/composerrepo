package repository

import (
	"fmt"
	"strings"

	"github.com/zachomedia/composerrepo/pkg/composer"
)

type Input interface {
	Init(id string, conf map[string]interface{}) error

	GetID() string
	GetName() string
	GetPackages() (Packages, error)
	GetPackage(packageName string) (composer.PackageVersions, error)
}

type Transform interface {
	Init(id string, conf map[string]interface{}) error

	Skip(input *Input, pkgName string) (bool, error)
	Transform(input *Input, pkg *composer.Package) error
}

type Output interface {
	Init(conf map[string]interface{}) error

	GetBasePath() string

	GetRepository() (*Repository, error)
	Get(name string) (*Repository, error)
	WriteRepository(repo *Repository) error
	Write(name string, repo *Repository) (string, error)
}

type Config struct {
	UseProviders bool
	Inputs       map[string]Input
	Transformers map[string]Transform
	Output       Output
}

type Reference struct {
	SHA256 string `json:"sha256"`
}

type Packages map[string]composer.PackageVersions

type Repository struct {
	Packages         Packages              `json:"packages,omitempty"`
	Providers        map[string]*Reference `json:"providers,omitempty"`
	ProviderIncludes map[string]*Reference `json:"provider-includes,omitempty"`
	ProvidersURL     string                `json:"providers-url,omitempty"`
}

type PackageInfo struct {
	InputID     string
	PackageName string
}

// Generate generates the repository.
func Generate(conf *Config) error {
	repo := &Repository{}

	if conf.UseProviders {
		repo.ProviderIncludes = make(map[string]*Reference, 0)
		repo.ProvidersURL = fmt.Sprintf("%s/p/%%package%%$%%hash%%.json", conf.Output.GetBasePath())
	} else {
		repo.Packages = make(Packages)
	}

	// If UseProviders is false, save packages directly to packages.json
	for _, connector := range conf.Inputs {
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

				hash, err := conf.Output.Write(fmt.Sprintf("p/%s$%%hash%%.json", name), &Repository{
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
			hash, err := conf.Output.Write(providerPath, provider)
			if err != nil {
				return err
			}

			repo.ProviderIncludes[providerPath] = &Reference{
				SHA256: hash,
			}
		}
	}

	return conf.Output.WriteRepository(repo)
}

// Update updates packages in the repository.
func Update(conf *Config, packageInfos []*PackageInfo) error {
	// Read the current repository
	repo, err := conf.Output.GetRepository()
	if err != nil {
		return err
	}

	providers := make(map[string]*Repository)

	for _, packageInfo := range packageInfos {
		pkg, err := conf.Inputs[packageInfo.InputID].GetPackage(packageInfo.PackageName)
		if err != nil {
			return err
		}

		if conf.UseProviders {
			if _, ok := repo.ProviderIncludes[packageInfo.InputID]; !ok {
				// Find the provider
				providerID := fmt.Sprintf("p/provider-%s$%%hash%%.json", packageInfo.InputID)
				providerInfo, ok := repo.ProviderIncludes[providerID]
				if !ok {
					return fmt.Errorf("No connector matching %q", packageInfo.InputID)
				}

				provider, err := conf.Output.Get(strings.Replace(providerID, "%hash%", providerInfo.SHA256, -1))
				if err != nil {
					return err
				}
				providers[packageInfo.InputID] = provider
			}

			hash, err := conf.Output.Write(fmt.Sprintf("p/%s$%%hash%%.json", packageInfo.PackageName), &Repository{
				Packages: Packages{
					packageInfo.PackageName: pkg,
				},
			})
			if err != nil {
				return err
			}

			providers[packageInfo.InputID].Providers[packageInfo.PackageName] = &Reference{
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
			hash, err := conf.Output.Write(providerPath, provider)
			if err != nil {
				return err
			}

			repo.ProviderIncludes[providerPath] = &Reference{
				SHA256: hash,
			}
		}
	}

	return conf.Output.WriteRepository(repo)
}
