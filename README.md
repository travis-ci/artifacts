artifacts
=========

Travis CI Artifact thingy


### NAME
artifacts - manage your artifacts!

### USAGE
`artifacts [global options] command [command options] [arguments...]`

### COMMANDS
* `upload, u`  upload some artifacts!
* `help, h`  Shows a list of commands or help for one command

### GLOBAL OPTIONS
* `--log-format, -f`     log output format (text, json, or multiline)
* `--debug, -D`        set log level to debug
* `--version, -v`    print the version
* `--help, -h`        show help

## upload

The upload commmand may be used to upload arbitrary files to an artifact
repository.  The only such artifact repository currently supported is
S3.  All of the required arguments may be provided as command line
arguments or environment variables.


### NAME
upload - upload some artifacts!

### USAGE
`command upload [command options] [arguments...]`

### DESCRIPTION
Upload a set of local paths to an artifact repository.  The paths may be
provided as either positional command-line arguments or as the `$ARTIFACTS_PATHS`
environmental variable, which should be ;-delimited.
Paths may be either files or directories.  Any path provided will be walked for
all child entries.  Each entry will have its mime type detected based first on
the file extension, then by sniffing up to the first 512 bytes via the net/http
function "DetectContentType".

### OPTIONS
* `--key, -k`         upload credentials key (`$ARTIFACTS_KEY`) *REQUIRED*
* `--secret, -s`     upload credentials secret (`$ARTIFACTS_SECRET`) *REQUIRED*
* `--bucket, -b`     destination bucket (`$ARTIFACTS_BUCKET`) *REQUIRED*
* `--cache-control`     artifact cache-control header value (`$ARTIFACTS_CACHE_CONTROL`) (default "private")
* `--concurrency`     upload worker concurrency (`$ARTIFACTS_CONCURRENCY`) (default 5)
* `--max-size`         max combined size of uploaded artifacts (`$ARTIFACTS_MAX_SIZE`) (default 1.0GB)
* `--permissions`     artifact access permissions (`$ARTIFACTS_PERMISSIONS`) (default "private")
* `--retries`         number of upload retries per artifact (`$ARTIFACT_RETRIES`) (default 2)
* `--target-paths, -t`     artifact target paths (';'-delimited) (`$ARTIFACTS_TARGET_PATHS`) (default []string{"artifacts"})
* `--working-dir`     working directory (`$TRAVIS_BUILD_DIR`) (default `$PWD`)

### S3 ENVIRONMENT COMPATIBILITY

In addition to the environmental variables listed above for defining the
access key, secret, and bucket, some additional variables will also work.

#### environmental variables accepted for "key"

0. `ARTIFACTS_KEY`
0. `ARTIFACTS_AWS_ACCESS_KEY`
0. `AWS_ACCESS_KEY_ID`
0. `AWS_ACCESS_KEY`

#### environmental variables accepted for "secret"

0. `ARTIFACTS_SECRET`
0. `ARTIFACTS_AWS_SECRET_KEY`
0. `AWS_SECRET_ACCESS_KEY`
0. `AWS_SECRET_KEY`

#### environmental variables accepted for "bucket"

0. `ARTIFACTS_BUCKET`
0. `ARTIFACTS_S3_BUCKET`


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

The same operation using environmental variables would look like this:

``` bash
export ARTIFACTS_KEY="AKIT339AFIY655O3Q9DZ"
export ARTIFACTS_SECRET="48TmqyraUyJ7Efpegi6Lfd10yUskAMB0G2TtRCX1"
export ARTIFACTS_BUCKET="my-fancy-bucket"
export ARTIFACTS_PATHS="log/;coverage/"

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

The same operation using environmental variables would look like this:

``` bash
export ARTIFACTS_KEY="AKIT339AFIY655O3Q9DZ"
export ARTIFACTS_SECRET="48TmqyraUyJ7Efpegi6Lfd10yUskAMB0G2TtRCX1"
export ARTIFACTS_BUCKET="my-fancy-bucket"
export ARTIFACTS_PATHS="$(git ls-files -o | tr "\n" ";")"

artifacts upload
```

#### Example: multiple target paths

Specifying one or more custom target path will override the default of
`artifacts/$TRAVIS_BUILD_NUMBER/$TRAVIS_JOB_NUMBER`.  Multiple target paths
must be specified in ';'-delimited strings:

``` bash
artifacts upload \
  --key AKIT339AFIY655O3Q9DZ \
  --secret 48TmqyraUyJ7Efpegi6Lfd10yUskAMB0G2TtRCX1 \
  --bucket my-fancy-bucket \
  --target-paths "artifacts/$TRAVIS_REPO_SLUG/$TRAVIS_BUILD_NUMBER/$TRAVIS_JOB_NUMBER;artifacts/$TRAVIS_REPO_SLUG/$TRAVIS_COMMIT" \
  $(git ls-files -o)
```

The same operation using environmental variables would look like this:

``` bash
export ARTIFACTS_TARGET_PATHS="artifacts/$TRAVIS_REPO_SLUG/$TRAVIS_BUILD_NUMBER/$TRAVIS_JOB_NUMBER;artifacts/$TRAVIS_REPO_SLUG/$TRAVIS_COMMIT"
export ARTIFACTS_KEY="AKIT339AFIY655O3Q9DZ"
export ARTIFACTS_SECRET="48TmqyraUyJ7Efpegi6Lfd10yUskAMB0G2TtRCX1"
export ARTIFACTS_BUCKET="my-fancy-bucket"
export ARTIFACTS_PATHS="$(git ls-files -o | tr "\n" ";")"

artifacts upload
```
