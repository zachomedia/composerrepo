package composer

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
)

type Input interface {
	Init(id string, conf map[string]interface{}) error

	GetID() string
	GetName() string
	GetPackages() (Packages, error)
	GetPackage(packageName string) (PackageVersions, error)
}

type Transform interface {
	Init(id string, conf map[string]interface{}) error

	Skip(input *Input, pkgName string) (bool, error)
	Transform(input *Input, pkg *Package) error
}

type Output interface {
	Init(conf map[string]interface{}) error

	GetBasePath() string

	Get(name string) ([]byte, error)
	Write(name string, data []byte) error
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

type Packages map[string]PackageVersions

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

func generateContentsAndHash(obj interface{}) ([]byte, string, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, "", err
	}

	return b, fmt.Sprintf("%x", sha256.Sum256(b)), nil
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

				contents, hash, err := generateContentsAndHash(&Repository{
					Packages: Packages{
						name: versions,
					},
				})
				if err != nil {
					return err
				}

				err = conf.Output.Write(fmt.Sprintf("p/%s$%s.json", name, hash), contents)
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

			contents, hash, err := generateContentsAndHash(provider)
			if err != nil {
				return err
			}

			err = conf.Output.Write(strings.Replace(providerPath, "%hash%", hash, -1), contents)
			if err != nil {
				return err
			}

			repo.ProviderIncludes[providerPath] = &Reference{
				SHA256: hash,
			}
		}
	}

	contents, _, err := generateContentsAndHash(repo)
	if err != nil {
		return err
	}

	return conf.Output.Write("packages.json", contents)
}

// Update updates packages in the repository.
func Update(conf *Config, packageInfos []*PackageInfo) error {
	repo := Repository{}

	// Read the current repository
	repoData, err := conf.Output.Get("packages.json")
	if err != nil {
		return err
	}

	err = json.Unmarshal(repoData, &repo)
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
					return fmt.Errorf("No input matching %q", packageInfo.InputID)
				}

				provider := &Repository{}
				providerData, err := conf.Output.Get(strings.Replace(providerID, "%hash%", providerInfo.SHA256, -1))
				if err != nil {
					return err
				}

				err = json.Unmarshal(providerData, &provider)
				if err != nil {
					return err
				}

				providers[packageInfo.InputID] = provider
			}

			contents, hash, err := generateContentsAndHash(&Repository{
				Packages: Packages{
					packageInfo.PackageName: pkg,
				},
			})
			if err != nil {
				return err
			}

			err = conf.Output.Write(fmt.Sprintf("p/%s$%s.json", packageInfo.PackageName, hash), contents)
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

			contents, hash, err := generateContentsAndHash(provider)
			if err != nil {
				return err
			}

			err = conf.Output.Write(strings.Replace(providerPath, "%hash%", hash, -1), contents)
			if err != nil {
				return err
			}

			repo.ProviderIncludes[providerPath] = &Reference{
				SHA256: hash,
			}
		}
	}

	contents, _, err := generateContentsAndHash(repo)
	if err != nil {
		return err
	}

	return conf.Output.Write("packages.json", contents)
}
