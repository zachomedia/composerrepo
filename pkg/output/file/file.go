package file

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/zachomedia/composerrepo/pkg/composer/repository"
)

type FileOutput struct {
	Out      string
	BasePath string
}

func (fo *FileOutput) Init(conf map[string]interface{}) error {
	if dir, ok := conf["dir"]; ok {
		fo.Out = dir.(string)
	}

	if basePath, ok := conf["basePath"]; ok {
		fo.BasePath = basePath.(string)
	}

	return nil
}

func (fo *FileOutput) ensureFolder(f string) error {
	_, err := os.Stat(f)
	if os.IsNotExist(err) {
		log.Printf("Directory doesn't exist, creating %q", f)

		merr := os.MkdirAll(f, os.ModePerm)

		if merr != nil {
			return merr
		}
	} else if err != nil {
		return err
	}

	return nil
}

func (fo *FileOutput) generateContentsAndHash(obj interface{}) ([]byte, string, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, "", err
	}

	return b, fmt.Sprintf("%x", sha256.Sum256(b)), nil
}

func (fo *FileOutput) GetBasePath() string {
	return fo.BasePath
}

func (fo *FileOutput) GetRepository() (*repository.Repository, error) {
	return fo.Get("packages.json")
}

func (fo *FileOutput) Get(name string) (*repository.Repository, error) {
	fPath := path.Join(fo.Out, path.Join(strings.Split(name, "/")...))
	log.Printf("Reading package %q", fPath)

	// Write packages.json
	f, err := os.Open(fPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)

	repo := repository.Repository{}
	err = decoder.Decode(&repo)
	if err != nil {
		return nil, err
	}

	return &repo, nil
}

func (fo *FileOutput) WriteRepository(repo *repository.Repository) error {
	_, err := fo.Write("packages.json", repo)
	return err
}

func (fo *FileOutput) Write(name string, repo *repository.Repository) (string, error) {
	components := strings.Split(name, "/")

	// Ensure the directory structure is correct.
	if err := fo.ensureFolder(path.Join(fo.Out, path.Join(components[:len(components)-1]...))); err != nil {
		return "", err
	}

	// To JSON + hash
	contents, hash, err := fo.generateContentsAndHash(repo)
	if err != nil {
		return "", err
	}

	fPath := path.Join(fo.Out, path.Join(strings.Split(strings.Replace(name, "%hash%", hash, -1), "/")...))
	log.Printf("Writing package %q", fPath)

	// Write packages.json
	f, err := os.Create(fPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = f.Write(contents)
	if err != nil {
		return "", err
	}

	return hash, nil
}
