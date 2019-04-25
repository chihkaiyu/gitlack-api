package slack

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"gitlack/resource/mocks"

	"github.com/stretchr/testify/assert"
)

var withCursorResponse = []byte(`
{
    "ok": true,
	"offset": "fake-offset",
	"response_metadata": {
        "next_cursor": "fake-next-cursor"
    },
    "members": [
        {
            "id": "slackbot-id",
            "name": "slackbot",
			"deleted": false,
			"is_bot": false,
            "profile": {
                "title": "",
                "first_name": "slackbot"
            }
		},
		{
            "id": "fake-id",
            "name": "fake-name",
			"deleted": false,
			"is_bot": false,
            "profile": {
                "title": "fake-title",
                "email": "fake@fake.com"
            }
		},
		{
            "id": "deleted-id",
            "name": "deleted-name",
			"deleted": true,
			"is_bot": false,
            "profile": {
                "title": "deleted-title",
                "email": "deleted@fake.com"
            }
		},
		{
            "id": "bot-id",
            "name": "bot-name",
			"deleted": false,
			"is_bot": true,
            "profile": {
                "title": "bot-title",
                "email": "bot@fake.com"
            }
		}
	]
}
`)

var withoutCursorResponse = []byte(`
{
    "ok": true,
	"offset": "fake-offset",
    "members": [
		{
            "id": "fake-id-2",
            "name": "fake-name-2",
			"deleted": false,
			"is_bot": false,
            "profile": {
                "title": "fake-title-2",
                "email": "fake-2@fake.com"
            }
		},
		{
            "id": "deleted-id",
            "name": "deleted-name",
			"deleted": true,
			"is_bot": false,
            "profile": {
                "title": "deleted-title",
                "email": "deleted@fake.com"
            }
		},
		{
            "id": "bot-id",
            "name": "bot-name",
			"deleted": false,
			"is_bot": true,
            "profile": {
                "title": "bot-title",
                "email": "bot@fake.com"
            }
		}
	]
}
`)

func TestGetUserOnePage(t *testing.T) {
	fakeSlackAPI := "https://fake.slack.com/api"
	var mapNil map[string]string
	params := map[string]string{
		"token": "fake-slack-token",
		"limit": "100",
	}
	mockedRes := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(withoutCursorResponse)),
	}
	mockedClient := &mocks.Client{}
	mockedClient.On("Get", fakeSlackAPI+"/users.list", mapNil, params, mapNil).Return(mockedRes, nil)

	s := &slack{
		client:     mockedClient,
		SlackAPI:   fakeSlackAPI,
		SlackToken: "fake-slack-token",
	}

	users, err := s.GetUser()

	expected := []map[string]string{
		map[string]string{
			"ID":    "fake-id-2",
			"Email": "fake-2@fake.com",
		},
	}
	assert := assert.New(t)
	mockedClient.AssertNumberOfCalls(t, "Get", 1)
	assert.Equal(1, len(users), "number of user should be equal")

	for i, e := range expected {
		assert.Equal(users[i].ID, e["ID"], "id are different")
		assert.Equal(users[i].Email, e["Email"], "email are different")
	}

	assert.Nil(err, "error should be nil")
}

func TestGetUserMultiplePages(t *testing.T) {
	fakeSlackAPI := "https://fake.slack.com/api"
	var mapNil map[string]string
	withoutCursorParams := map[string]string{
		"token": "fake-slack-token",
		"limit": "100",
	}
	mockedWithCursorRes := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(withCursorResponse)),
	}
	withCursorParams := map[string]string{
		"token":  "fake-slack-token",
		"limit":  "100",
		"cursor": "fake-next-cursor",
	}
	mockedWithoutCursorRes := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(withoutCursorResponse)),
	}

	mockedClient := &mocks.Client{}
	mockedClient.On("Get", fakeSlackAPI+"/users.list", mapNil, withoutCursorParams, mapNil).Return(mockedWithCursorRes, nil)
	mockedClient.On("Get", fakeSlackAPI+"/users.list", mapNil, withCursorParams, mapNil).Return(mockedWithoutCursorRes, nil)

	s := &slack{
		client:     mockedClient,
		SlackAPI:   fakeSlackAPI,
		SlackToken: "fake-slack-token",
	}

	users, err := s.GetUser()

	expected := []map[string]string{
		map[string]string{
			"ID":    "slackbot-id",
			"Email": "",
		},
		map[string]string{
			"ID":    "fake-id",
			"Email": "fake@fake.com",
		},
		map[string]string{
			"ID":    "fake-id-2",
			"Email": "fake-2@fake.com",
		},
	}
	assert := assert.New(t)
	mockedClient.AssertNumberOfCalls(t, "Get", 2)
	assert.Equal(3, len(users), "number of user should be equal")

	for i, e := range expected {
		assert.Equal(users[i].ID, e["ID"], "id are different")
		assert.Equal(users[i].Email, e["Email"], "email are different")
	}

	assert.Nil(err, "error should be nil")
}

func TestGetUserRequestFail(t *testing.T) {
	fakeSlackAPI := "https://fake.slack.com/api"
	var mapNil map[string]string
	params := map[string]string{
		"token": "fake-slack-token",
		"limit": "100",
	}
	mockedErr := errors.New("fake-error")

	mockedClient := &mocks.Client{}
	mockedClient.On("Get", fakeSlackAPI+"/users.list", mapNil, params, mapNil).Return(nil, mockedErr)

	s := &slack{
		client:     mockedClient,
		SlackAPI:   fakeSlackAPI,
		SlackToken: "fake-slack-token",
	}

	users, err := s.GetUser()

	assert := assert.New(t)
	mockedClient.AssertNumberOfCalls(t, "Get", 1)
	assert.Equal(err.Error(), "fake-error", "error messages shoule be equal")
	assert.Nil(users, "users shoule be nil")
}

func TestGetUserSlackError(t *testing.T) {
	fakeSlackAPI := "https://fake.slack.com/api"
	var mapNil map[string]string
	params := map[string]string{
		"token": "fake-slack-token",
		"limit": "100",
	}
	mockedRes := &http.Response{
		StatusCode: 304,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte("fake-slack-error"))),
	}

	mockedClient := &mocks.Client{}
	mockedClient.On("Get", fakeSlackAPI+"/users.list", mapNil, params, mapNil).Return(mockedRes, nil)

	s := &slack{
		client:     mockedClient,
		SlackAPI:   fakeSlackAPI,
		SlackToken: "fake-slack-token",
	}

	users, err := s.GetUser()

	assert := assert.New(t)
	mockedClient.AssertNumberOfCalls(t, "Get", 1)
	assert.Equal(err.Error(), "Slack error: fake-slack-error", "error messages shoule be equal")
	assert.Nil(users, "users should be nil")
}

// mockedRes can't be read twice and I don't know how to deal with it yet
func GetUserInfiniteLoop(t *testing.T) {
	fakeSlackAPI := "https://fake.slack.com/api"
	var mapNil map[string]string
	withoutCursorParams := map[string]string{
		"token": "fake-slack-token",
		"limit": "100",
	}
	withCursorParams := map[string]string{
		"token":  "fake-slack-token",
		"limit":  "100",
		"cursor": "fake-next-cursor",
	}
	mockedRes := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(withCursorResponse)),
	}

	mockedClient := &mocks.Client{}
	mockedClient.On("Get", fakeSlackAPI+"/users.list", mapNil, withoutCursorParams, mapNil).Return(mockedRes, nil)
	mockedClient.On("Get", fakeSlackAPI+"/users.list", mapNil, withCursorParams, mapNil).Return(mockedRes, nil)

	s := &slack{
		client:     mockedClient,
		SlackAPI:   fakeSlackAPI,
		SlackToken: "fake-slack-token",
	}

	users, err := s.GetUser()

	assert := assert.New(t)
	mockedClient.AssertNumberOfCalls(t, "Get", 100)
	assert.Equal(200, len(users), "number of user should be equal")
	assert.Nil(err)

}
