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

var listUserResponse = []byte(`
[
    {
        "id": 1,
		"email": "fake.1@fake.com",
		"name": "fake-name-1"
	},
	{
        "id": 2,
		"email": "fake.2@fake.com",
		"name": "fake-name-2"
	},
	{
        "id": 3,
		"email": "fake.3@fake.com",
		"name": "fake-name-3"
	}
]
`)

var listUserResponse2 = []byte(`
[
    {
        "id": 4,
		"email": "fake.4@fake.com",
		"name": "fake-name-4"
	}
]
`)

func TestGetUserOnePage(t *testing.T) {
	fakeGitLabToken := "fake-gitlab-token"
	fakeGitLabDomain := "fake"
	fakeGitLabAPI := fmt.Sprintf("https://%v/api/v4", fakeGitLabDomain)
	var mapNil map[string]string
	params := map[string]string{
		"private_token": fakeGitLabToken,
		"per_page":      "100",
		"active":        "true",
	}
	mockedRes := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(listUserResponse)),
		Header:     map[string][]string{"X-Next-Page": []string{""}},
	}

	mockedUtil := &mocks.Util{}
	mockedUtil.On("Request", http.MethodGet, fakeGitLabAPI+"/users", mapNil, params, mapNil).Return(mockedRes, nil)

	g := &gitlab{
		tool:         mockedUtil,
		GitLabDomain: fakeGitLabDomain,
		GitLabAPI:    fakeGitLabAPI,
		GitLabToken:  fakeGitLabToken,
	}

	users, err := g.GetUser()

	expected := []map[string]interface{}{
		map[string]interface{}{
			"ID":    1,
			"Email": "fake.1@fake.com",
			"Name": "fake-name-1",
		},
		map[string]interface{}{
			"ID":    2,
			"Email": "fake.2@fake.com",
			"Name": "fake-name-2",
		},
		map[string]interface{}{
			"ID":    3,
			"Email": "fake.3@fake.com",
			"Name": "fake-name-3",
		},
	}
	assert := assert.New(t)
	mockedUtil.AssertNumberOfCalls(t, "Request", 1)
	assert.Equal(3, len(users), "number of user should be equal")

	for i, e := range expected {
		assert.Equal(users[i].ID, e["ID"], "id are different")
		assert.Equal(users[i].Email, e["Email"], "email are different")
	}
	assert.Nil(err)
}

func TestGetUserMultiplePage(t *testing.T) {
	fakeGitLabToken := "fake-gitlab-token"
	fakeGitLabDomain := "fake"
	fakeGitLabAPI := fmt.Sprintf("https://%v/api/v4", fakeGitLabDomain)
	var mapNil map[string]string
	withoutPageParams := map[string]string{
		"private_token": fakeGitLabToken,
		"per_page":      "100",
		"active":        "true",
	}
	mockedWithPageRes := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(listUserResponse)),
		Header:     map[string][]string{"X-Next-Page": []string{"fake-next-page"}},
	}
	withPageParams := map[string]string{
		"private_token": fakeGitLabToken,
		"per_page":      "100",
		"active":        "true",
		"page":          "fake-next-page",
	}
	mockedWithoutPageRes := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(listUserResponse2)),
		Header:     map[string][]string{"X-Next-Page": []string{""}},
	}

	mockedUtil := &mocks.Util{}
	mockedUtil.On("Request", http.MethodGet, fakeGitLabAPI+"/users", mapNil, withoutPageParams, mapNil).Return(mockedWithPageRes, nil)
	mockedUtil.On("Request", http.MethodGet, fakeGitLabAPI+"/users", mapNil, withPageParams, mapNil).Return(mockedWithoutPageRes, nil)

	g := &gitlab{
		tool:         mockedUtil,
		GitLabDomain: fakeGitLabDomain,
		GitLabAPI:    fakeGitLabAPI,
		GitLabToken:  fakeGitLabToken,
	}

	users, err := g.GetUser()

	expected := []map[string]interface{}{
		map[string]interface{}{
			"ID":    1,
			"Email": "fake.1@fake.com",
			"Name": "fake-name-1",
		},
		map[string]interface{}{
			"ID":    2,
			"Email": "fake.2@fake.com",
			"Name": "fake-name-2",
		},
		map[string]interface{}{
			"ID":    3,
			"Email": "fake.3@fake.com",
			"Name": "fake-name-3",
		},
		map[string]interface{}{
			"ID":    4,
			"Email": "fake.4@fake.com",
			"Name": "fake-name-4",
		},
	}
	assert := assert.New(t)
	mockedUtil.AssertNumberOfCalls(t, "Request", 2)
	assert.Equal(4, len(users), "number of users should be equal")

	for i, u := range expected {
		assert.Equal(users[i].ID, u["ID"], "id are different")
		assert.Equal(users[i].Email, u["Email"], "email are different")
	}
	assert.Nil(err, "error should be nil")
}

func TestGetUserRequestFail(t *testing.T) {
	fakeGitLabToken := "fake-gitlab-token"
	fakeGitLabDomain := "fake"
	fakeGitLabAPI := fmt.Sprintf("https://%v/api/v4", fakeGitLabDomain)
	var mapNil map[string]string
	params := map[string]string{
		"private_token": fakeGitLabToken,
		"per_page":      "100",
		"active":        "true",
	}
	mockedErr := errors.New("fake-error")

	mockedUtil := &mocks.Util{}
	mockedUtil.On("Request", http.MethodGet, fakeGitLabAPI+"/users", mapNil, params, mapNil).Return(nil, mockedErr)

	g := &gitlab{
		tool:         mockedUtil,
		GitLabDomain: fakeGitLabDomain,
		GitLabAPI:    fakeGitLabAPI,
		GitLabToken:  fakeGitLabToken,
	}

	users, err := g.GetUser()

	assert := assert.New(t)
	mockedUtil.AssertNumberOfCalls(t, "Request", 1)
	assert.Equal(err.Error(), "fake-error", "error messages should be equal")
	assert.Nil(users)
}

func TestGetUserGitLabError(t *testing.T) {
	fakeGitLabToken := "fake-gitlab-token"
	fakeGitLabDomain := "fake"
	fakeGitLabAPI := fmt.Sprintf("https://%v/api/v4", fakeGitLabDomain)
	var mapNil map[string]string
	params := map[string]string{
		"private_token": fakeGitLabToken,
		"per_page":      "100",
		"active":        "true",
	}
	mockedRes := &http.Response{
		StatusCode: 304,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte("fake-gitlab-error"))),
		Header:     map[string][]string{"X-Next-Page": []string{""}},
	}

	mockedUtil := &mocks.Util{}
	mockedUtil.On("Request", http.MethodGet, fakeGitLabAPI+"/users", mapNil, params, mapNil).Return(mockedRes, nil)

	g := &gitlab{
		tool:         mockedUtil,
		GitLabDomain: fakeGitLabDomain,
		GitLabAPI:    fakeGitLabAPI,
		GitLabToken:  fakeGitLabToken,
	}

	users, err := g.GetUser()

	assert := assert.New(t)
	mockedUtil.AssertNumberOfCalls(t, "Request", 1)
	assert.Equal(err.Error(), "GitLab error: fake-gitlab-error", "error messages should be equal")
	assert.Nil(users)
}
