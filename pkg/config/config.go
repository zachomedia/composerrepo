package config

import (
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/zachomedia/composerrepo/pkg/composer"
	"github.com/zachomedia/composerrepo/pkg/input/gitlab"
	"github.com/zachomedia/composerrepo/pkg/output/azure"
	"github.com/zachomedia/composerrepo/pkg/output/file"
	yaml "gopkg.in/yaml.v2"
)

var InputTypes = map[string]composer.Input{
	"gitlab": &gitlab.GitLabInput{},
}

var TransformerTypes = map[string]composer.Transform{}

var OutputTypes = map[string]composer.Output{
	"azure": &azure.AzureOutput{},
	"file":  &file.FileOutput{},
}

func ConfigFromYAML(reader io.Reader) (*composer.Config, error) {
	type config struct {
		UseProviders bool                              `yaml:"providers"`
		Inputs       map[string]map[string]interface{} `yaml:"inputs"`
		Transformers map[string]map[string]interface{} `yaml:"transformers"`
		Output       map[string]interface{}            `yaml:"output"`
	}

	rawConfig := config{}

	// Open the YAML file and decode it
	decoder := yaml.NewDecoder(reader)
	decoder.SetStrict(true)
	err := decoder.Decode(&rawConfig)
	if err != nil {
		return nil, err
	}

	// Initialize each config item
	conf := &composer.Config{
		UseProviders: rawConfig.UseProviders,
		Inputs:       make(map[string]composer.Input),
		Transformers: make(map[string]composer.Transform),
	}

	for k, raw := range rawConfig.Inputs {
		inputType, ok := InputTypes[raw["type"].(string)]
		if !ok {
			return nil, fmt.Errorf("Unknown input type %q for %q", raw["type"].(string), k)
		}

		conf.Inputs[k] = reflect.New(reflect.ValueOf(inputType).Elem().Type()).Interface().(composer.Input)
		err = conf.Inputs[k].Init(k, raw)
		if err != nil {
			return nil, err
		}
	}

	for k, raw := range rawConfig.Transformers {
		transformerType, ok := TransformerTypes[raw["type"].(string)]
		if !ok {
			return nil, fmt.Errorf("Unknown transformer type %q for %q", raw["type"].(string), k)
		}

		conf.Transformers[k] = reflect.New(reflect.ValueOf(transformerType).Elem().Type()).Interface().(composer.Transform)
		err = conf.Transformers[k].Init(k, raw)
		if err != nil {
			return nil, err
		}
	}

	if rawConfig.Output == nil {
		return nil, errors.New("Output config cannot be empty")
	}
	outputType, ok := OutputTypes[rawConfig.Output["type"].(string)]
	if !ok {
		return nil, fmt.Errorf("Unknown output type %q", rawConfig.Output["type"].(string))
	}

	conf.Output = reflect.New(reflect.ValueOf(outputType).Elem().Type()).Interface().(composer.Output)
	err = conf.Output.Init(rawConfig.Output)
	if err != nil {
		return nil, err
	}

	return conf, nil
}
