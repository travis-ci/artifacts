package upload

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

var (
	isDebug = os.Getenv("ARTIFACTS_DEBUG") != ""
)

func setUploaderEnv() {
	setenvs(map[string]string{
		"TRAVIS_BUILD_NUMBER":       "3",
		"TRAVIS_JOB_NUMBER":         "3.2",
		"ARTIFACTS_S3_BUCKET":       "foo",
		"ARTIFACTS_TARGET_PATHS":    "baz:artifacts/$TRAVIS_BUILD_NUMBER/$TRAVIS_JOB_NUMBER",
		"ARTIFACTS_PATHS":           "bin/:derp",
		"ARTIFACTS_UPLOAD_PROVIDER": "null",
	})
}

func getPanicLogger() *logrus.Logger {
	log := logrus.New()
	log.Level = logrus.PanicLevel
	if isDebug {
		log.Level = logrus.DebugLevel
	}
	return log
}

func getTestUploader() *uploader {
	setUploaderEnv()

	log := getPanicLogger()
	u := newUploader(NewOptions(), log)
	u.Provider = newNullProvider(nil, log)
	return u
}

func TestNewUploader(t *testing.T) {
	u := getTestUploader()
	if u == nil {
		t.Errorf("options are %v", u)
	}

	if u.Opts.BucketName != "foo" {
		t.Errorf("bucket name is %v", u.Opts.BucketName)
	}

	if len(u.Opts.TargetPaths) != 2 {
		t.Errorf("target paths length != 2: %v", len(u.Opts.TargetPaths))
	}

	if u.Opts.TargetPaths[0] != "baz" {
		t.Errorf("target paths[0] != baz: %v", u.Opts.TargetPaths)
	}

	if u.Opts.TargetPaths[1] != "artifacts/3/3.2" {
		t.Errorf("target paths[1] != artifacts/3/3.2: %v", u.Opts.TargetPaths)
	}

	if len(u.Paths.All()) != 2 {
		t.Errorf("all paths length != 2: %v", len(u.Paths.All()))
	}
}

var testOptsProviderCases = map[string]string{
	"artifacts": "artifacts",
	"s3":        "s3",
	"null":      "null",
	"foo":       "s3",
	"":          "s3",
}

func TestNewUploaderProviderOptions(t *testing.T) {
	opts := NewOptions()
	for opt, name := range testOptsProviderCases {
		opts.Provider = opt
		u := newUploader(opts, getPanicLogger())
		if u.Provider.Name() != name {
			t.Fatalf("new uploader does not have %s provider: %q != %q",
				name, u.Provider.Name(), name)
		}
	}
}

func TestNewUploaderUnsetCacheControlOption(t *testing.T) {
	opts := NewOptions()
	opts.CacheControl = ""
	u := newUploader(opts, getPanicLogger())
	if u.Opts.CacheControl != defaultPublicCacheControl {
		t.Fatalf("new uploader cache control option not defaulted")
	}
}

func TestUpload(t *testing.T) {
	setUploaderEnv()
	err := Upload(NewOptions(), getPanicLogger())
	if err != nil {
		t.Errorf("go boom: %v", err)
	}
}

func TestUploaderUpload(t *testing.T) {
	u := getTestUploader()
	if u == nil {
		t.Errorf("options are %v", u)
	}

	err := u.Upload()
	if err != nil {
		t.Errorf("failed to not really upload: %v", err)
	}
}
