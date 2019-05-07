package gitlab

import (
	"gitlack/model"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProjectGetOnce(t *testing.T) {
	// arrange
	stubByte, _ := getProjectResponse(1)
	stubClient := getGetClientWithResponse(stubByte, http.StatusOK, "")
	g := getGitLab(stubClient)

	// act
	g.GetProject()

	// assert
	stubClient.AssertNumberOfCalls(t, "Get", 1)
}

func TestGetProjectWithFiveProjects(t *testing.T) {
	// arrange
	stubByte, gitlabProjects := getProjectResponse(5)
	var expected []*model.Project
	for _, p := range gitlabProjects {
		expected = append(expected, &model.Project{
			Name: p.Name,
			ID:   p.ID,
		})
	}
	stubClient := getGetClientWithResponse(stubByte, http.StatusOK, "")
	g := getGitLab(stubClient)

	// act
	actual, _ := g.GetProject()

	// assert
	assert.Equal(t, 5, len(actual), "Number of projects should be equal")
	assert.ElementsMatch(t, expected, actual, "Projects' content should be equal")
}

func TestGetProjectRequestError(t *testing.T) {
	// arrange
	stubClient := getGetClientWithError("fake-error")
	g := getGitLab(stubClient)

	// act
	_, err := g.GetProject()

	// assert
	assert.NotNil(t, err, "Error should not be nil")
	assert.Equal(t, "fake-error", err.Error(), "Error message should be equal")
}

func TestGetProjectInvalidGitLabAPI(t *testing.T) {
	// arrange
	stubClient := getGetClientWithResponse([]byte(`fake-error`), http.StatusNotFound, "")
	g := getGitLab(stubClient)

	// act
	_, err := g.GetProject()

	// assert
	assert.NotNil(t, err, "Error should not be nil")
	assert.Equal(t, "Invalid GitLab API error: fake-error", err.Error(), "Error message should be equal")
}
