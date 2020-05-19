Travis CI Artifacts Uploader
============================

A smart little Go app to help aid in uploading build artifacts.

## Installation

There are pre-built binaries of the latest stable build for 64-bit
Linux, OSX, and Windows available here via the following links.  Please
note that the tests run on 64-bit Linux.

* [Linux/amd64](https://s3.amazonaws.com/travis-ci-gmbh/artifacts/stable/build/linux/amd64/artifacts)
* [Linux/ppc64le](https://s3.amazonaws.com/travis-ci-gmbh/artifacts/stable/build/linux/ppc64le/artifacts)
* [OSX/amd64](https://s3.amazonaws.com/travis-ci-gmbh/artifacts/stable/build/darwin/amd64/artifacts)
* [Windows/amd64](https://s3.amazonaws.com/travis-ci-gmbh/artifacts/stable/build/windows/amd64/artifacts.exe)
* [FreeBSD/amd64](https://s3.amazonaws.com/travis-ci-gmbh/artifacts/stable/build/freebsd/amd64/artifacts)
* [SHA-256 checksums](https://s3.amazonaws.com/travis-ci-gmbh/artifacts/stable/SHA256SUMS)

There is also an [install script](./install) for Linux and OSX that may
be used like so:

``` bash
curl -sL https://raw.githubusercontent.com/travis-ci/artifacts/master/install | bash
```

## Usage

Once the binary is in your `$PATH` and has been made executable, you
will have access to the help system via `artifacts help` (also available
[here](./USAGE.md)).

### S3 ENVIRONMENT COMPATIBILITY

In addition to the environment variables listed above for defining the
access key, secret, and bucket, some additional variables will also work.

#### environment variables accepted for "key"

0. `ARTIFACTS_KEY`
0. `ARTIFACTS_AWS_ACCESS_KEY`
0. `AWS_ACCESS_KEY_ID`
0. `AWS_ACCESS_KEY`

#### environment variables accepted for "secret"

0. `ARTIFACTS_SECRET`
0. `ARTIFACTS_AWS_SECRET_KEY`
0. `AWS_SECRET_ACCESS_KEY`
0. `AWS_SECRET_KEY`

#### environment variables accepted for "bucket"

0. `ARTIFACTS_BUCKET`
0. `ARTIFACTS_S3_BUCKET`

#### environment variables accepted for "region"

0. `ARTIFACTS_REGION`
0. `ARTIFACTS_S3_REGION`


### EXAMPLES

#### Example: logs and coverage

In this case, the key and secret are passed as command line flags and
the `log/` and `coverage/` directories are passed as positional path
arguments:

``` bash
artifacts upload \
  --key AKIT339AFIY655O3Q9DZ \
  --secret 48TmqyraUyJ7Efpegi6Lfd10yUskAMB0G2TtRCX1 \
  --bucket my-fancy-bucket \
  log/ coverage/
```

The same operation using environment variables would look like this:

``` bash
export ARTIFACTS_KEY="AKIT339AFIY655O3Q9DZ"
export ARTIFACTS_SECRET="48TmqyraUyJ7Efpegi6Lfd10yUskAMB0G2TtRCX1"
export ARTIFACTS_BUCKET="my-fancy-bucket"
export ARTIFACTS_PATHS="log/:coverage/"

artifacts upload
```

#### Example: untracked files

In order to upload all of the untracked files (according to git), one
might do this:

``` bash
artifacts upload \
  --key AKIT339AFIY655O3Q9DZ \
  --secret 48TmqyraUyJ7Efpegi6Lfd10yUskAMB0G2TtRCX1 \
  --bucket my-fancy-bucket \
  $(git ls-files -o)
```

The same operation using environment variables would look like this:

``` bash
export ARTIFACTS_KEY="AKIT339AFIY655O3Q9DZ"
export ARTIFACTS_SECRET="48TmqyraUyJ7Efpegi6Lfd10yUskAMB0G2TtRCX1"
export ARTIFACTS_BUCKET="my-fancy-bucket"
export ARTIFACTS_PATHS="$(git ls-files -o | tr "\n" ":")"

artifacts upload
```

#### Example: multiple target paths

Specifying one or more custom target path will override the default of
`artifacts/$TRAVIS_BUILD_NUMBER/$TRAVIS_JOB_NUMBER`.  Multiple target paths
must be specified in ':'-delimited strings:

``` bash
artifacts upload \
  --key AKIT339AFIY655O3Q9DZ \
  --secret 48TmqyraUyJ7Efpegi6Lfd10yUskAMB0G2TtRCX1 \
  --bucket my-fancy-bucket \
  --target-paths "artifacts/$TRAVIS_REPO_SLUG/$TRAVIS_BUILD_NUMBER/$TRAVIS_JOB_NUMBER:artifacts/$TRAVIS_REPO_SLUG/$TRAVIS_COMMIT" \
  $(git ls-files -o)
```

The same operation using environment variables would look like this:

``` bash
export ARTIFACTS_TARGET_PATHS="artifacts/$TRAVIS_REPO_SLUG/$TRAVIS_BUILD_NUMBER/$TRAVIS_JOB_NUMBER:artifacts/$TRAVIS_REPO_SLUG/$TRAVIS_COMMIT"
export ARTIFACTS_KEY="AKIT339AFIY655O3Q9DZ"
export ARTIFACTS_SECRET="48TmqyraUyJ7Efpegi6Lfd10yUskAMB0G2TtRCX1"
export ARTIFACTS_BUCKET="my-fancy-bucket"
export ARTIFACTS_PATHS="$(git ls-files -o | tr "\n" ":")"

artifacts upload
```
