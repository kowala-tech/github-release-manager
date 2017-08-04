## GRM: A tool for fetching the latest release assets from public GitHub Repositories

### Installation

`go get -u github.com/kowala-tech/github-release-manager/cmd/grm`

### Usage

To get the latest tarball for a repository, run

`grm my-user/my-repo`

This will download the latest release to the disk to a file in the current directory named after the tag, if there is a release. A file called `.grm/my-user/my-repo` will be created on the disk with the latest tag name in it, and the `grm` command will return 0.

If there is an existing file called `.grm/my-user/my-repo` that contains the current name, then nothing will be downloaded and `grm` will return 1.

### Full command syntax:

```
Syntax: ../core/bin/grm [flags] <owner/repo>

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
