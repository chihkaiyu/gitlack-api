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

var listProjectResponse = []byte(`
[
    {
        "id": 1,
        "path_with_namespace": "fake-a/fake-b/fake-c"
    },
    {
        "id": 2,
        "path_with_namespace": "fake.a/fake.b/fake.c"
    },
    {
        "id": 3,
        "path_with_namespace": "fake_a/fake_b/fake_c"
	}
]
`)

var listProjectResponse2 = []byte(`
[
    {
        "id": 4,
        "path_with_namespace": "fake-a/fake.b/fake_c"
    }
]
`)

func TestGetProjectOnePage(t *testing.T) {
	fakeGitLabToken := "fake-gitlab-token"
	fakeGitLabDomain := "fake"
	fakeGitLabAPI := fmt.Sprintf("https://%v/api/v4", fakeGitLabDomain)
	var mapNil map[string]string
	params := map[string]string{
		"private_token": fakeGitLabToken,
		"per_page":      "100",
		"archived":      "false",
		"simple":        "true",
	}
	mockedRes := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(listProjectResponse)),
		Header:     map[string][]string{"X-Next-Page": []string{""}},
	}

	mockedClient := &mocks.Client{}
	mockedClient.On("Get", fakeGitLabAPI+"/projects", mapNil, params, mapNil).Return(mockedRes, nil)

	g := &gitlab{
		client:       mockedClient,
		GitLabDomain: fakeGitLabDomain,
		GitLabAPI:    fakeGitLabAPI,
		GitLabToken:  fakeGitLabToken,
	}

	projects, err := g.GetProject()

	expected := []map[string]interface{}{
		map[string]interface{}{
			"ID":   1,
			"Name": "fake-a/fake-b/fake-c",
		},
		map[string]interface{}{
			"ID":   2,
			"Name": "fake.a/fake.b/fake.c",
		},
		map[string]interface{}{
			"ID":   3,
			"Name": "fake_a/fake_b/fake_c",
		},
	}
	assert := assert.New(t)
	mockedClient.AssertNumberOfCalls(t, "Get", 1)
	assert.Equal(3, len(projects), "number of projects should be equal")

	for i, e := range expected {
		assert.Equal(projects[i].ID, e["ID"], "id are different")
		assert.Equal(projects[i].Name, e["Name"], "name are different")
	}

	assert.Nil(err, "error should be nil")
}

func TestGetProjectMultiplePages(t *testing.T) {
	fakeGitLabToken := "fake-gitlab-token"
	fakeGitLabDomain := "fake"
	fakeGitLabAPI := fmt.Sprintf("https://%v/api/v4", fakeGitLabDomain)
	var mapNil map[string]string
	withoutPageParams := map[string]string{
		"private_token": fakeGitLabToken,
		"per_page":      "100",
		"archived":      "false",
		"simple":        "true",
	}
	mockedWithPageRes := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(listProjectResponse)),
		Header:     map[string][]string{"X-Next-Page": []string{"fake-next-page"}},
	}
	withPageParams := map[string]string{
		"private_token": fakeGitLabToken,
		"per_page":      "100",
		"archived":      "false",
		"simple":        "true",
		"page":          "fake-next-page",
	}
	mockedWithoutPageRes := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(listProjectResponse2)),
		Header:     map[string][]string{"X-Next-Page": []string{""}},
	}

	mockedClient := &mocks.Client{}
	mockedClient.On("Get", fakeGitLabAPI+"/projects", mapNil, withoutPageParams, mapNil).Return(mockedWithPageRes, nil)
	mockedClient.On("Get", fakeGitLabAPI+"/projects", mapNil, withPageParams, mapNil).Return(mockedWithoutPageRes, nil)

	g := &gitlab{
		client:       mockedClient,
		GitLabDomain: fakeGitLabDomain,
		GitLabAPI:    fakeGitLabAPI,
		GitLabToken:  fakeGitLabToken,
	}

	projects, err := g.GetProject()

	expected := []map[string]interface{}{
		map[string]interface{}{
			"ID":   1,
			"Name": "fake-a/fake-b/fake-c",
		},
		map[string]interface{}{
			"ID":   2,
			"Name": "fake.a/fake.b/fake.c",
		},
		map[string]interface{}{
			"ID":   3,
			"Name": "fake_a/fake_b/fake_c",
		},
		map[string]interface{}{
			"ID":   4,
			"Name": "fake-a/fake.b/fake_c",
		},
	}
	assert := assert.New(t)
	mockedClient.AssertNumberOfCalls(t, "Get", 2)
	assert.Equal(4, len(projects), "number of project shoule be equal")

	for i, e := range expected {
		assert.Equal(projects[i].ID, e["ID"], "id shoule be equal")
		assert.Equal(projects[i].Name, e["Name"], "name shoule be equal")
		assert.Equal(projects[i].DefaultChannel, "", "default channel shoule be equal")
	}

	assert.Nil(err, "error shoule be nil")
}

func TestGetProjectRequestFail(t *testing.T) {
	fakeGitLabToken := "fake-gitlab-token"
	fakeGitLabDomain := "fake"
	fakeGitLabAPI := fmt.Sprintf("https://%v/api/v4", fakeGitLabDomain)
	var mapNil map[string]string
	params := map[string]string{
		"private_token": fakeGitLabToken,
		"per_page":      "100",
		"archived":      "false",
		"simple":        "true",
	}
	mockedErr := errors.New("fake-error")

	mockedClient := &mocks.Client{}
	mockedClient.On("Get", fakeGitLabAPI+"/projects", mapNil, params, mapNil).Return(nil, mockedErr)

	g := &gitlab{
		client:       mockedClient,
		GitLabDomain: fakeGitLabDomain,
		GitLabAPI:    fakeGitLabAPI,
		GitLabToken:  fakeGitLabToken,
	}

	projects, err := g.GetProject()

	assert := assert.New(t)
	mockedClient.AssertNumberOfCalls(t, "Get", 1)
	assert.Equal(err.Error(), "fake-error", "error messages should be equal")
	assert.Nil(projects)
}

func TestGetProjectGitLabError(t *testing.T) {
	fakeGitLabToken := "fake-gitlab-token"
	fakeGitLabDomain := "fake"
	fakeGitLabAPI := fmt.Sprintf("https://%v/api/v4", fakeGitLabDomain)
	var mapNil map[string]string
	params := map[string]string{
		"private_token": fakeGitLabToken,
		"per_page":      "100",
		"archived":      "false",
		"simple":        "true",
	}
	mockedRes := &http.Response{
		StatusCode: 304,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte("fake-gitlab-error"))),
		Header:     map[string][]string{"X-Next-Page": []string{""}},
	}

	mockedClient := &mocks.Client{}
	mockedClient.On("Get", fakeGitLabAPI+"/projects", mapNil, params, mapNil).Return(mockedRes, nil)

	g := &gitlab{
		client:       mockedClient,
		GitLabDomain: fakeGitLabDomain,
		GitLabAPI:    fakeGitLabAPI,
		GitLabToken:  fakeGitLabToken,
	}

	projects, err := g.GetProject()

	assert := assert.New(t)
	mockedClient.AssertNumberOfCalls(t, "Get", 1)
	assert.Equal(err.Error(), "GitLab error: fake-gitlab-error", "error messages should be equal")
	assert.Nil(projects)
}
