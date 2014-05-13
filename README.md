artifacts
=========

Travis CI Artifact thingy

## available commands and global options

```
NAME:
   artifacts - manage your artifacts!

USAGE:
   artifacts [global options] command [command options] [arguments...]

VERSION:
   v0.2.0-1-g51cad36-dirty

COMMANDS:
   upload, u	upload some artifacts!
   help, h	Shows a list of commands or help for one command
   
GLOBAL OPTIONS:
   --log-format, -f 'text'	log output format (text or json)
   --debug, -D			set log level to debug
   --version, -v		print the version
   --help, -h			show help
   
```

## upload

The upload commmand may be used to upload arbitrary files to an artifact
repository.  The only such artifact repository currently supported is
S3.  All of the required arguments may be provided as command line
arguments or environment variables.

```
NAME:
   upload - upload some artifacts!

USAGE:
   command upload [command options] [arguments...]

DESCRIPTION:
   

OPTIONS:
   --key, -k 		upload credentials key [ARTIFACTS_KEY]
   --secret, -s 	upload credentials secret [ARTIFACTS_SECRET]
   
```
