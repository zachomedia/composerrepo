package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/zachomedia/composerrepo/pkg/composer"
	"github.com/zachomedia/composerrepo/pkg/output-connectors/file"
	yaml "gopkg.in/yaml.v2"

	"github.com/urfave/cli"
	gogitlab "github.com/xanzy/go-gitlab"
	"github.com/zachomedia/composerrepo/pkg/connectors/gitlab"
)

func newGitLabConnector(id, url, token, group string) (*gitlab.GitLabConnector, error) {
	client := gogitlab.NewClient(nil, token)
	client.SetBaseURL(url)

	return gitlab.NewConnector(id, client, group)
}

func getConnectorsConfig(c *cli.Context) ([]map[string]string, error) {
	// Read config
	config := make([]map[string]string, 0)

	f, err := os.Open(c.GlobalString("config"))
	if err != nil {
		return nil, err
	}

	dec := yaml.NewDecoder(f)
	err = dec.Decode(&config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func loadConnectors(c *cli.Context) (map[string]composer.Connector, error) {
	connectors := make(map[string]composer.Connector, 0)

	config, err := getConnectorsConfig(c)
	if err != nil {
		return nil, err
	}

	for _, connectorDef := range config {
		if _, ok := connectorDef["type"]; !ok {
			return nil, errors.New("No connector type defined")
		}

		switch connectorDef["type"] {
		case "gitlab":
			connector, err := newGitLabConnector(connectorDef["id"], connectorDef["url"], connectorDef["token"], connectorDef["group"])
			if err != nil {
				return nil, err
			}

			connectors[connector.GetID()] = connector
		default:
			return nil, fmt.Errorf("Unknown connector %q", connectorDef["type"])
		}
	}

	return connectors, nil
}

func loadOutputConnector(c *cli.Context) (composer.OutputConnector, error) {
	return &file.FileOutput{
		Out:      c.GlobalString("out-dir"),
		BasePath: c.GlobalString("base-path"),
	}, nil
}

func generate(c *cli.Context) error {
	connectors, err := loadConnectors(c)
	if err != nil {
		return err
	}

	output, err := loadOutputConnector(c)
	if err != nil {
		return err
	}

	return composer.Generate(&composer.GenerateConfig{
		OutputConnector: output,
		Connectors:      connectors,
		UseProviders:    c.GlobalBool("providers"),
	})
}

func regenerate(c *cli.Context) error {
	connectors, err := loadConnectors(c)
	if err != nil {
		return err
	}

	output, err := loadOutputConnector(c)
	if err != nil {
		return err
	}

	packages := make([]*composer.PackageInfo, 0)

	for _, arg := range c.Args() {
		components := strings.SplitN(arg, ":", 2)

		packages = append(packages, &composer.PackageInfo{
			ConnectorID: components[0],
			PackageName: components[1],
		})
	}

	return composer.Update(&composer.GenerateConfig{
		OutputConnector: output,
		Connectors:      connectors,
		UseProviders:    c.GlobalBool("providers"),
	}, packages)
}

func main() {
	app := cli.NewApp()

	app.Name = "repoctl"
	app.Usage = "Generate a composer repository"
	app.Version = "0.1.0"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config, c",
			Usage:  "Location of the connectors YAML configuration file.",
			EnvVar: "CONNECTORS_CONFIG",
			Value:  "connectors.yml",
		},
		cli.StringFlag{
			Name:   "out-dir, o",
			Usage:  "Directory to place the generated repository.",
			EnvVar: "OUTPUT_DIRECTORY",
			Value:  "out",
		},
		cli.BoolFlag{
			Name:   "providers, p",
			Usage:  "Seperate packages into providers.",
			EnvVar: "PROVIDERS",
		},
		cli.StringFlag{
			Name:   "base-path, b",
			Usage:  "Base path (web) to use when generating providers (NOTE: must start with a /).",
			EnvVar: "BASE_PATH",
			Value:  "/",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "generate",
			Aliases: []string{"g"},
			Usage:   "Generate the packages.json file.",
			Action:  generate,
		},
		{
			Name:    "regenerate",
			Aliases: []string{"r"},
			Usage:   "Regenerates a specific package.",
			Action:  regenerate,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
