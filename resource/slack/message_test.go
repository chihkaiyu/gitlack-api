package slack

import (
	"encoding/json"
	"gitlack/model"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostSlackMessageWithOnlyChannelAndText(t *testing.T) {
	// arrange
	channel := "fake-channel"
	text := "fake-text"
	expected := getRequestBody()
	expected["channel"] = channel
	expected["text"] = text
	stubClient := getPostClientWithRequestBody(getOKResponse(), http.StatusOK, expected)
	s := getSlack(stubClient)

	// assert
	stubClient.On(
		"Post",
		"/chat.postMessage",
		getURLEncodedHeader(),
		mapNil,
		expected).Return(nil, nil)

	// act
	s.PostSlackMessage(channel, text, nil, nil)
}

func TestPostSlackMessageWithAuthor(t *testing.T) {
	// arrange
	author := &model.User{
		Name:      "fake-name",
		AvatarURL: "fake-icon-url",
	}
	expected := getRequestBody()
	expected["username"] = "fake-name (Gitlack)"
	expected["icon_url"] = "fake-icon-url"
	stubClient := getPostClientWithRequestBody(getOKResponse(), http.StatusOK, expected)
	s := getSlack(stubClient)

	// assert
	stubClient.On(
		"Post",
		"/chat.postMessage",
		getURLEncodedHeader(),
		mapNil,
		expected).Return(nil, nil)

	// act
	s.PostSlackMessage("", "", author, nil)
}

func TestPostSlackMessageWithAttachment(t *testing.T) {
	// arrange
	atm := &Attachment{
		Color: "fake-color",
		Title: "fake-title",
		Text:  "fake-text",
	}
	marshaledAtm, _ := json.Marshal([]*Attachment{atm})

	expected := getRequestBody()
	expected["attachments"] = string(marshaledAtm)
	stubClient := getPostClientWithRequestBody(getOKResponse(), http.StatusOK, expected)
	s := getSlack(stubClient)

	// assert
	stubClient.On(
		"Post",
		"/chat.postMessage",
		getURLEncodedHeader(),
		mapNil,
		expected).Return(nil, nil)

	// act
	s.PostSlackMessage("", "", nil, atm)
}

func TestPostSlackMessageWithThread(t *testing.T) {
	// arrange
	threadTS := "fake-thread-ts"
	expected := getRequestBody()
	expected["thread_ts"] = threadTS
	stubClient := getPostClientWithRequestBody(getOKResponse(), http.StatusOK, expected)
	s := getSlack(stubClient)

	// assert
	stubClient.On(
		"Post",
		"/chat.postMessage",
		getURLEncodedHeader(),
		mapNil,
		expected).Return(nil, nil)

	// act
	s.PostSlackMessage("", "", nil, nil, threadTS)
}

func TestPostSlackMessageRequestError(t *testing.T) {
	// arrange
	stubClient := getPostClientWithError("fake-error")
	s := getSlack(stubClient)

	// act
	_, err := s.PostSlackMessage("", "", nil, nil)

	// assert
	assert.NotNil(t, err, "Error should not be nil")
	assert.Equal(t, "fake-error", err.Error(), "Error message should be equal")
}

func TestPostSlackMessageResponseError(t *testing.T) {
	// arrange
	stubClient := getPostClientWithResponse([]byte(`fake-body`), http.StatusNotFound)
	s := getSlack(stubClient)

	// act
	_, err := s.PostSlackMessage("", "", nil, nil)

	// assert
	assert.NotNil(t, err, "Error should not be nil")
	assert.Equal(t, "HTTP response error: fake-body", err.Error(), "Error message should be equal")
}

func TestPostSlackMessageInvalidSlackAPI(t *testing.T) {
	// arrange
	stubClient := getPostClientWithResponse(getErrorResponse(), http.StatusOK)
	s := getSlack(stubClient)

	// act
	_, err := s.PostSlackMessage("", "", nil, nil)

	// assert
	assert.NotNil(t, err, "err should not be nil")
	assert.Equal(t, "Invalid Slack API: fake-error", err.Error(), "err should be fake-error")
}
