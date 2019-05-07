package gitlab

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTagListGetOnce(t *testing.T) {
	// arrange
	stubByte, _ := getTagListResponse(1)
	stubClient := getGetClientWithResponse(stubByte, http.StatusOK, "")
	g := getGitLab(stubClient)

	// act
	g.GetTagList(1)

	// assert
	stubClient.AssertNumberOfCalls(t, "Get", 1)
}

func TestGetTagListWithFiveTags(t *testing.T) {
	// arrange
	stubByte, expected := getTagListResponse(5)
	stubClient := getGetClientWithResponse(stubByte, http.StatusOK, "")
	g := getGitLab(stubClient)

	// act
	actual, _ := g.GetTagList(1)

	// assert
	assert.Equal(t, 5, len(actual), "Number of tags should be equal")
	assert.ElementsMatch(t, expected, actual, "Tags' content should be equal")
}

func TestGetTagListRequestError(t *testing.T) {
	// arrange
	stubClient := getGetClientWithError("fake-error")
	g := getGitLab(stubClient)

	// act
	_, err := g.GetTagList(1)

	// assert
	assert.NotNil(t, err, "Error should not be nil")
	assert.Equal(t, "fake-error", err.Error(), "Error message should be equal")
}

func TestGetTagListInvalidGitLabAPI(t *testing.T) {
	// arrange
	stubClient := getGetClientWithResponse([]byte(`fake-body`), http.StatusNotFound, "")
	g := getGitLab(stubClient)

	// act
	_, err := g.GetTagList(1)

	// assert
	assert.NotNil(t, err, "Error should not be nil")
	assert.Equal(t, "Invalid GitLab API error: fake-body", err.Error(), "Error message should be equal")
}
