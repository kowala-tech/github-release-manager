package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-github/github"
	"github.com/stretchr/testify/require"
)

const (
	repoOwner    = "owner"
	repoName     = "repo"
	assetName    = "asset.file"
	outputFile   = "output.file"
	assetContent = "ASSET"
)

func newMockApplication(t *testing.T) (*application, <-chan string, func()) {

	outputChan := make(chan string, 10)

	mockWrapper := func(format string, a ...interface{}) {
		outputChan <- fmt.Sprintf(format, a...)
	}

	tmpdir, err := ioutil.TempDir("", "grm")
	require.Nil(t, err)

	app := &application{
		requestedAsset:  assetName,
		requestedOutput: filepath.Join(tmpdir, outputFile),
		ownerAndRepo:    fmt.Sprintf("%s/%s", repoOwner, repoName),
		writer:          mockWrapper,
		client:          github.NewClient(nil),
		context:         context.Background(),
		workingDir:      tmpdir,
	}

	return app,
		outputChan,
		func() {
			close(outputChan)
			require.Nil(t, os.RemoveAll(tmpdir))
		}
}

func Test_AnApplicationCanFetchAReleaseAsset(t *testing.T) {

	app, _, cleanup := newMockApplication(t)
	defer cleanup()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	url, _ := url.Parse(server.URL)
	app.client.BaseURL = url

	mux.HandleFunc("/repos/owner/repo/releases/latest", func(w http.ResponseWriter, r *http.Request) {

		names := []string{"asset.file"}
		tarballURL := "http://github.com/release/tarball.tar"
		assetURL := "http://github.com/release/asset.file"

		release := &github.RepositoryRelease{
			TarballURL: &tarballURL,
			Assets: []github.ReleaseAsset{
				github.ReleaseAsset{
					Name:               &names[0],
					BrowserDownloadURL: &assetURL,
				},
			},
		}

		jsn, err := json.Marshal(release)
		require.Nil(t, err)

		w.Write(jsn)
	})

	require.Nil(t, app.Fetch())
}

func Test_AnApplicationCanReadATagFromTheDisk(t *testing.T) {

	app, _, cleanup := newMockApplication(t)
	defer cleanup()

	expectedPath := filepath.Join(app.workingDir, repoOwner, repoName)
	require.Nil(t, os.MkdirAll(filepath.Join(app.workingDir, repoOwner), os.ModePerm))
	require.Nil(t, ioutil.WriteFile(expectedPath, []byte("tag0"), os.ModePerm))

	require.Nil(t, app.assertTag(repoOwner, repoName, "tag1"))
	require.Equal(t, fmt.Errorf("Already up to date, nothing to download"), app.assertTag(repoOwner, repoName, "tag0"))
}

func Test_AnApplicationCanWriteATagToDisk(t *testing.T) {

	app, _, cleanup := newMockApplication(t)
	defer cleanup()

	require.Nil(t, app.writeTag(repoOwner, repoName, "tag0"))

	expectedPath := filepath.Join(app.workingDir, repoOwner, repoName)

	content, err := ioutil.ReadFile(expectedPath)
	require.Nil(t, err)

	require.Equal(t, "tag0", string(content))
}

func Test_AnApplicationCanFindTheDownloadURLFromARepositoryRelease(t *testing.T) {

	app, _, cleanup := newMockApplication(t)
	defer cleanup()

	names := []string{"A"}
	tarballURL := "http://github.com/release/tarball.tar"
	assetURL := "http://github.com/release/asset.file"

	release := &github.RepositoryRelease{
		TarballURL: &tarballURL,
		Assets: []github.ReleaseAsset{
			github.ReleaseAsset{
				Name:               &names[0],
				BrowserDownloadURL: &assetURL,
			},
		},
	}

	for cycle, test := range []struct {
		requestedAsset string
		url            string
		err            error
	}{
		{
			"doesn't exist",
			"",
			fmt.Errorf("Can't find 'doesn't exist' in release assets"),
		},
		{
			"A",
			assetURL,
			nil,
		},
		{
			"",
			tarballURL,
			nil,
		},
	} {
		t.Logf("Cycle %d", cycle)

		app.requestedAsset = test.requestedAsset
		url, err := app.findDownloadURL(release)
		require.Equal(t, test.url, url)
		require.Equal(t, test.err, err)
	}

}

func Test_ANewApplicationCanBeCreatedWithAnOwnerRepoAssetAndOutput(t *testing.T) {
	require.NotNil(t, newApplication(repoOwner, repoName, outputFile))
}

func Test_AnApplicationCanChooseTheProperOutputPathGivenADownloadURL(t *testing.T) {

	app, _, cleanup := newMockApplication(t)
	defer cleanup()

	app.requestedOutput = ""
	require.Equal(t, assetName, app.findOutputPath("http://test.com/asset.file"))

	app.requestedOutput = "my-output.file"
	require.Equal(t, "my-output.file", app.findOutputPath("http://test.com/asset.file"))
}

func Test_AnApplicatioNCanFindTheOwnerAndRepoFromAShorthandsString(t *testing.T) {

	app, _, cleanup := newMockApplication(t)
	defer cleanup()

	for cycle, test := range []struct {
		input string
		owner string
		repo  string
		err   error
	}{
		{
			"kowala-tech/kUSD",
			"kowala-tech",
			"kUSD",
			nil,
		},
		{
			"too-short",
			"",
			"",
			fmt.Errorf("Expected owner/repo format"),
		},
	} {
		t.Logf("Cycle %d", cycle)

		owner, repo, err := app.findOwnerAndRepo(test.input)
		require.Equal(t, test.owner, owner)
		require.Equal(t, test.repo, repo)
		require.Equal(t, test.err, err)
	}

}

func Test_AnApplicationCanFindAReleaseAssetFromASet(t *testing.T) {

	app, _, cleanup := newMockApplication(t)
	defer cleanup()

	names := []string{"A", "B"}

	sampleAssets := []github.ReleaseAsset{
		github.ReleaseAsset{Name: &names[0]},
		github.ReleaseAsset{Name: &names[1]},
	}

	for cycle, test := range []struct {
		name  string
		asset github.ReleaseAsset
		err   error
	}{
		{
			names[0],
			sampleAssets[0],
			nil,
		},
		{
			names[1],
			sampleAssets[1],
			nil,
		},
		{
			"doesn't exist",
			github.ReleaseAsset{},
			fmt.Errorf("Can't find 'doesn't exist' in release assets"),
		},
	} {
		t.Logf("Cycle %d", cycle)

		asset, err := app.findReleaseAsset(test.name, sampleAssets)
		require.Equal(t, test.asset, asset)
		require.Equal(t, test.err, err)
	}
}

func Test_AnApplicationCanDownloadANominatedFile(t *testing.T) {

	app, _, cleanup := newMockApplication(t)
	defer cleanup()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, assetContent)
	}))
	defer ts.Close()

	require.Nil(t, app.downloadFile(app.requestedOutput, ts.URL))

	downloadedContent, err := ioutil.ReadFile(app.requestedOutput)
	require.Nil(t, err)

	require.Equal(t, assetContent, string(downloadedContent))
}
