package upload

import (
	"fmt"
	"testing"
	"time"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"github.com/mitchellh/goamz/s3/s3test"
	"github.com/travis-ci/artifacts/artifact"
)

var (
	s3srv = &localS3Server{
		config: &s3test.Config{
			Send409Conflict: true,
		},
	}
	testS3 *s3.S3
)

func init() {
	s3srv.SetUp()
	testS3 = s3.New(s3srv.Auth, s3srv.Region)
	err = testS3.Bucket("bucket").PutBucket(s3.ACL("public-read"))
	if err != nil {
		panic(err)
	}
}

type localS3Server struct {
	Auth   aws.Auth
	Region aws.Region
	srv    *s3test.Server
	config *s3test.Config
}

func (s *localS3Server) SetUp() {
	if s.srv != nil {
		return
	}

	srv, err := s3test.NewServer(s.config)
	if err != nil {
		panic(err)
	}

	s.srv = srv
	s.Region = aws.Region{
		Name:                 "faux-region-9000",
		S3Endpoint:           srv.URL(),
		S3LocationConstraint: true,
	}
}

func TestNewS3Provider(t *testing.T) {
	s3p := newS3Provider(NewOptions(), getPanicLogger())

	if s3p.RetryInterval != defaultProviderRetryInterval {
		t.Fatalf("RetryInterval %v != %v", s3p.RetryInterval, defaultProviderRetryInterval)
	}

	if s3p.Name() != "s3" {
		t.Fatalf("Name %v != s3", s3p.Name())
	}
}

func TestS3ProviderUpload(t *testing.T) {
	opts := NewOptions()
	s3p := newS3Provider(opts, getPanicLogger())
	s3p.overrideConn = testS3
	s3p.overrideAuth = aws.Auth{
		AccessKey: "whatever",
		SecretKey: "whatever",
		Token:     "whatever",
	}

	in := make(chan *artifact.Artifact)
	out := make(chan *artifact.Artifact)
	done := make(chan bool)

	go s3p.Upload("test-0", opts, in, out, done)

	go func() {
		for _, p := range testArtifactPaths {
			if !p.Valid {
				continue
			}

			a := artifact.New("bucket", p.Path, "linux/foo", &artifact.Options{
				Perm:     s3.PublicRead,
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

func TestS3ProviderRegionOption(t *testing.T) {
	opts := NewOptions()

	for input, output := range map[string]string{
		"us-west-2":  "us-west-2",
		"bogus-9000": "us-east-1",
	} {
		opts.S3Region = input
		s3p := newS3Provider(opts, getPanicLogger())

		region := s3p.getRegion()
		if region.Name != output {
			t.Fatalf("region %v != %v", region.Name, output)
		}
	}
}
