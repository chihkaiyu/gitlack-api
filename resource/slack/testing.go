package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gitlack/resource/mocks"
	"io/ioutil"
	"net/http"

	"github.com/stretchr/testify/mock"
)

func getSlack(client *mocks.Client) *slack {
	return &slack{
		client: client,
	}
}

func getSlackUserResopnse(bot, deleted, user int, nextCursor bool) []byte {
	fakeAvatarURL := "https://fake.com/a.jpg"

	var members []Member
	// bot
	for i := 0; i < bot; i++ {
		botMember := Member{
			ID:      fmt.Sprintf("bot-%v", i),
			Deleted: false,
			IsBot:   true,
			Profile: Profile{
				Email:     fmt.Sprintf("bot-%v@fake.com", i),
				AvatarURL: fakeAvatarURL,
			},
		}
		members = append(members, botMember)
	}

	// deleted
	for i := 0; i < deleted; i++ {
		deletedMember := Member{
			ID:      fmt.Sprintf("deleted-%v", i),
			Deleted: true,
			IsBot:   false,
			Profile: Profile{
				Email:     fmt.Sprintf("deleted-%v@fake.com", i),
				AvatarURL: fakeAvatarURL,
			},
		}
		members = append(members, deletedMember)
	}

	// user
	for i := 0; i < user; i++ {
		userMember := Member{
			ID:      fmt.Sprintf("user-%v", i),
			Deleted: false,
			IsBot:   false,
			Profile: Profile{
				Email:     fmt.Sprintf("user-%v@fake.com", i),
				AvatarURL: fakeAvatarURL,
			},
		}
		members = append(members, userMember)
	}

	var slackUserResponse = SlackUserResponse{
		OK:      true,
		Members: members,
	}

	if nextCursor {
		slackUserResponse.ResponseMetadata.NextCursor = "fake-next-cursor"
	}

	res, _ := json.Marshal(slackUserResponse)
	return res
}

func getClient() *mocks.Client {
	return &mocks.Client{}
}

func getResponse(stubByte []byte, statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(bytes.NewReader(stubByte)),
	}
}

func getParams(token, limit string) map[string]string {
	params := map[string]string{
		"token": token,
		"limit": limit,
	}
	return params
}

func getGetClientWithResponse(stubByte []byte, statusCode int) *mocks.Client {
	stubResponse := getResponse(stubByte, statusCode)
	stubClient := getClient()
	stubClient.On(
		"Get",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(stubResponse, nil)

	return stubClient
}

func getGetClientWithError(errMsg string) *mocks.Client {
	stubClient := getClient()
	stubClient.On(
		"Get",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(nil, errors.New(errMsg))

	return stubClient
}

var mapNil map[string]string

func getRequestBody() map[string]string {
	return map[string]string{
		"token":   "",
		"channel": "",
		"text":    "",
	}
}

func getChannelAndText() (string, string) {
	return "fake-channel", "fake-text"
}

func getOKResponse() []byte {
	return []byte(`{"ok": true}`)
}

func getErrorResponse() []byte {
	return []byte(`{"ok": false, "error": "fake-error"}`)
}

func getURLEncodedHeader() map[string]string {
	return map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
}

func getPostClientWithResponse(stubByte []byte, statusCode int) *mocks.Client {
	stubReponse := getResponse(stubByte, statusCode)
	stubClient := getClient()
	stubClient.On(
		"Post",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(stubReponse, nil)

	return stubClient
}

func getPostClientWithRequestBody(stubByte []byte, statusCode int, reqBody map[string]string) *mocks.Client {
	stubReponse := getResponse(stubByte, statusCode)
	stubClient := getClient()
	stubClient.On(
		"Post",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		reqBody).Return(stubReponse, nil)

	return stubClient
}

func getPostClientWithError(errMsg string) *mocks.Client {
	stubClient := getClient()
	stubClient.On(
		"Post",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(nil, errors.New(errMsg))

	return stubClient
}
