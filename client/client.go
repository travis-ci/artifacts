package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/meatballhat/artifacts/artifact"
)

var (
	errFailedPut = fmt.Errorf("failed to put artifact to artifacts service")
)

// Client does stuff with the server
type Client struct {
	SaveURL       string
	Token         string
	RetryInterval time.Duration

	log *logrus.Logger
}

// New creates a new *Client
func New(url, token string, log *logrus.Logger) *Client {
	return &Client{
		SaveURL:       url,
		Token:         token,
		RetryInterval: 3 * time.Second,

		log: log,
	}
}

// PutArtifact puts ... an ... artifact
func (c *Client) PutArtifact(a *artifact.Artifact) error {
	reader, err := a.Reader()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", c.SaveURL, reader)
	if err != nil {
		return err
	}

	size, err := a.Size()
	if err != nil {
		return err
	}

	req.Header.Set("Artifacts-Repo-Slug", a.RepoSlug)
	req.Header.Set("Artifacts-Source", a.Path.From)
	req.Header.Set("Artifacts-Destination", a.FullDestination())
	req.Header.Set("Artifacts-Job-Number", a.JobNumber)
	req.Header.Set("Artifacts-Size", fmt.Sprintf("%d", size))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errFailedPut
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	c.log.WithFields(logrus.Fields{
		"artifact": a,
		"response": string(body),
	}).Debug("successfully uploaded artifact")

	return nil
}
