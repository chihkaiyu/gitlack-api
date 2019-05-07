package gitlab

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserOnePage(t *testing.T) {
	// arragne
	stubByte, _ := getGitLabUserResponse(1)
	stubClient := getGetClientWithResponse(stubByte, http.StatusOK, "")
	g := getGitLab(stubClient)

	// act
	g.GetUser()

	// assert
	stubClient.AssertNumberOfCalls(t, "Get", 1)
}

func TestGetUserRequestError(t *testing.T) {
	// arrange
	stubClient := getGetClientWithError("fake-error")
	g := getGitLab(stubClient)

	// act
	_, err := g.GetUser()

	// assert
	assert.NotNil(t, err, "Error should not be nil")
	assert.Equal(t, "fake-error", err.Error(), "Error message should be equal")
}

func TestGetUserInvalidGitLabAPI(t *testing.T) {
	// arrange
	stubClient := getGetClientWithResponse([]byte(`fake-body`), http.StatusNotFound, "")
	g := getGitLab(stubClient)

	// act
	_, err := g.GetUser()

	// assert
	assert.NotNil(t, err, "Error should not be nil")
	assert.Equal(t, "Invalid GitLab API error: fake-body", err.Error(), "Error message should be equal")
}

func TestGetUserWithFiveUsers(t *testing.T) {
	// arrange
	stubByte, expected := getGitLabUserResponse(5)
	stubClient := getGetClientWithResponse(stubByte, http.StatusOK, "")
	g := getGitLab(stubClient)

	// act
	users, _ := g.GetUser()

	// assert
	assert.Equal(t, 5, len(users), "Number of users should be equal")
	assert.ElementsMatch(t, expected, users, "Users' content should be equal")
}
