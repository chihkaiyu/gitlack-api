package slack

import (
	"bytes"
	"encoding/json"
	"gitlack/resource/mocks"
	"io/ioutil"
	"net/http"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

var slackResponseTemplate = `
{
    "ok": {{.OK}},
    "channel": "{{.Channel}}",
    "ts": "{{.TS}}"
}
`

var fakeFailSlackResponse = []byte(`
{
    "ok": false,
    "error": "not_authed"
}
`)

func renderTemplate(tpl string, data interface{}) *bytes.Buffer {
	t := template.Must(template.New("tpl").Parse(tpl))
	rendered := &bytes.Buffer{}
	t.Execute(rendered, data)
	return rendered
}

func TestPostSlackMessage(t *testing.T) {
	fakeSlackAPI := "https://fake.slack.com/api"
	fakeSlackToken := "fake-slack-token"
	fakeURL := fakeSlackAPI + "/chat.postMessage"
	fakeHeader := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	fakeChannel := "fake-channel"
	fakeText := "fake-text"
	fakeReqBody := map[string]string{
		"token":   fakeSlackToken,
		"channel": fakeChannel,
		"text":    fakeText,
	}
	expectedRes := map[string]interface{}{
		"OK": "true",
		"Channel": fakeReqBody["channel"],
		"TS": "1234567890.123456",
	}
	slackRes := renderTemplate(slackResponseTemplate, expectedRes)
	mockedRes := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(slackRes.Bytes())),
	}
	mockedClient := &mocks.Client{}
	var mapNil map[string]string
	mockedClient.On("Post", fakeURL, fakeHeader, mapNil, fakeReqBody).Return(mockedRes, nil)

	s := &slack{
		client:     mockedClient,
		SlackAPI:   fakeSlackAPI,
		SlackToken: fakeSlackToken,
	}

	smr, err := s.PostSlackMessage(fakeChannel, fakeText, nil)

	assert := assert.New(t)
	mockedClient.AssertNumberOfCalls(t, "Post", 1)
	assert.Nil(err)
	assert.True(smr.OK)
	assert.Equal(expectedRes["Channel"].(string), smr.Channel, "channel should be equal")
	assert.Equal(expectedRes["TS"].(string), smr.TS, "timestamp should be equal")
}

func TestPostSlackMessageWithAttachment(t *testing.T) {
	fakeSlackAPI := "https://fake.slack.com/api"
	fakeSlackToken := "fake-slack-token"
	fakeURL := fakeSlackAPI + "/chat.postMessage"
	fakeHeader := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	fakeChannel := "fake-channel"
	fakeText := "fake-text"
	fakeAtm := &Attachment{
		Color: AttachmentColor,
		Title: "fake-title",
		Text: fakeText,

	}
	serAtm, _ := json.Marshal([]*Attachment{fakeAtm})
	fakeReqBody := map[string]string{
		"token":   fakeSlackToken,
		"channel": fakeChannel,
		"text":    fakeText,
		"attachments": string(serAtm),
	}
	
	expectedRes := map[string]interface{}{
		"OK": "true",
		"Channel": fakeReqBody["channel"],
		"TS": "1234567890.123456",
	}
	slackRes := renderTemplate(slackResponseTemplate, expectedRes)
	mockedRes := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(slackRes.Bytes())),
	}
	mockedClient := &mocks.Client{}
	var mapNil map[string]string
	mockedClient.On("Post", fakeURL, fakeHeader, mapNil, fakeReqBody).Return(mockedRes, nil)

	s := &slack{
		client:     mockedClient,
		SlackAPI:   fakeSlackAPI,
		SlackToken: fakeSlackToken,
	}

	smr, err := s.PostSlackMessage(fakeChannel, fakeText, fakeAtm)

	assert := assert.New(t)
	mockedClient.AssertNumberOfCalls(t, "Post", 1)
	assert.Nil(err)
	assert.True(smr.OK)
	assert.Equal(fakeChannel, smr.Channel, "channel should be equal")
	assert.Equal("1234567890.123456", smr.TS, "timestamp should be equal")
}

func TestPostSlackMessageWithThread(t *testing.T) {
	fakeSlackAPI := "https://fake.slack.com/api"
	fakeSlackToken := "fake-slack-token"
	fakeURL := fakeSlackAPI + "/chat.postMessage"
	fakeHeader := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	fakeChannel := "fake-channel"
	fakeText := "fake-text"
	fakeThreadTS := "9876543210.654321"
	fakeReqBody := map[string]string{
		"token":     fakeSlackToken,
		"channel":   fakeChannel,
		"text":      fakeText,
		"thread_ts": fakeThreadTS,
	}
	expectedRes := map[string]interface{}{
		"OK": "true",
		"Channel": fakeReqBody["channel"],
		"TS": "1234567890.123456",
	}
	slackRes := renderTemplate(slackResponseTemplate, expectedRes)
	mockedRes := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(slackRes.Bytes())),
	}
	mockedClient := &mocks.Client{}
	var mapNil map[string]string
	mockedClient.On("Post", fakeURL, fakeHeader, mapNil, fakeReqBody).Return(mockedRes, nil)

	s := &slack{
		client:     mockedClient,
		SlackAPI:   fakeSlackAPI,
		SlackToken: fakeSlackToken,
	}

	smr, err := s.PostSlackMessage(fakeChannel, fakeText, nil, fakeThreadTS)

	assert := assert.New(t)
	mockedClient.AssertNumberOfCalls(t, "Post", 1)
	assert.Nil(err)
	assert.True(smr.OK)
	assert.Equal(fakeChannel, smr.Channel, "channel should be equal")
	assert.Equal("1234567890.123456", smr.TS, "timestamp should be equal")
}

func TestPostSlackMessageFail(t *testing.T) {
	fakeSlackAPI := "https://fake.slack.com/api"
	fakeSlackToken := "fake-slack-token"
	fakeURL := fakeSlackAPI + "/chat.postMessage"
	fakeHeader := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	fakeChannel := "fake-channel"
	fakeText := "fake-text"
	fakeThreadTS := "9876543210.654321"
	fakeReqBody := map[string]string{
		"token":     fakeSlackToken,
		"channel":   fakeChannel,
		"text":      fakeText,
		"thread_ts": fakeThreadTS,
	}

	mockedRes := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(fakeFailSlackResponse)),
	}
	mockedClient := &mocks.Client{}
	var mapNil map[string]string
	mockedClient.On("Post", fakeURL, fakeHeader, mapNil, fakeReqBody).Return(mockedRes, nil)

	s := &slack{
		client:     mockedClient,
		SlackAPI:   fakeSlackAPI,
		SlackToken: fakeSlackToken,
	}

	smr, err := s.PostSlackMessage(fakeChannel, fakeText, nil, fakeThreadTS)

	assert := assert.New(t)
	mockedClient.AssertNumberOfCalls(t, "Post", 1)
	assert.Nil(smr)
	assert.Equal("Invalid Slack API: not_authed", err.Error(), "error message should be equal")
}
