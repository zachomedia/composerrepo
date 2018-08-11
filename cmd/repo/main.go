package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/zachomedia/composerrepo/pkg/composer"
	"github.com/zachomedia/composerrepo/pkg/config"

	"github.com/urfave/cli"
)

func getConfig(c *cli.Context) (*composer.Config, error) {
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

	return composer.Generate(conf)
}

func update(c *cli.Context) error {
	conf, err := getConfig(c)
	if err != nil {
		return err
	}

	packages := make([]*composer.PackageInfo, 0)

	for _, arg := range c.Args() {
		components := strings.SplitN(arg, ":", 2)

		packages = append(packages, &composer.PackageInfo{
			InputID:     components[0],
			PackageName: components[1],
		})
	}

	return composer.Update(conf, packages)
}

func serve(c *cli.Context) error {
	conf, err := getConfig(c)
	if err != nil {
		return err
	}

	// Do an initial generation of the repository
	if !c.Bool("no-generate") {
		log.Println("Generating initial repository")

		err = composer.Generate(conf)
		if err != nil {
			return err
		}
	}

	// Handle incoming requests and update packages as requested
	http.HandleFunc(c.String("listen-path"), func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("input") == "" || r.URL.Query().Get("package") == "" {
			log.Printf("Expected 'input' and 'package' params")

			w.WriteHeader(400)
			fmt.Fprintf(w, "Expected 'input' and 'package' params")
			return
		}

		log.Printf("Updating %s:%s", r.URL.Query().Get("input"), r.URL.Query().Get("package"))

		pkgInfo := &composer.PackageInfo{
			InputID:     r.URL.Query().Get("input"),
			PackageName: r.URL.Query().Get("package"),
		}

		// Check that the input exists
		if _, ok := conf.Inputs[pkgInfo.InputID]; !ok {
			log.Printf("Unknown input %q", pkgInfo.InputID)

			w.WriteHeader(404)
			fmt.Fprintf(w, "Unknown input %q", pkgInfo.InputID)
			return
		}

		err := composer.Update(conf, []*composer.PackageInfo{pkgInfo})
		if err != nil {
			log.Print(err)

			w.WriteHeader(500)
			fmt.Fprintf(w, err.Error())
			return
		}

		fmt.Fprintf(w, "OK")
	})

	log.Printf("Listening on %q", c.String("listen"))
	return http.ListenAndServe(c.String("listen"), nil)
}

func main() {
	app := cli.NewApp()

	app.Name = "repoctl"
	app.Usage = "Generate a composer repository"
	app.Version = "0.1.0"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "Location of the YAML configuration file.",
			Value: "config.yml",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "generate",
			Aliases: []string{"g"},
			Usage:   "Generate the composer.",
			Action:  generate,
		},
		{
			Name:    "update",
			Aliases: []string{"u"},
			Usage:   "Updates a specific package in the composer.",
			Action:  update,
		},
		{
			Name:    "serve",
			Aliases: []string{"s"},
			Usage:   "Generates the repository then listens for HTTP requests to update packages.",
			Action:  serve,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "listen",
					Value: ":8080",
				},
				cli.StringFlag{
					Name:  "listen-path",
					Value: "/",
				},
				cli.BoolFlag{
					Name:  "no-generate",
					Usage: "Don't generate the entire repository before listening",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
