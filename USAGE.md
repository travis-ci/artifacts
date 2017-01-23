Travis CI Artifacts Uploader Usage
==================================

:boom: **Warning: file generated from [USAGE.in.md](./USAGE.in.md) ** :boom:


## global options


### NAME
artifacts - manage your artifacts!

### USAGE
`artifacts [global options] command [command options] [arguments...]`

### COMMANDS
* `upload, u`  upload some artifacts!
* `help, h`  Shows a list of commands or help for one command

### GLOBAL OPTIONS
* `--log-format, -f`     log output format (text, json, or multiline) [`$ARTIFACTS_LOG_FORMAT`]
* `--log-level, -l`     set log level (debug, info, warning, error, fatal, panic) [`$ARTIFACTS_LOG_LEVEL`]
* `--debug, -D`        set log level to debug [`$ARTIFACTS_DEBUG`]
* `--quiet, -q`        set log level to panic [`$ARTIFACTS_QUIET`]
* `--help, -h`        show help
* `--version, -v`    print the version

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
environment variable, which should be :-delimited.
Paths may be either files or directories.  Any path provided will be walked for
all child entries.  Each entry will have its mime type detected based first on
the file extension, then by sniffing up to the first 512 bytes via the net/http
function "DetectContentType".

### OPTIONS
* `--key, -k`             upload credentials key *REQUIRED* (default "") [`$ARTIFACTS_KEY`]
* `--bucket, -b`         destination bucket *REQUIRED* (default "") [`$ARTIFACTS_BUCKET`]
* `--cache-control`         artifact cache-control header value (default "private") [`$ARTIFACTS_CACHE_CONTROL`]
* `--permissions`         artifact access permissions (default "private") [`$ARTIFACTS_PERMISSIONS`]
* `--secret, -s`         upload credentials secret *REQUIRED* (default "") [`$ARTIFACTS_SECRET`]
* `--s`3-region             region used when storing to S3 (default "us-east-1") [`$ARTIFACTS_REGION`]
* `--repo-slug, -r`         repo owner/name slug (default "") [`$ARTIFACTS_REPO_SLUG`]
* `--build-number`         build number (default "") [`$ARTIFACTS_BUILD_NUMBER`]
* `--build-id`             build id (default "") [`$ARTIFACTS_BUILD_ID`]
* `--job-number`         job number (default "") [`$ARTIFACTS_JOB_NUMBER`]
* `--job-id`             job id (default "") [`$ARTIFACTS_JOB_ID`]
* `--concurrency`         upload worker concurrency (default "5") [`$ARTIFACTS_CONCURRENCY`]
* `--max-size`             max combined size of uploaded artifacts (default "1048576000") [`$ARTIFACTS_MAX_SIZE`]
* `--upload-provider, -p`     artifact upload provider (artifacts, s3, null) (default "s3") [`$ARTIFACTS_UPLOAD_PROVIDER`]
* `--retries`             number of upload retries per artifact (default "2") [`$ARTIFACTS_RETRIES`]
* `--target-paths, -t`         artifact target paths (':'-delimited) (default "[:]") [`$ARTIFACTS_TARGET_PATHS`]
* `--working-dir`         working directory (default ".") [`$ARTIFACTS_WORKING_DIR`]
* `--save-host, -H`         artifact save host (default "") [`$ARTIFACTS_SAVE_HOST`]
* `--auth-token, -T`         artifact save auth token (default "") [`$ARTIFACTS_AUTH_TOKEN`]

<!-- 4uyjkwjm99WSbPyXt0/S2iltnVf8vbJqO0lgyCKxkL8= -->
