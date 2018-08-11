package config

import (
	"errors"
	"fmt"
	"io"

	"github.com/zachomedia/composerrepo/pkg/composer/repository"
	"github.com/zachomedia/composerrepo/pkg/input/gitlab"
	"github.com/zachomedia/composerrepo/pkg/output/file"
	yaml "gopkg.in/yaml.v2"
)

var InputTypes = map[string]repository.Input{
	"gitlab": &gitlab.GitLabInput{},
}

var TransformerTypes = map[string]repository.Transform{}

var OutputTypes = map[string]repository.Output{
	"file": &file.FileOutput{},
}

func ConfigFromYAML(reader io.Reader) (*repository.Config, error) {
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
	conf := &repository.Config{
		UseProviders: rawConfig.UseProviders,
		Inputs:       make(map[string]repository.Input),
		Transformers: make(map[string]repository.Transform),
	}

	for k, raw := range rawConfig.Inputs {
		inputType, ok := InputTypes[raw["type"].(string)]
		if !ok {
			return nil, fmt.Errorf("Unknown input type %q for %q", raw["type"].(string), k)
		}

		conf.Inputs[k] = inputType
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

		conf.Transformers[k] = transformerType
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

	conf.Output = outputType
	err = conf.Output.Init(rawConfig.Output)
	if err != nil {
		return nil, err
	}

	return conf, nil
}
