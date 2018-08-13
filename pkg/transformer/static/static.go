package static

import (
	"fmt"
	"log"
	"sort"

	"github.com/zachomedia/composerrepo/pkg/composer"
)

type StaticTransformer struct {
	ID       int
	Packages []string
	Values   map[string]interface{}
}

func (transformer *StaticTransformer) Init(id int, conf map[string]interface{}) error {
	transformer.ID = id
	transformer.Packages = make([]string, 0)
	transformer.Values = make(map[string]interface{})

	confPackages, ok := conf["packages"]
	if ok {
		for _, rPkgName := range confPackages.([]interface{}) {
			transformer.Packages = append(transformer.Packages, rPkgName.(string))
		}
	}
	sort.Strings(transformer.Packages)

	values, ok := conf["values"]
	if ok {
		for k, v := range values.(map[interface{}]interface{}) {
			transformer.Values[k.(string)] = v
		}
	}

	return nil
}

func (transformer *StaticTransformer) GetID() int {
	return transformer.ID
}

func (transformer *StaticTransformer) Transform(input composer.Input, name string, pkg composer.PackageVersions) error {
	indx := sort.SearchStrings(transformer.Packages, name)
	if len(transformer.Packages) > 0 && (indx == len(transformer.Packages) || transformer.Packages[indx] != name) {
		return nil
	}

	log.Printf("Transforming %q", name)

	for k, v := range transformer.Values {
		switch k {
		case "type":
			for _, vers := range pkg {
				vers.Type = v.(string)
			}

		default:
			return fmt.Errorf("Unsupported property %q", k)
		}
	}

	return nil
}
