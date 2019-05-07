package slack

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserOnePage(t *testing.T) {
	// arrange
	stubByte := getSlackUserResopnse(1, 1, 1, false)
	stubClient := getGetClientWithResponse(stubByte, http.StatusOK)
	s := getSlack(stubClient)

	// act
	s.GetUser()

	// assert
	stubClient.AssertNumberOfCalls(t, "Get", 1)
}

func TestGetUserRequestError(t *testing.T) {
	// arrange
	stubClient := getGetClientWithError("fake-error")
	s := getSlack(stubClient)

	// act
	_, err := s.GetUser()

	// assert
	assert.Equal(t, "fake-error", err.Error(), "Error message should be equal")
}

func TestGetUserResponseError(t *testing.T) {
	// arrange
	stubClient := getGetClientWithResponse([]byte(`fake-body`), http.StatusNotFound)
	s := getSlack(stubClient)

	// act
	_, err := s.GetUser()

	// assert
	assert.NotNil(t, err, "Return err should not be nil")
	assert.Equal(t, "HTTP response error: fake-body", err.Error())
}

func TestGetUserInvalidSlackAPI(t *testing.T) {
	// arrange
	stubClient := getGetClientWithResponse(getErrorResponse(), http.StatusOK)
	s := getSlack(stubClient)

	// act
	_, err := s.GetUser()

	// assert
	assert.NotNil(t, err, "err should be nil")
	assert.Equal(t, "Invalid Slack API: fake-error", err.Error())
}

func TestGetUserOneBot(t *testing.T) {
	// arrange
	stubByte := getSlackUserResopnse(1, 0, 0, false)
	stubClient := getGetClientWithResponse(stubByte, 200)
	s := getSlack(stubClient)

	// act
	users, err := s.GetUser()

	// assert
	assert.Nil(t, err, "Return err should be nil")
	assert.Equal(t, 0, len(users), "Return users length should be 0")
}

func TestGetUserOneDeleted(t *testing.T) {
	// arrange
	stubByte := getSlackUserResopnse(0, 1, 0, false)
	stubClient := getGetClientWithResponse(stubByte, 200)
	s := getSlack(stubClient)

	// act
	users, err := s.GetUser()

	// assert
	assert.Nil(t, err, "Return err should be nil")
	assert.Equal(t, 0, len(users), "Return users length should be 0")
}

func TestGetUserOneUser(t *testing.T) {
	// arrange
	stubByte := getSlackUserResopnse(0, 0, 1, false)
	stubClient := getGetClientWithResponse(stubByte, 200)
	s := getSlack(stubClient)

	// act
	users, err := s.GetUser()

	// assert
	assert.Nil(t, err, "Return err should be nil")
	assert.Equal(t, 1, len(users), "Return users length should be 0")
}

func TestGetUserOneBotOneDeletedOneUser(t *testing.T) {
	// arrange
	stubByte := getSlackUserResopnse(1, 1, 1, false)
	stubClient := getGetClientWithResponse(stubByte, 200)
	s := getSlack(stubClient)

	// act
	users, err := s.GetUser()

	// assert
	assert.Nil(t, err, "Return err should be nil")
	assert.Equal(t, 1, len(users), "Return users length should be 0")
}
