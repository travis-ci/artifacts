package upload

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/travis-ci/artifacts/artifact"
)

type nullPutter struct {
	Putted []*artifact.Artifact
}

func (np *nullPutter) PutArtifact(a *artifact.Artifact) error {
	if np.Putted == nil {
		np.Putted = []*artifact.Artifact{}
	}
	np.Putted = append(np.Putted, a)
	return nil
}

func TestArtifactsProviderDefaults(t *testing.T) {
	opts := NewOptions()
	log := getPanicLogger()
	ap := newArtifactsProvider(opts, log)

	if ap.RetryInterval != defaultProviderRetryInterval {
		t.Fatalf("RetryInterval %v != %v", ap.RetryInterval, defaultProviderRetryInterval)
	}

	if ap.opts != opts {
		t.Fatalf("opts %v != %v", ap.opts, opts)
	}

	if ap.log != log {
		t.Fatalf("log %v != %v", ap.log, log)
	}

	if ap.Name() != "artifacts" {
		t.Fatalf("Name %v != artifacts", ap.Name())
	}
}

func TestArtifactsUpload(t *testing.T) {
	opts := NewOptions()
	log := getPanicLogger()
	ap := newArtifactsProvider(opts, log)
	ap.overrideClient = &nullPutter{}

	in := make(chan *artifact.Artifact)
	out := make(chan *artifact.Artifact)
	done := make(chan bool)

	go ap.Upload("test-0", opts, in, out, done)

	go func() {
		for _, p := range testArtifactPaths {
			if !p.Valid {
				continue
			}

			a := artifact.New("bucket", p.Path, "linux/foo", &artifact.Options{
				Perm:     s3.BucketCannedACLPublicRead,
				RepoSlug: "owner/foo",
			})

			in <- a
			fmt.Printf("---> Fed artifact: %#v\n", a)
		}
		close(in)
	}()

	accum := []*artifact.Artifact{}
	for {
		select {
		case <-time.After(5 * time.Second):
			t.Fatalf("took too long oh derp")
		case a := <-out:
			accum = append(accum, a)
		case <-done:
			if len(accum) == 0 {
				t.Fatalf("nothing uploaded")
			}
			return
		}
	}
}
