## GRM: A tool for fetching the latest release assets from public GitHub Repositories

### Installation

`go get -u github.com/kowala-tech/github-release-manager/cmd/grm`

### Usage

To get the latest tarball for a repository, run

`grm my-user/my-repo`

This will download the latest release to the disk to a file in the current directory named after the tag, if there is a release. A file called `.grm/my-user/my-repo` will be created on the disk with the latest tag name in it, and the `grm` command will return 0.

If there is an existing file called `.grm/my-user/my-repo` that contains the current name, then nothing will be downloaded and `grm` will return 1.

This can then be used in shell scripts and Makefiles to trigger actions when the script is run:

```
#!/bin/bash

echo "Making sure my-repo is up to date
grm -o my-repo-lastest.tar my-user/my-repo

if ! [ $? -eq 1 ]; then 
    echo "The repo has changed or been installed for the first time"
    tar -xvf my-repo-latest.tar /some/path
else 
    echo "The reapw as already up to date`
fi
```

### Full command syntax:

```
Syntax: grm [flags] <owner/repo>

Flags:

  -a string
    	Asset name to download from the release
  -asset string
    	Asset name to download from the release
  -o string
    	Path to write the downloaded asset
  -output string
    	Path to write the downloaded asset
```
