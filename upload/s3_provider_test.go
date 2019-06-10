package upload

import (
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
	"github.com/travis-ci/artifacts/artifact"
)

var testOpts = &Options{
	AccessKey:    "fakeAccessKey",
	BucketName:   "someBucket",
	CacheControl: defaultPublicCacheControl,
	Perm:         s3.BucketCannedACLPrivate,
	S3Region:     "s3-region-1",
	SecretKey:    "verySecretString",
}

type testS3Manager struct {
	c client.ConfigProvider
	u *testS3Uploader
	o sync.Once
}

func (m *testS3Manager) NewUploader(c client.ConfigProvider, options ...func(*s3manager.Uploader)) s3Uploader {
	m.o.Do(func() {
		m.c = c
		m.u = new(testS3Uploader)
	})
	return m.u
}

type testS3Uploader struct {
	inputs []*s3manager.UploadInput
	m      sync.Mutex
}

func (u *testS3Uploader) Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	u.m.Lock()
	defer u.m.Unlock()
	u.inputs = append(u.inputs, input)
	return &s3manager.UploadOutput{}, nil
}

func TestS3Provider_Name(t *testing.T) {
	name := "s3"

	p := newS3Provider(testOpts, logrus.New())
	if p.Name() != name {
		t.Fatalf("provider name incorrect, want %q have %q", name, p.Name())
	}
}

func TestS3Provider_Upload(t *testing.T) {

	tm := new(testS3Manager)

	p := newS3Provider(testOpts, logrus.New())
	p.s3manager = tm

	in := make(chan *artifact.Artifact)
	out := make(chan *artifact.Artifact)
	done := make(chan bool)

	go p.Upload("test-0", testOpts, in, out, done)

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
			t.Logf("---> Fed artifact: %#v\n", a)
		}
		close(in)
	}()

	var isDone bool
	var accum []*artifact.Artifact
	for isDone == false {
		select {
		case <-time.After(5 * time.Second):
			t.Fatalf("upload took too long")
		case a := <-out:
			accum = append(accum, a)
		case <-done:
			if len(accum) == 0 {
				t.Fatalf("nothing uploaded")
			}
			isDone = true
		}
	}

	// validate that we received two upload results
	if len(accum) != 2 {
		t.Fatalf("wanted 2 artifacts, have %d", len(accum))
	}
	// validate that upload has been called twice
	if len(tm.u.inputs) != 2 {
		t.Fatalf("wanted uploader to have been called 2 times, have %d", len(tm.u.inputs))
	}
	// validate all upload methods have been called correctly
	for i := 0; i < 2; i++ {
		if aws.StringValue(tm.u.inputs[i].Bucket) != testOpts.BucketName {
			t.Errorf("Bucket for upload %d does not match, want %q, have %q", i, testOpts.BucketName, aws.StringValue(tm.u.inputs[i].Bucket))
		}
		if aws.StringValue(tm.u.inputs[i].CacheControl) != testOpts.CacheControl {
			t.Errorf("CacheControl for upload %d does not match, want %q, have %q", i, testOpts.CacheControl, aws.StringValue(tm.u.inputs[i].CacheControl))
		}
		if aws.StringValue(tm.u.inputs[i].ACL) != testOpts.Perm {
			t.Errorf("ACL for upload %d does not match, want %q, have %q", i, testOpts.Perm, aws.StringValue(tm.u.inputs[i].ACL))
		}
		if aws.StringValue(tm.u.inputs[i].Key) != accum[i].FullDest() {
			t.Errorf("Key for upload %d does not match, want %q, have %q", i, accum[i].FullDest(), aws.StringValue(tm.u.inputs[i].Key))
		}
		if aws.StringValue(tm.u.inputs[i].ContentType) != accum[i].ContentType() {
			t.Errorf("ContentType for upload %d does not match, want %q, have %q", i, accum[i].ContentType(), aws.StringValue(tm.u.inputs[i].ContentType))
		}

		body, err := ioutil.ReadAll(tm.u.inputs[i].Body)
		if err != nil {
			t.Fatal(err)
		}
		wantLen := 5120
		if len(body) != wantLen {
			t.Errorf("body size for upload %d does not match, want %q, have %q", i, wantLen, len(body))
		}
	}

	// validate we passed correct credentials and region
	creds, _ := tm.c.ClientConfig("S3").Config.Credentials.Get()
	if creds.AccessKeyID != testOpts.AccessKey {
		t.Errorf("AccessKey does not match, want %q, have %q", testOpts.AccessKey, creds.AccessKeyID)
	}
	if creds.SecretAccessKey != testOpts.SecretKey {
		t.Errorf("SecretKey does not match, want %q, have %q", testOpts.SecretKey, creds.SecretAccessKey)
	}
	if aws.StringValue(tm.c.ClientConfig("S3").Config.Region) != testOpts.S3Region {
		t.Errorf("S3Region does not match, want %q, have %q", testOpts.S3Region, aws.StringValue(tm.c.ClientConfig("S3").Config.Region))
	}
}
