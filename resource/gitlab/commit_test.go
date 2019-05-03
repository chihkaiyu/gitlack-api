package gitlab

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSingleCommitWithCommit(t *testing.T) {
	// arrange
	stubByte, expected := getCommit()
	stubClient := getGetClientWithResponse(stubByte, http.StatusOK, "")
	g := getGitLab(stubClient)

	// act
	actual, _ := g.GetSingleCommit(1, "fake-sha")

	// assert
	assert.Equal(t, expected, actual, "Commit' content should be equal")
}

func TestGetSingleCommitRequestError(t *testing.T) {
	// arrange
	stubClient := getGetClientWithError("fake-error")
	g := getGitLab(stubClient)

	// act
	_, err := g.GetSingleCommit(1, "fake-sha")

	// assert
	assert.NotNil(t, err, "Error should not be nil")
	assert.Equal(t, "fake-error", err.Error(), "Error message should be equal")
}

func TestGetSingleCommitInvalidGitLabAPI(t *testing.T) {
	// arrange
	stubClient := getGetClientWithResponse([]byte(`fake-body`), http.StatusNotFound, "")
	g := getGitLab(stubClient)

	// act
	_, err := g.GetSingleCommit(1, "fake-sha")

	// assert
	assert.NotNil(t, err, "Error should not be nil")
	assert.Equal(t, "Invalid GitLab API error: fake-body", err.Error(), "Error message should be equal")
}
