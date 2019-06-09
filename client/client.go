package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/travis-ci/artifacts/artifact"
)

var (
	errFailedPut = fmt.Errorf("failed to put artifact to artifacts service")

	defaultRetryInterval = 3 * time.Second
)

// Client does stuff with the server
type Client struct {
	SaveHost      string
	Token         string
	RetryInterval time.Duration

	log *logrus.Logger
}

// New creates a new *Client
func New(host, token string, log *logrus.Logger) *Client {
	return &Client{
		SaveHost:      host,
		Token:         token,
		RetryInterval: defaultRetryInterval,

		log: log,
	}
}

// PutArtifact puts ... an ... artifact
func (c *Client) PutArtifact(a *artifact.Artifact) error {
	reader, err := a.Reader()
	if err != nil {
		return err
	}

	// e.g. hostname.example.org/owner/repo/jobs/123456/path/to/artifact
	fullURL := fmt.Sprintf("%s/%s",
		c.SaveHost,
		path.Join(a.RepoSlug, "jobs", a.JobID, a.Dest))

	c.log.WithFields(logrus.Fields{
		"url":    fullURL,
		"source": a.Source,
	}).Debug("putting artifact to url")

	req, err := http.NewRequest("PUT", fullURL, reader)
	if err != nil {
		return err
	}

	size, err := a.Size()
	if err != nil {
		return err
	}

	req.Header.Set("Artifacts-Repo-Slug", a.RepoSlug)
	req.Header.Set("Artifacts-Source", a.Source)
	req.Header.Set("Artifacts-Dest", a.FullDest())
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
