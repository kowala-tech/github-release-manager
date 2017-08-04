package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/github"
)

type fetcher interface {
	Fetch() error
}

type writer func(format string, a ...interface{})

type application struct {
	requestedAsset  string
	requestedOutput string
	ownerAndRepo    string
	writer          writer
	client          *github.Client
	context         context.Context
	workingDir      string
}

func newApplication(ownerAndRepo, requestedAsset, requestedOutput string) fetcher {

	printlnWrapper := func(format string, a ...interface{}) {
		fmt.Printf(format, a...)
	}

	return &application{
		requestedAsset:  requestedAsset,
		requestedOutput: requestedOutput,
		ownerAndRepo:    ownerAndRepo,
		writer:          printlnWrapper,
		client:          github.NewClient(nil),
		context:         context.Background(),
		workingDir:      ".grm",
	}
}

func (a *application) Fetch() error {

	owner, repo, err := a.findOwnerAndRepo(a.ownerAndRepo)

	if err != nil {
		return err
	}

	release, _, err := a.client.Repositories.GetLatestRelease(a.context, owner, repo)

	tag := release.GetTagName()

	if err := a.assertTag(owner, repo, tag); err != nil {
		return err
	}

	downloadURL, err := a.findDownloadURL(release)

	if err != nil {
		return err
	}

	output := a.findOutputPath(downloadURL)

	a.writer("Downloading '%s' to %s...", downloadURL, output)

	if err := a.downloadFile(output, downloadURL); err != nil {
		return err
	}

	if err := a.writeTag(owner, repo, tag); err != nil {
		return err
	}

	a.writer("Done.\n")

	return nil
}

func (a *application) assertTag(owner, repo, tag string) error {

	file := filepath.Join(a.workingDir, owner, repo)

	// If the file exists and has content, check it
	if content, err := ioutil.ReadFile(file); err == nil {

		if string(content) == tag {
			return fmt.Errorf("Already up to date, nothing to download")
		}
	}

	return nil
}

func (a *application) writeTag(owner, repo, tag string) error {

	file := filepath.Join(a.workingDir, owner, repo)

	if err := os.MkdirAll(filepath.Dir(file), os.ModePerm); err != nil {
		return err
	}

	out, err := os.Create(file)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.WriteString(out, tag); err != nil {
		return fmt.Errorf("Failed to write tag to %s: %s", file, err)
	}

	return nil
}

func (a *application) findDownloadURL(release *github.RepositoryRelease) (string, error) {

	if a.requestedAsset == "" {
		return release.GetTarballURL(), nil
	}

	asset, err := a.findReleaseAsset(a.requestedAsset, release.Assets)

	if err != nil {
		return "", err
	}

	return asset.GetBrowserDownloadURL(), nil
}

func (a *application) findOutputPath(downloadURL string) (output string) {

	if a.requestedOutput == "" {
		parts := strings.Split(downloadURL, "/")
		output = parts[len(parts)-1]
	} else {
		output = a.requestedOutput
	}

	return
}

func (a *application) findOwnerAndRepo(input string) (owner, repo string, err error) {

	parts := strings.Split(input, "/")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("Expected owner/repo format")
	}

	return parts[0], parts[1], nil
}

func (a *application) findReleaseAsset(name string, assets []github.ReleaseAsset) (github.ReleaseAsset, error) {

	for _, asset := range assets {
		if name == asset.GetName() {
			return asset, nil
		}
	}

	return github.ReleaseAsset{}, fmt.Errorf("Can't find '%s' in release assets", name)
}

func (a *application) downloadFile(filepath string, url string) (err error) {

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)

	return err
}
