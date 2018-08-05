package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/zachomedia/composerrepo/pkg/composer"
	yaml "gopkg.in/yaml.v2"

	"github.com/urfave/cli"
	gogitlab "github.com/xanzy/go-gitlab"
	"github.com/zachomedia/composerrepo/pkg/gitlab"
)

func newGitLabConnector(url, token, group string) (*gitlab.GitLabConnector, error) {
	client := gogitlab.NewClient(nil, token)
	client.SetBaseURL(url)

	return gitlab.NewConnector(client, group)
}

func getConnectorsConfig(c *cli.Context) ([]map[string]string, error) {
	// Read config
	config := make([]map[string]string, 0)

	f, err := os.Open(c.String("config"))
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

func generate(c *cli.Context) error {
	connectors := make([]composer.Connector, 0)

	config, err := getConnectorsConfig(c)
	if err != nil {
		return err
	}

	for _, connectorDef := range config {
		if _, ok := connectorDef["type"]; !ok {
			return errors.New("No connector type defined")
		}

		switch connectorDef["type"] {
		case "gitlab":
			connector, err := newGitLabConnector(connectorDef["url"], connectorDef["token"], connectorDef["group"])
			if err != nil {
				return err
			}

			connectors = append(connectors, connector)
		default:
			return fmt.Errorf("Unknown connector %q", connectorDef["type"])
		}
	}

	repo, err := composer.Generate(connectors...)
	if err != nil {
		return err
	}

	b, _ := json.Marshal(repo)
	fmt.Println(string(b))

	return nil
}

func main() {
	app := cli.NewApp()

	app.Name = "repoctl"
	app.Usage = "Generate a composer repository"
	app.Version = "0.1.0"

	app.Commands = []cli.Command{
		{
			Name:    "generate",
			Aliases: []string{"g"},
			Usage:   "Generate the packages.json file",
			Action:  generate,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "config, c",
					EnvVar: "CONNECTORS_CONFIG",
					Value:  "connectors.yml",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
