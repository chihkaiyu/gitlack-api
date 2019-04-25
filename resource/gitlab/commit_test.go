package gitlab

import (
	"bytes"
	"fmt"
	"gitlack/resource/mocks"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var singleCommitResponse = []byte(`
{
	"committer_email": "fake-1@fake.com",
	"last_pipeline": {
		"id": 1,
		"status": "success",
		"WebURL": "http://gitlab.fake.com"
	}
}
`)

func TestGetSingleCommit(t *testing.T) {
	fakeID := 999
	fakeSha := "fake-sha"
	fakeGitLabToken := "fake-gitlab-token"
	fakeGitLabDomain := "fake"
	fakeGitLabAPI := fmt.Sprintf("https://%v/api/v4", fakeGitLabDomain)
	var mapNil map[string]string
	params := map[string]string{
		"private_token": fakeGitLabToken,
	}
	mockedRes := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(singleCommitResponse)),
	}

	mockedClient := &mocks.Client{}
	mockedClient.On("Get", fakeGitLabAPI+fmt.Sprintf("/projects/%v/repository/commits/%v", fakeID, fakeSha), mapNil, params, mapNil).Return(mockedRes, nil)

	g := &gitlab{
		client:       mockedClient,
		GitLabDomain: fakeGitLabDomain,
		GitLabAPI:    fakeGitLabAPI,
		GitLabToken:  fakeGitLabToken,
	}

	_, err := g.GetSingleCommit(fakeID, fakeSha)

	assert := assert.New(t)
	mockedClient.AssertNumberOfCalls(t, "Get", 1)
	assert.Nil(err)
}
