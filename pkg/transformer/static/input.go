package static

import (
	"github.com/vmihailenco/msgpack"
	"github.com/zachomedia/composerrepo/pkg/composer"
)

type StaticInput struct {
	ID       string
	Packages composer.Packages `msgpack:"packages"`
}

func (input *StaticInput) Init(id string, conf map[string]interface{}) error {
	input.ID = id
	input.Packages = make(composer.Packages)

	for rawPackageName, rawPackageVersions := range conf["packages"].(map[interface{}]interface{}) {
		packageName := rawPackageName.(string)
		packageVersions := rawPackageVersions.(map[interface{}]interface{})

		input.Packages[packageName] = make(composer.PackageVersions)

		for rawVersion, rawPackage := range packageVersions {
			version := rawVersion.(string)

			b, _ := msgpack.Marshal(rawPackage)
			var pkg composer.Package
			msgpack.Unmarshal(b, &pkg)

			pkg.Name = packageName
			pkg.Version = version
			input.Packages[packageName][version] = &pkg
		}
	}

	return nil
}

func (input *StaticInput) GetID() string {
	return input.ID
}

func (input *StaticInput) GetName() string {
	return input.ID
}

func (input *StaticInput) GetPackages() (composer.Packages, error) {

	return input.Packages, nil
}

func (input *StaticInput) GetPackage(packageName string) (composer.PackageVersions, error) {
	return input.Packages[packageName], nil
}
