package download

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/alecthomas/assert"
	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/release/releaseutils/upload"
	"gitlab.com/gitlab-org/cli/pkg/httpmock"
	"gitlab.com/gitlab-org/cli/pkg/iostreams"
)

func doesFileExist(fileName string) bool {
	_, error := os.Stat(fileName)

	if os.IsNotExist(error) {
		return false
	} else {
		return true
	}
}

func Test_downloadAssets(t *testing.T) {
	assetUrl := "https://gitlab.com/gitlab-org/cli/-/archive/"

	fakeHTTP := &httpmock.Mocker{
		MatchURL: httpmock.HostAndPath,
	}

	client, _ := api.TestClient(&http.Client{Transport: fakeHTTP}, "", "", false)

	var tests = []struct {
		name     string
		filename string
		path     string
		want     string
		errWant  error
	}{
		{
			name:     "A regular filename",
			filename: "cli-v1.22.0.tar",
			want:     "cli-v1.22.0.tar",
		},
		{
			name:     "A filename with directory traversal",
			filename: "cli-v1.../../22.0.tar",
			want:     "cli-v1.22.0.tar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullUrl := assetUrl + tt.filename
			fakeHTTP.RegisterResponder("GET", fullUrl, httpmock.NewStringResponse(200, `test_data`))

			io, _, _, _ := iostreams.Test()

			release := &upload.ReleaseAsset{
				Name:     &tt.filename,
				URL:      &fullUrl,
			}

			releases := []*upload.ReleaseAsset{release}

			filePathWanted := "./" + tt.want

			err := downloadAssets(client, io, releases, tt.path)

			assert.True(t, doesFileExist(filePathWanted), "File should exist")
			assert.NoError(t, err, "Should not have errors")

			if doesFileExist(filePathWanted) {
				removeError := os.Remove(filePathWanted) // remove any leftover test files
				if removeError != nil {
					fmt.Println("file error")
				}
			}
		})
	}
}
