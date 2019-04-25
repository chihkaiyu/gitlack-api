package gitlab

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlack/resource/mocks"
)

var listTagResponse = []byte(`
[
	{
		"name": "v0.0.3",
        "commit": {
            "committer_email": "fake-3@fake.com"
        },
        "release": {
            "description": "fake-release-notes-3"
        }
    },
    {
		"name": "v0.0.2",
        "commit": {
            "committer_email": "fake-2@fake.com"
        },
        "release": {
            "description": "fake-release-notes-2"
        }
    },
    {
		"name": "v0.0.1",
        "commit": {
            "committer_email": "fake-1@fake.com"
        },
        "release": {
            "description": "fake-release-notes-1"
        }
    }
]
`)

func TestGetTagList(t *testing.T) {
	fakeID := 999
	fakeGitLabToken := "fake-gitlab-token"
	fakeGitLabDomain := "fake"
	fakeGitLabAPI := fmt.Sprintf("https://%v/api/v4", fakeGitLabDomain)
	var mapNil map[string]string
	params := map[string]string{
		"private_token": fakeGitLabToken,
	}
	mockedRes := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(listTagResponse)),
	}

	mockedClient := &mocks.Client{}
	mockedClient.On("Get", fakeGitLabAPI+fmt.Sprintf("/projects/%v/repository/tags", fakeID), mapNil, params, mapNil).Return(mockedRes, nil)

	g := &gitlab{
		client:       mockedClient,
		GitLabDomain: fakeGitLabDomain,
		GitLabAPI:    fakeGitLabAPI,
		GitLabToken:  fakeGitLabToken,
	}

	tags, err := g.GetTagList(fakeID)

	expectedCommit := []map[string]string{
		map[string]string{
			"CommitterEmail": "fake-3@fake.com",
		},
		map[string]string{
			"CommitterEmail": "fake-2@fake.com",
		},
		map[string]string{
			"CommitterEmail": "fake-1@fake.com",
		},
	}
	expectedRelease := []map[string]string{
		map[string]string{
			"Name":        "v0.0.3",
			"Description": "fake-release-notes-3",
		},
		map[string]string{
			"Name":        "v0.0.2",
			"Description": "fake-release-notes-2",
		},
		map[string]string{
			"Name":        "v0.0.1",
			"Description": "fake-release-notes-1",
		},
	}

	assert := assert.New(t)
	mockedClient.AssertNumberOfCalls(t, "Get", 1)
	assert.Equal(3, len(tags), "number of tag should be equal")

	for i, e := range expectedCommit {
		assert.Equal(tags[i].CommitInfo.CommitterEmail, e["CommitterEmail"], "committer email should be equal")
	}
	for i, e := range expectedRelease {
		assert.Equal(e["Name"], tags[i].Name, "tag name should be equal")
		assert.Equal(e["Description"], tags[i].ReleaseInfo.Description, "description should be equal")
	}

	assert.Nil(err)
}

func TestGetTagListRequestFail(t *testing.T) {
	fakeID := 999
	fakeGitLabToken := "fake-gitlab-token"
	fakeGitLabDomain := "fake"
	fakeGitLabAPI := fmt.Sprintf("https://%v/api/v4", fakeGitLabDomain)
	var mapNil map[string]string
	params := map[string]string{
		"private_token": fakeGitLabToken,
	}
	mockedErr := errors.New("fake-error")

	mockedClient := &mocks.Client{}
	mockedClient.On("Get", fakeGitLabAPI+fmt.Sprintf("/projects/%v/repository/tags", fakeID), mapNil, params, mapNil).Return(nil, mockedErr)

	g := &gitlab{
		client:       mockedClient,
		GitLabDomain: fakeGitLabDomain,
		GitLabAPI:    fakeGitLabAPI,
		GitLabToken:  fakeGitLabToken,
	}

	tags, err := g.GetTagList(fakeID)

	assert := assert.New(t)
	mockedClient.AssertNumberOfCalls(t, "Get", 1)
	assert.Equal(err.Error(), "fake-error", "error messages should be equal")
	assert.Nil(tags)
}

func TestGetTagListGitLabError(t *testing.T) {
	fakeID := 999
	fakeGitLabToken := "fake-gitlab-token"
	fakeGitLabDomain := "fake"
	fakeGitLabAPI := fmt.Sprintf("https://%v/api/v4", fakeGitLabDomain)
	var mapNil map[string]string
	params := map[string]string{
		"private_token": fakeGitLabToken,
	}
	mockedRes := &http.Response{
		StatusCode: 304,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte("fake-gitlab-error"))),
		Header:     map[string][]string{"X-Next-Page": []string{""}},
	}

	mockedClient := &mocks.Client{}
	mockedClient.On("Get", fakeGitLabAPI+fmt.Sprintf("/projects/%v/repository/tags", fakeID), mapNil, params, mapNil).Return(mockedRes, nil)

	g := &gitlab{
		client:       mockedClient,
		GitLabDomain: fakeGitLabDomain,
		GitLabAPI:    fakeGitLabAPI,
		GitLabToken:  fakeGitLabToken,
	}

	tags, err := g.GetTagList(fakeID)

	assert := assert.New(t)
	mockedClient.AssertNumberOfCalls(t, "Get", 1)
	assert.Equal(err.Error(), "GitLab error: fake-gitlab-error", "error messages should be equal")
	assert.Nil(tags)
}
