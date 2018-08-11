package azure

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"path"
	"strings"

	"github.com/Azure/azure-storage-blob-go/2018-03-28/azblob"
)

type AzureOutput struct {
	AccountName string
	Credentials *azblob.SharedKeyCredential
	Container   string
}

func (ao *AzureOutput) getContainerURL() (*azblob.ContainerURL, error) {
	pipeline := azblob.NewPipeline(ao.Credentials, azblob.PipelineOptions{})
	u, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", ao.AccountName))
	if err != nil {
		return nil, err
	}

	serviceURL := azblob.NewServiceURL(*u, pipeline)
	containerURL := serviceURL.NewContainerURL(ao.Container)
	return &containerURL, nil
}

func (ao *AzureOutput) Init(conf map[string]interface{}) error {
	var accountName, accountKey string
	if accountNameInt, ok := conf["accountName"]; ok {
		if accountName, ok = accountNameInt.(string); !ok {
			return errors.New("Expected Azure account name as a string")
		}
	} else {
		return errors.New("Expected Azure account name")
	}

	if accountKeyInt, ok := conf["accountKey"]; ok {
		if accountKey, ok = accountKeyInt.(string); !ok {
			return errors.New("Expected Azure account key as a string")
		}
	} else {
		return errors.New("Expected Azure account key")
	}

	ao.AccountName = accountName
	ao.Credentials = azblob.NewSharedKeyCredential(accountName, accountKey)

	if containerInt, ok := conf["container"]; ok {
		if container, ok := containerInt.(string); ok {
			ao.Container = container
		} else {
			return errors.New("Expected Azure container as a string")
		}
	} else {
		return errors.New("Expected Azure container")
	}

	return nil
}

func (ao *AzureOutput) GetBasePath() string {
	return fmt.Sprintf("/%s", ao.Container)
}

func (ao *AzureOutput) Get(name string) ([]byte, error) {
	log.Printf("Loading %q", name)

	containerURL, err := ao.getContainerURL()
	if err != nil {
		return nil, err
	}

	blobURL := containerURL.NewBlockBlobURL(name)
	get, err := blobURL.Download(context.Background(), 0, 0, azblob.BlobAccessConditions{}, false)
	if err != nil {
		return nil, err
	}

	body := get.Body(azblob.RetryReaderOptions{})
	defer body.Close()

	return ioutil.ReadAll(body)
}

func (ao *AzureOutput) Write(name string, data []byte) error {
	log.Printf("Writing %q", name)

	containerURL, err := ao.getContainerURL()
	if err != nil {
		return err
	}

	fPath := path.Join(strings.Split(name, "/")...)
	blobURL := containerURL.NewBlockBlobURL(fPath)

	_, err = blobURL.Upload(context.Background(), bytes.NewReader(data), azblob.BlobHTTPHeaders{ContentType: "application/json"}, azblob.Metadata{}, azblob.BlobAccessConditions{})
	if err != nil {
		return err
	}

	return nil
}
