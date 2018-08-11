package main

import (
	"log"
	"os"
	"strings"

	"github.com/zachomedia/composerrepo/pkg/composer/repository"
	"github.com/zachomedia/composerrepo/pkg/config"

	"github.com/urfave/cli"
)

func getConfig(c *cli.Context) (*repository.Config, error) {
	f, err := os.Open(c.GlobalString("config"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return config.ConfigFromYAML(f)
}

func generate(c *cli.Context) error {
	conf, err := getConfig(c)
	if err != nil {
		return err
	}

	return repository.Generate(conf)
}

func update(c *cli.Context) error {
	conf, err := getConfig(c)
	if err != nil {
		return err
	}

	packages := make([]*repository.PackageInfo, 0)

	for _, arg := range c.Args() {
		components := strings.SplitN(arg, ":", 2)

		packages = append(packages, &repository.PackageInfo{
			InputID:     components[0],
			PackageName: components[1],
		})
	}

	return repository.Update(conf, packages)
}

func main() {
	app := cli.NewApp()

	app.Name = "repoctl"
	app.Usage = "Generate a composer repository"
	app.Version = "0.1.0"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config, c",
			Usage:  "Location of the YAML configuration file.",
			EnvVar: "CONFIG_PATH",
			Value:  "config.yml",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "generate",
			Aliases: []string{"g"},
			Usage:   "Generate the repository.",
			Action:  generate,
		},
		{
			Name:    "update",
			Aliases: []string{"u"},
			Usage:   "Updates a specific package in the repository.",
			Action:  update,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
